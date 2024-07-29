package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"proxy/utils"
	"strings"
	"sync"
)

type Handler struct {
	repository Repository
}

func NewHandler(repository Repository) *Handler {
	return &Handler{repository}
}

func (h Handler) handleRequest(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Proxy-Authorization")
	tokenPart := strings.Split(token, " ")
	if len(tokenPart) != 2 {
		w.Header().Add("Proxy-Authenticate", `Basic`)
		w.WriteHeader(407)
		w.Write([]byte(fmt.Sprintf(`{"id": "%s", "message": "%s"}, "method": "%s", "url": "%s"`, "no_proxy", "No token found", r.Method, r.URL)))
		return
	}
	if tokenPart[0] != "Basic" && tokenPart[1] == "" {
		w.Header().Add("Proxy-Authenticate", `Basic`)
		w.WriteHeader(407)
		w.Write([]byte(fmt.Sprintf(`{"id": "%s", "message": "%s"}, "method": "%s", "url": "%s"`, "no_proxy", "No token found", r.Method, r.URL)))
		return
	}

	project, err := h.repository.GetProjectByToken(tokenPart[1])
	if err != nil {
		w.WriteHeader(407)
		w.Write([]byte(fmt.Sprintf(`{"id": "%s", "message": "%s"}, "method": "%s", "url": "%s"`, "no_project", err, r.Method, r.URL)))
		log.Println(err)
		return
	}

	proxy, err := h.repository.GetProxyAndUpdateConnection(*project)
	if err != nil {
		w.WriteHeader(407)
		w.Write([]byte(fmt.Sprintf(`{"id": "%s", "message": "%s"}, "method": "%s", "url": "%s"`, "no_proxy", err, r.Method, r.URL)))
		return
	}

	// Open the TLS tunnel
	//net.DialTimeout()
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM([]byte(proxy.Config.Certificate.Cert)); !ok {
		log.Fatalf("unable to parse proxy cert")
	}

	cert, err := tls.X509KeyPair([]byte(proxy.Config.Certificate.Cert), []byte(proxy.Config.Certificate.Key))
	if err != nil {
		log.Fatalf("unable to parse proxy cert and key: %s", err.Error())
	}

	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	}

	log.Printf("Connecting to %s:%d\n", proxy.Config.Address.Hostname, proxy.Config.Address.Port)
	proxyConn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", proxy.Config.Address.Hostname, proxy.Config.Address.Port), config)
	//proxyConn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", proxy.Config.Address.Port), config)
	if err != nil {
		log.Println(err)
		return
	}
	defer proxyConn.Close()

	host := r.Host
	if len(strings.Split(host, ":")) == 1 {
		if r.URL.Scheme == "http" {
			host = fmt.Sprintf("%s:80", host)
		} else {
			host = fmt.Sprintf("%s:443", host)
		}
	}

	dataWritted, err := proxyConn.Write([]byte(fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", host, host)))
	fmt.Printf("Data writted: %i\n", dataWritted)
	if err != nil {
		fmt.Println(err.Error())
	}

	scanner := bufio.NewScanner(proxyConn)
	scanner.Scan()
	line := scanner.Bytes()
	if string(line) != "HTTP/1.1 200 OK" {
		return
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hj.Hijack()

	if r.URL.Scheme == "https" || (len(strings.Split(host, ":")) == 2 && strings.Split(host, ":")[1] == "443") {
		clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	} else {
		r.RequestURI = ""
		r.Header.Del("Accept-Encoding")
		r.Header.Del("Proxy-Connection")
		r.Header.Del("Proxy-Authenticate")
		r.Header.Del("Proxy-Authorization")
		if r.Header.Get("Connection") == "close" {
			r.Close = false
		}
		r.Header.Del("Connection")
		proxyReq, _ := http.NewRequest(r.Method, r.URL.String(), r.Body)
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}
		err = proxyReq.Write(proxyConn)
		if err != nil {
			log.Println(err.Error())
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go utils.PipeSocket(clientConn, proxyConn, &wg)
	go utils.PipeSocket(proxyConn, clientConn, &wg)

	wg.Wait()
	log.Println("Done")
	return
}

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"proxy/utils"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	caCertPool *x509.CertPool
}

func (h Handler) ServeRequest(req *http.Request, conn net.Conn) {
	if req != nil && req.Method == "CONNECT" {
		_, certError := conn.(*tls.Conn).ConnectionState().PeerCertificates[0].Verify(struct {
			DNSName                   string
			Intermediates             *x509.CertPool
			Roots                     *x509.CertPool
			CurrentTime               time.Time
			KeyUsages                 []x509.ExtKeyUsage
			MaxConstraintComparisions int
		}{Roots: h.caCertPool})

		if certError != nil {
			fmt.Println("Error verifying certificate")
			fmt.Println(certError)
			conn.Write([]byte("HTTP/1.1 401 connect_error\r\nX-Scrapoxy-Proxyerror: invalid certificate\r\n\r\n\r\n"))
			return
		}
		host := req.Host
		if len(strings.Split(host, ":")) != 2 {
			fmt.Println(req.URL)
		}
		remoteConn, err := net.DialTimeout("tcp", host, 60*time.Second)
		if err != nil {
			fmt.Println(err)
			conn.Write([]byte("HTTP/1.1 500 connect_error\r\nX-Scrapoxy-Proxyerror: ${errMessage}\r\n\r\n\r\n"))
			return
		}
		defer remoteConn.Close()

		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

		var wg sync.WaitGroup
		wg.Add(2)

		go utils.PipeSocket(conn, remoteConn, &wg)
		go utils.PipeSocket(remoteConn, conn, &wg)

		wg.Wait()

		return
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\n\n"))
		return
	}
}

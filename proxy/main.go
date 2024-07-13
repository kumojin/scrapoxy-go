package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("addr", ":3128", "HTTPS network address")
	certFile := flag.String("certfile", "certificate.pem", "certificate PEM file")
	keyFile := flag.String("keyfile", "certificate.key", "key PEM file")
	flag.Parse()

	caCert, err := os.ReadFile(*certFile)
	if err != nil {
		log.Fatal("Error opening cert file", certFile, ", error ", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		ClientAuth:   tls.RequireAnyClientCert,
		ClientCAs:    caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	log.Printf("Starting server on %s", *addr)
	l, err := tls.Listen("tcp", *addr, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("accepted connection from %s\n", conn.RemoteAddr())

		go func(c net.Conn) {
			scanner := bufio.NewScanner(c)
			header := []byte{}
			for scanner.Scan() {
				line := scanner.Bytes()
				header = append(header, fmt.Sprintf("%s\r\n", line)...)
				if string(line) == "\r\n" || string(line) == "" {
					break
				}
			}

			req, _ := http.ReadRequest(bufio.NewReader(bytes.NewReader(header)))

			h := Handler{caCertPool: caCertPool}
			h.ServeRequest(req, c)
			c.Close()
			log.Printf("closing connection from %s\n", conn.RemoteAddr())
		}(conn)
	}

}

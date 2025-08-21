// server.go
package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

var clients = make(map[string]net.Conn)

// Allowed tokens for subdomains
var tokens = map[string]string{
	"demo": "supersecrettoken123",
	"test": "othertoken456",
}

func main() {
	domain := "example.com" // change to your domain

	// Start tunnel server (for client connections)
	go startTunnelServer(":9000")

	// HTTPS server with Let's Encrypt
	m := &autocert.Manager{
		Cache:      autocert.DirCache(".certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain, "*."+domain),
	}

	srv := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate: m.GetCertificate,
		},
		Handler: http.HandlerFunc(handleRequest),
	}

	log.Printf("Public server running with TLS on %s", domain)
	log.Fatal(srv.ListenAndServeTLS("", ""))
}

func startTunnelServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Tunnel server running on", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	handshake := strings.TrimSpace(string(buf[:n]))
	parts := strings.Split(handshake, ":")
	if len(parts) != 2 {
		log.Println("Invalid handshake")
		return
	}
	subdomain, token := parts[0], parts[1]

	expected, ok := tokens[subdomain]
	if !ok || expected != token {
		log.Printf("Auth failed for %s", subdomain)
		return
	}

	clients[subdomain] = conn
	log.Printf("Client connected: %s", subdomain)

	io.Copy(io.Discard, conn)
	delete(clients, subdomain)
	log.Printf("Client disconnected: %s", subdomain)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	sub := strings.Split(host, ".")[0]

	client, ok := clients[sub]
	if !ok {
		http.Error(w, "No client connected", 502)
		return
	}

	if err := r.Write(client); err != nil {
		http.Error(w, "Failed to forward request", 500)
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(client), r)
	if err != nil {
		http.Error(w, "Failed to read response", 500)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

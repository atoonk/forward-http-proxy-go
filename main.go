package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	// Modify the host if necessary.
	//r.Host = mapHost(r.Host)
	if r.Host == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Printf("Failed to connect to host %s: %v", r.Host, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	log.Printf("Tunneling from %s to %s", r.RemoteAddr, r.Host)
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go connect(destConn, clientConn)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	//req.Host = mapHost(req.Host)
	if req.Host == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	req.URL.Host = req.Host
	req.URL.Scheme = "http" // You can also dynamically set the scheme based on the original request

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func connect(destConn, clientConn net.Conn) {
	defer destConn.Close()
	defer clientConn.Close()
	go copy(destConn, clientConn)
	copy(clientConn, destConn)
}

func copy(dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
}

func main() {
	server := &http.Server{
		//Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Received request %s %s %s\n", r.Method, r.Host, r.RemoteAddr)
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
	}
	// create a listener
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	//log.Fatal(server.ListenAndServe())
	log.Fatal(server.Serve(listener))
}

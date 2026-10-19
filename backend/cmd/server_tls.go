package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// The certificate is loaded before the server starts so a broken certificate stops the process
// instead of leaving it running without a Web UI.
func loadServerTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load %s and %s: %w", certFile, keyFile, err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// With a TLS config the one listener speaks only HTTPS; Go answers a plain HTTP request on it
// with 400 and closes the connection.
func listenAndServe(srv *http.Server) error {
	if srv.TLSConfig != nil {
		return srv.ListenAndServeTLS("", "")
	}

	return srv.ListenAndServe()
}

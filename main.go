package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

func main() {
	pair, err := loadTLSKey()
	if err != nil {
		glog.Error(err)
		return
	}
	server := setupListener(pair)
	glog.Fatal(server.ListenAndServeTLS("", ""))
}
func setupListener(pair *tls.Certificate) *http.Server {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", 443),
	}

	if pair != nil {
		server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{*pair}}
	}

	rtr := mux.NewRouter()
	rtr.HandleFunc("/test", CAWebhookHandleRequest)

	server.Handler = rtr
	return server
}

func loadTLSKey() (*tls.Certificate, error) {
	certFile := ""
	keyFile := ""

	flag.StringVar(&certFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&keyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	// flag.BoolVar(&localAgent, "debug", false, "Run with local agent")
	flag.Parse()

	fmt.Printf("!!!!Startung Webhook!!!!\n")

	if keyFile == "" || certFile == "" {
		return nil, fmt.Errorf("keyFile or certFile not provided")
	}

	pair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("Filed to load key pair: %v", err)
	}
	return &pair, nil

}

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	sal "github.com/salrashid123/vault_pki_mtls/signer"
	"golang.org/x/net/http2"
)

var ()

func fronthandler(w http.ResponseWriter, r *http.Request) {
	log.Println("/ called")
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		cn := strings.ToLower(r.TLS.PeerCertificates[0].Subject.CommonName)
		log.Printf("CN: %s\n", cn)
	}
	fmt.Fprint(w, "ok")
}

func main() {

	clientCaCert, err := os.ReadFile("certs/Vault_CA.pem")
	if err != nil {
		fmt.Printf("Could not read vault issued CA: %v", err)
		return
	}
	clientCaCertPool := x509.NewCertPool()
	clientCaCertPool.AppendCertsFromPEM(clientCaCert)

	r, err := sal.NewVaultCrypto(&sal.Vault{
		VaultToken:         "s.5Nv6F6WKSQl0ycfzhaj1JeCY",
		CertCN:             "server.domain.com",
		VaultCAcert:        "certs/CA_crt.pem",
		VaultAddr:          "https://vault.domain.com:8200",
		VaultPath:          "pki/issue/domain-dot-com",
		SignatureAlgorithm: x509.SHA256WithRSAPSS,
		ExtTLSConfig: &tls.Config{
			ClientCAs:  clientCaCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	})
	if err != nil {
		fmt.Printf("Unable to initialize vault crypto: %v", err)
		return
	}

	http.HandleFunc("/", fronthandler)

	var server *http.Server
	server = &http.Server{
		Addr:      ":8081",
		TLSConfig: r.TLSConfig(),
	}
	http2.ConfigureServer(server, &http2.Server{})
	log.Println("Starting Server..")
	err = server.ListenAndServeTLS("", "")
	log.Fatalf("Unable to start Server %v", err)
}

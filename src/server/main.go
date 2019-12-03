package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	sal "github.com/salrashid123/signer/vault"
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

func healthhandler(w http.ResponseWriter, r *http.Request) {
	log.Println("heathcheck...")
	fmt.Fprint(w, "ok")
}

func main() {

	clientCaCert, err := ioutil.ReadFile("Vault_CA.pem")
	clientCaCertPool := x509.NewCertPool()
	clientCaCertPool.AppendCertsFromPEM(clientCaCert)

	r, err := sal.NewVaultCrypto(&sal.Vault{

		CertCN:      "grpc.domain.com",
		VaultToken:  "s.IumzeFZVsWqYcJ2IjlGaqZby",
		VaultPath:   "pki/issue/domain-dot-com",
		VaultCAcert: "CA_crt.pem",
		VaultAddr:   "https://grpc.domain.com:8200",
		ClientCAs:   clientCaCertPool,
		ClientAuth:  tls.RequireAndVerifyClientCert,
	})
	if err != nil {
		fmt.Printf("Unable to initialize vault crypto: %v", err)
		return
	}
	// ******************************

	// hash := sha256.New()
	// msg := []byte("foo")
	// ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, r.Public().(*rsa.PublicKey), msg, nil)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// log.Printf("Encrypted Data: %v", base64.StdEncoding.EncodeToString(ciphertext))
	// plaintext, err := r.Decrypt(rand.Reader, ciphertext, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("Decrypted Data: %v, ", string(plaintext))

	// ******************************

	http.HandleFunc("/", fronthandler)
	http.HandleFunc("/_ah/health", healthhandler)

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

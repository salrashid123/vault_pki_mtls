package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	sal "github.com/salrashid123/vault_pki_mtls/signer"
)

var ()

func main() {

	trustCaCert, err := os.ReadFile("certs/Vault_CA.pem")
	if err != nil {
		fmt.Printf("Unable to initialize vault crypto: %v", err)
		return
	}
	trustCaCertPool := x509.NewCertPool()
	trustCaCertPool.AppendCertsFromPEM(trustCaCert)

	r, err := sal.NewVaultCrypto(&sal.Vault{
		VaultToken:         "s.egwc0xAvSho80sViVhtjGHVM",
		CertCN:             "client.domain.com",
		VaultPath:          "pki/issue/domain-dot-com",
		VaultCAcert:        "certs/CA_crt.pem",
		VaultAddr:          "https://vault.domain.com:8200",
		SignatureAlgorithm: x509.SHA256WithRSAPSS,
		ExtTLSConfig: &tls.Config{
			RootCAs: trustCaCertPool,
		},
	})
	if err != nil {
		fmt.Printf("Unable to initialize vault crypto: %v", err)
		return
	}
	tr := &http.Transport{
		TLSClientConfig: r.TLSConfig(),
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("https://server.domain.com:8081")
	if err != nil {
		log.Println(err)
		return
	}
	if err != nil {
		fmt.Printf("Unable to initialize vault crypto: %v", err)
		return
	}
	htmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("%v\n", resp.Status)
	fmt.Printf(string(htmlData))

}

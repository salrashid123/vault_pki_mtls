package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	sal "github.com/salrashid123/signer/vault"
)

var ()

func main() {

	r, err := sal.NewVaultCrypto(&sal.Vault{

		CertCN:      "client.domain.com",
		VaultToken:  "s.BtpHNHEpxaWkEF1ThQKEupwL",
		VaultPath:   "pki/issue/domain-dot-com",
		VaultCAcert: "CA_crt.pem",
		VaultAddr:   "https://grpc.domain.com:8200",
	})
	if err != nil {
		fmt.Printf("Unable to initialize vault crypto: %v", err)
		return
	}
	tr := &http.Transport{
		TLSClientConfig: r.TLSConfig(),
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("https://grpc.domain.com:8081")
	if err != nil {
		log.Println(err)
		return
	}
	if err != nil {
		fmt.Printf("Unable to initialize vault crypto: %v", err)
		return
	}
	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("%v\n", resp.Status)
	fmt.Printf(string(htmlData))

}

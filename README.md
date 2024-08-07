
# mTLS using Hashcorp Vault's PKI Secrets

>> ** NOTE: this repo is archived, not supported (but it works)**

Sample code demonstrating an implementation of [crypto.Signer](https://golang.org/pkg/crypto/#Signer) for `HashiCorp Vault` where the TLS connection certificates are provided by its [PKI Secrets](https://www.vaultproject.io/docs/secrets/pki/index.html) engine.

That is, you can start a golang HTTPS server and client where the certificates are provided by Vault. The client and server will use a provided `VAULT_TOKEN` to acquire the secret.   Vaults' PKI secrets engine allows generates a new PKI Keypair for each authorized call

Note:  there are a [number of considerations](https://www.vaultproject.io/docs/secrets/pki/index.html#keep-certificate-lifetimes-short-for-crl-39-s-sake) you need to account for before using Vault based certificates as one end of mTLS:

1.  Max lease of a certificate:
  Vault based certificates are desinged to be temprorary.  You can extend the lease of a cert by setting `-max-lease-ttl=` but in the end, ths is just a temp token..

2. Public/Private keys are reissued and not saved.
  Vault does _NOT_ save the public private keys so everytime you initialize Vault, a **NEW** keypair is generated and lives on the TTL cited above.  Needless to say, this has significant operational ramifications on usage at scale for just TLS....But if you use this just to run a webswerver instance, that should be fine.

  From the Vault documentation:
  `This secrets engine aligns with Vault's philosophy of short-lived secrets. As such it is not expected that CRLs will grow large; the only place a private key is ever returned is to the requesting client (this secrets engine does not store generated private keys, except for CA certificates).`


Anyway, this article is similar to one for [Google Cloud KMS mtls](https://github.com/salrashid123/kms_golang_signer)

>> Note: if it wasn't clear already..this repo is _not_ supported by Google

If you're still interested in using Vault for a TLS server for any reason...


### Usage

1. Install Vault, unseal in production mode

  You know the drill...this repo uses vault server running in `https` itself (i.,e not `-dev` mode).  You can bootstrap with any CA but keep the public key for that handy (you'll need it later when authenticating the clients)
  Here is a sample Vault `server.conf`.  You can find a sample CA, public, private server in this repo.

```hcl
    backend "file" {
      path = "filebackend"
    }

    ui = true

    listener "tcp" {
      address = "127.0.0.1:8200"
      tls_cert_file = "crt_vault.pem"
      tls_key_file = "key_vault.pem"
    }
```

```bash
cd example
$ vault server -config=server.conf 
  (add to /etc/hosts)
    127.0.0.1 vault.domain.com server.domain.com
# if needed:

export VAULT_ADDR='https://vault.domain.com:8200'
export VAULT_CACERT=certs/CA_crt.pem

$ vault operator init
$ export VAULT_TOKEN=<tokenfrominit>
$ vault  operator unseal
```

2. Use `RootToken` to enable the pki backend

```bash
vault secrets enable pki

vault write pki/config/urls \
    issuing_certificates="https://vault.domain.com:8200/v1/pki/ca" \
    crl_distribution_points="http://vault.domain.com:8200/v1/pki/crl"
```

The last command creates the CA and CRL urls at `vault.domain.com`.  Since this is just a test and because i'm sure you don't own `domain.com`, add the following to your `/etc/hosts`

```
127.0.0.1  vault.domain.com server.domain.com
```

>> Note `vault.domain.com` is the actual Vault server address...the SNI values for `crt_vault.pem` are bound to that...i'm just lazy and didn't reissue the cert...

3. Create a CA for a given domain
  In the following, we're creating CA within Vault with CN domain restrictions for, you know, `domain.com`

  ```bash
  vault write pki/root/generate/internal  common_name=domain.com  ttl=8760h

  vault write pki/config/urls \
     issuing_certificates="https://vault.domain.com:8200/v1/pki/ca"  crl_distribution_points="https://vault.domain.com:8200/v1/pki/crl"
  ```

  Save the public cert as `Vault_CA.pem` by running the command below:

  Once you initialize the PKI engine, download the CA Vault just generated for you (infact you should see the CA cert cain once you run the previous command)

  ```bash
  curl  -s  --cacert CA_crt.pem   https://vault.domain.com:8200/v1/pki/ca  | openssl x509 -inform DER -outform PEM  -out certs/Vault_CA.pem 
  ```

4. Create a Role for the domain

```bash
  vault write pki/roles/domain-dot-com \
    allowed_domains=domain.com \
    allow_subdomains=true \
    max_ttl=72h
```

5. Create a Policy for the mTLS Server and Client

```bash
   vault policy write pki-policy-server pki_server.hcl
   vault policy write pki-policy-client pki_client.hcl
```

Where the pki_server creates a certificate with `CN=server.domain.com` and the client with `CN=client.domain.com`


You can tune/refine the poicies as needed/necessary (i.,e the pki "read" path isn't ever used)

6. Generate `VAULT_TOKEN` representing the server and client:

```bash
$ vault token create -policy=pki-policy-server
	Key                  Value
	---                  -----
	token                s.IumzeFZVsWqYcJ2IjlGaqZby
	token_accessor       ogu3k56jcqeQFOAS2dyQgT5g
	token_duration       768h
	token_renewable      true
	token_policies       ["default" "pki-policy-server"]
	identity_policies    []
	policies             ["default" "pki-policy-server"]


$ vault token create -policy=pki-policy-client
  Key                  Value
  ---                  -----
  token                s.BtpHNHEpxaWkEF1ThQKEupwL
  token_accessor       dCyZhYvuIsiWjEILcAFlyiDU
  token_duration       768h
  token_renewable      true
  token_policies       ["default" "pki-policy-client"]
  identity_policies    []
  policies             ["default" "pki-policy-client"]
```


7. Start Server

Edit `server/main.go` and add in the `VAULT_TOKEN`

The root path where you run the client and server should also include both the CA for Vault and the CA that vault generated for your domain (`CA_crt.pem`, `Vault_CA.pem`)

```golang
	r, err := sal.NewVaultCrypto(&sal.Vault{
		CertCN:      "server.domain.com",
		VaultToken:  "s.IumzeFZVsWqYcJ2IjlGaqZby",
		VaultPath:   "pki/issue/domain-dot-com",
		VaultCAcert: "CA_crt.pem",
		VaultAddr:   "https://vault.domain.com:8200",
		SignatureAlgorithm: x509.SHA256WithRSAPSS,
		ExtTLSConfig: &tls.Config{
			ClientCAs:  clientCaCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	})
```

8. Start Client

Same thing as above on `client/main.go` but use the client's token

```golang
	r, err := sal.NewVaultCrypto(&sal.Vault{
		CertCN:      "client.domain.com",
		VaultToken:  "s.BtpHNHEpxaWkEF1ThQKEupwL",
		VaultPath:   "pki/issue/domain-dot-com",
		VaultCAcert: "CA_crt.pem",
		VaultAddr:   "https://vault.domain.com:8200",
		SignatureAlgorithm: x509.SHA256WithRSAPSS,
		ExtTLSConfig: &tls.Config{
			RootCAs: trustCaCertPool,
		},    
	})
```

if all goes well, you should see a lowly `200 ok` indicating a successful mTLS connection


Finally, you don't ofcourse have to use mTLS here...you can just run a webserver with one certificate in memory and be done with it.

---

### References

- As background on Vault Auth on GCP itself (unrelated to PKI secrets here): [Vault auth and secrets on GCP](https://medium.com/google-cloud/vault-auth-and-secrets-on-gcp-51bd7bbaceb)


### API Documenation
- [Logical.Writes](https://godoc.org/github.com/hashicorp/vault/api#Logical.Write)
- [Vault Policies](https://www.vaultproject.io/docs/concepts/policies.html)

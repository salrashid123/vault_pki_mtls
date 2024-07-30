path "auth/token/lookup-self" {
  capabilities = ["read"]
}

path "auth/token/renew" {
  capabilities = ["update", "create"]
}

path "pki/issue/domain-dot-com" {
    capabilities = ["create", "update", "delete", "list", "read"]
    allowed_parameters = {
      "common_name" = ["server.domain.com"]
  }
}

path "pki/config/urls" {
    capabilities = ["read"]
}
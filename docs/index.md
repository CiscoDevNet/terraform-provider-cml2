---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cml2 Provider"
subcategory: ""
description: |-
  provider schema description
---

# cml2 Provider

provider schema description

## Example Usage

```terraform
provider "cml2" {
  # for use of variables, see
  # https://developer.hashicorp.com/terraform/language/values/variables

  # address must use https://
  address = var.address

  # credentials, either a JWT or username/password are required
  # an error is raised if neither token or username / password are set
  # token   = var.token
  username = var.username
  password = var.password

  # read the CA certificate from file
  # if not specified, the system root CAs are used
  # cacert = file("ca.pem")

  # should the certificate be verified?
  # (defaults to true)
  # skip_verify = false

  # should the API client cache responses?
  # this will improve performance but isn't guaranteed
  # to work for all scenarios
  # (defaults to false)
  # use_cache = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `address` (String) CML2 controller address, must start with `https://`.

### Optional

- `cacert` (String) A CA CERT, PEM encoded. When provided, the controller cert will be checked against it.  Otherwise, the system trust anchors will be used.
- `password` (String, Sensitive) CML2 password.
- `skip_verify` (Boolean) Disables TLS certificate verification.
- `token` (String, Sensitive) CML2 API token (JWT).
- `use_cache` (Boolean) Enables the client cache, this is considered experimental.
- `username` (String) CML2 username.

# Terraform Provider Cisco CML2 (Terraform Plugin Framework)

_This template repository is built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework). The template repository built on the [Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk) can be found at [terraform-provider-scaffolding](https://github.com/hashicorp/terraform-provider-scaffolding). See [Which SDK Should I Use?](https://www.terraform.io/docs/plugin/which-sdk.html) in the Terraform documentation for additional information._

This repository implements a [Terraform](https://www.terraform.io) provider for Cisco Modeling Labs version 2.4 and later. It's current state is "work-in-progress".  The current implementation provides:

- A resource and a data source (`internal/provider/`),
  - resource `cml2_lab` to create, update and destroy a lab based on a YAML topology file
  - update allows to modify state, e.g. from STOPPED to STARTED, ...
  - data source `cml2_lab_details` to retrieve operational state from a running lab.
- Examples (`examples/`) and generated documentation (`docs/`),
- Miscellaneous meta files.

Note:  The examples and docs as well as the tests are pretty much identical to
  the files found in the original templates...  See the [TODO.md](TODO.md) file!

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.17
- [CML2](https://cisco.com/go/cml) >= 2.3.0

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

Fill this in for each provider

```hcl
terraform {
  required_providers {
    cml2 = {
      version = "~>0.0.1"
      source  = "cisco.com/dev/cml2"
    }
}

variable "address" {
  description = "CML controller address"
  type        = string
  default     = "https://192.168.122.245"
}

variable "token" {
  description = "CML API token"
  type        = string
}

variable "toponame" {
  description = "topology name"
  type        = string
  default     = "absolute bananas"
}

provider "cml2" {
  address = var.address
  token   = var.token
  # token       = null
  skip_verify = true
}

resource "cml2_lab" "bananas" {
  topology = templatefile("topology.yaml", { toponame = var.toponame })
  start    = false
  # wait     = false
  # state    = "STARTED"
}

data "cml2_lab_details" "example" {
  id = cml2_lab.bananas.id
  only_with_ip = true
}

output "bla" {
  value = data.cml2_lab_details.example
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

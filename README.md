[![CodeQL](https://github.com/ciscodevnet/terraform-provider-cml2/actions/workflows/codeql-analysis.yml/badge.svg?branch=main)](https://github.com/ciscodevnet/terraform-provider-cml2/actions/workflows/codeql-analysis.yml) [![Go](https://github.com/ciscodevnet/terraform-provider-cml2/actions/workflows/test.yml/badge.svg)](https://github.com/ciscodevnet/terraform-provider-cml2/actions/workflows/test.yml) [![Coverage Status](https://coveralls.io/repos/github/CiscoDevNet/terraform-provider-cml2/badge.svg?branch=main)](https://coveralls.io/github/CiscoDevNet/terraform-provider-cml2?branch=main)

# Terraform Provider for Cisco CML2

This repository implements a [Terraform](https://www.terraform.io) provider for
Cisco Modeling Labs version 2.6 and later. It's current state is "beta". Changes
can be expected, for example:

- configuration (provider, resources, data-sources)
- provider behavior
- features (additional resources, ...)

> [!NOTE]
> The provider needs CML 2.4 or newer. This is due to some additional API
> capabilities which were introduced with 2.4.0. Older versions are blocked
> within by the `gocmlclient`.

The current implementation provides:

- Resources and a data sources (`internal/provider/`),
  - resource `cml2_lab` to create, update and destroy labs
  - resource `cml2_node` to create, update and destroy nodes in a lab
  - resource `cml2_link` to create, update and destroy links between nodes in a lab
  - resource `cml2_lifecycle` to control the state of a lab (like `STARTED`,
  `STOPPED`), including staged starting and configuration injection
  - resource `cml2_group` to create, update and destroy groups
  - resource `cml2_user` to create, update and destroy users
  - data source `cml2_lab` to retrieve state of an existing lab
  - data source `cml2_node` to retrieve state of an existing node in a lab
  - data source `cml2_images` to retrieve the available node images from the controller
  - data source `cml2_groups` to retrieve user groups from the controller
  - data source `cml2_extconn` to retrieve external connector information from
  the controller
  - data source `cml2_system` to retrieve system state (ready state, version,
  ...) from the controller
  - data source `cml2_users` to retrieve users from the controller
- Examples (`examples/`) and generated documentation (`docs/`),
- Miscellaneous meta files.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22
- [CML2](https://cisco.com/go/cml) >= 2.6.0

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

Please refer to the `examples` directory and look at the built-in documentation
provided via the registry.

> [!NOTE]
>
> It's recommended to use the UI token (from the User menu, top right "Copy
> JWT") instead of the user-name and password as that will provide better
> performance.  Using the token avoids repeated client authentication via the
> API which takes quite a bit of time.

### HCL

For some basic examples look in the `examples` directory

## Developing the Provider

If you wish to work on the provider, you'll first need
[Go](http://www.golang.org) installed on your machine (see
[Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put
the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of acceptance tests, run `make testacc`. For this
to work, the provider needs to be configured via environment variables.  Here's
an example:

```shell
# for testing purposes, suggest to use direnv

TF_VAR_username="admin"
TF_VAR_password="secret"
TF_VAR_address="https://cml-controller.cml.lab"
# alternatively, this is faster:
# TF_VAR_token="your-token-here"

export TF_VAR_username TF_VAR_password TF_VAR_address
```

Those variables are referenced for acceptance testing in `internal/provider/testing`.

```shell
make testacc
```

Acceptance testing with GitHub actions must properly set secrets which are used
in the test workflow:

- the NGROK_URL where where the CML API is reachable
- the USERNAME and PASSWORD or alternatively, the CML API TOKEN

This can use Ngrok or other tunneling tools if the CML API isn't directly
reachable from the internet.

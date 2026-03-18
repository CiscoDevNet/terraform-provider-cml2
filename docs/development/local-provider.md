# Local Provider Build + Install (Beta Testing)

This doc describes how to build a local `terraform-provider-cml2` binary and make
Terraform use it instead of downloading the provider from the public registry.

Provider address (as compiled into the binary): `registry.terraform.io/ciscodevnet/cml2`.

## Prereqs

- Terraform CLI >= 1.0
- Go >= 1.22

## Build A Local Binary

Two common ways:

### Option A: Use the Makefile (recommended)

Build a binary with version metadata:

```bash
make build
```

This produces `./terraform-provider-cml2`.

### Option B: Use Go directly

```bash
go install
```

This installs the binary into your Go bin dir (typically `$(go env GOPATH)/bin`).

## Make Terraform Use Your Local Build

Pick one of the following workflows.

### Workflow 1: Install into Terraform's plugin directory (registry-like)

This mimics a registry install by putting the provider at the exact path where
Terraform looks for "already installed" providers.

Build + install using the repo Makefile:

```bash
make devinstall
```

By default, that copies the binary to:

`~/.terraform.d/plugins/registry.terraform.io/ciscodevnet/cml2/<version>/<os>_<arch>/terraform-provider-cml2`

Notes:

- `<version>` comes from `git describe` in `GNUmakefile`. If your Terraform
  configuration pins a version, use the same string.
- `<os>_<arch>` must match your local machine, e.g. `linux_amd64`, `darwin_arm64`.

Example `required_providers` block to ensure Terraform selects the installed
version:

```hcl
terraform {
  required_providers {
    cml2 = {
      source  = "ciscodevnet/cml2"
      version = "= 0.8.3" # replace with the installed <version>
    }
  }
}
```

Then run:

```bash
terraform init
```

### Workflow 2: Use Terraform CLI `dev_overrides` (fastest local iteration)

This tells Terraform to always use a provider binary from a local directory and
skip version selection and registry downloads.

1) Create a directory for the provider binary:

```bash
mkdir -p "$HOME/.terraform.d/dev-plugins"
```

2) Build the provider and copy it there:

```bash
make build
cp ./terraform-provider-cml2 "$HOME/.terraform.d/dev-plugins/terraform-provider-cml2"
```

3) Create a Terraform CLI config file:

- Linux/macOS: `~/.terraformrc` (or `~/.config/terraform.rc`)
- Windows: `%APPDATA%\terraform.rc`

Example `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/ciscodevnet/cml2" = "~/.terraform.d/dev-plugins"
  }

  direct {}
}
```

4) Run `terraform init` in your test configuration.

Important:

- With `dev_overrides`, Terraform ignores version constraints for that provider.
- Do not commit `~/.terraformrc` into any repo; it's a machine-local setting.

## Verifying Terraform Uses The Local Provider

- `terraform providers` should list `provider[registry.terraform.io/ciscodevnet/cml2]`.
- For detailed install logs: `TF_LOG=debug terraform init` and look for lines
  mentioning the local plugin dir / dev override.

## Troubleshooting

- Wrong OS/arch directory: confirm `go env GOOS` and `go env GOARCH` and use
  `<os>_<arch>` accordingly.
- Terraform still downloads from registry: remove `.terraform/` and re-run
  `terraform init` (and check `~/.terraformrc` if using `dev_overrides`).
- Permission / exec errors: ensure the binary is executable (`chmod +x`).

default: testacc

# temporary, for convenience
ORG := ciscodevnet
NAME := cml2
ARCH := linux_amd64

VERSION := 0.1.0
DEST := ~/.terraform.d/plugins/$(ORG)/$(NAME)/$(VERSION)/$(ARCH)

MIRROR := /tmp/terraform/registry.terraform.io/$(ORG)/$(NAME)

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

build: main.go
	go build -o terraform-provider-cml2 .

devinstall: build
	test -d $(DEST) || mkdir -p $(DEST)
	mv terraform-provider-cml2 $(DEST)

# this needs goreleaser installed and the following env vars defined
# GITHUB_TOKEN
# GPG_FINGERPRINT
mirror:
	goreleaser release --skip-publish --rm-dist
	test -d $(DEST) || mkdir -p $(MIRROR)
	cp dist/*.zip $(MIRROR)

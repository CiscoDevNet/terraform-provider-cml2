default: testacc

# temporary, for convenience
ORG := registry.terraform.io/ciscodevnet
NAME := cml2
ARCH := linux_amd64

VERSION := $(shell git describe --long | sed -re 's/^v(.*)$$/\1/')
DEST := ~/.terraform.d/plugins/$(ORG)/$(NAME)/$(VERSION)/$(ARCH)

MIRROR := /tmp/terraform/$(ORG)/$(NAME)
TESTARGS := -cover -v

# Run tests
.PHONY: tests
tests:
	go test ./... -v $(TESTARGS) -timeout 120m

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

build: main.go
	go build -o terraform-provider-$(NAME) -ldflags "-X main.version=$(VERSION)" .

devinstall: build
	test -d $(DEST) || mkdir -p $(DEST)
	mv terraform-provider-$(NAME) $(DEST)

# this needs goreleaser installed and the following env vars defined
# GITHUB_TOKEN
# GPG_FINGERPRINT
mirror:
	goreleaser release --skip-publish --rm-dist
	test -d $(DEST) || mkdir -p $(MIRROR)
	cp dist/*.zip $(MIRROR)

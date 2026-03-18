SHELL := /usr/bin/env bash
.SHELLFLAGS := -euo pipefail -c

.DEFAULT_GOAL := help

# temporary, for convenience
ORG := registry.terraform.io/ciscodevnet
NAME := cml2

GO ?= go
TERRAFORM ?= terraform

GOOS := $(shell $(GO) env GOOS)
GOARCH := $(shell $(GO) env GOARCH)
ARCH := $(GOOS)_$(GOARCH)

VERSION := $(shell git describe --tags --long --always 2>/dev/null | sed -re 's/^v(.*)$$/\1/')
DEST := ~/.terraform.d/plugins/$(ORG)/$(NAME)/$(VERSION)/$(ARCH)

MIRROR := /tmp/terraform/$(ORG)/$(NAME)

COVEROUT := coverage
COVERPROFILE := $(COVEROUT).out

TEST_TIMEOUT ?= 120m

TEST_PKGS ?= ./...

.PHONY: help
help:
	@printf "%s\n" \
		"Common targets:" \
		"  make test         Run unit tests" \
		"  make acc          Run acceptance tests (TF_ACC=1)" \
		"  make lint         Run golangci-lint" \
		"  make fmt          Run gofmt" \
		"  make generate     Run go generate (docs/examples)" \
		"  make tidy-check   Ensure 'go mod tidy' produces no diff" \
		"  make build        Build provider binary" \
		"  make devinstall   Install locally into ~/.terraform.d/plugins/..." \
		"  make cover        Render HTML coverage report" \
		"  make clean        Remove generated coverage artifacts"

.PHONY: deps
deps:
	$(GO) mod download

# Update dependencies (use with care)
.PHONY: update
update:
	$(GO) get -u ./...
	$(GO) mod download && $(GO) mod verify && $(GO) mod tidy

# Format
.PHONY: fmt
fmt:
	$(GO)fmt -w .

# Lint
.PHONY: lint
lint:
	command -v golangci-lint >/dev/null 2>&1 || (echo "ERROR: golangci-lint not found"; echo "Install: https://golangci-lint.run/welcome/install/"; exit 1)
	golangci-lint run

# Run tests
.PHONY: test tests
test tests:
	$(GO) test $(TEST_PKGS) -v -timeout $(TEST_TIMEOUT) -cover -coverprofile $(COVERPROFILE)

# Run acceptance tests
.PHONY: acc testacc
acc testacc:
	TF_ACC=1 $(GO) test $(TEST_PKGS) -v -timeout $(TEST_TIMEOUT) -cover -coverprofile $(COVERPROFILE)

.PHONY: generate
generate:
	command -v $(TERRAFORM) >/dev/null 2>&1 || (echo "ERROR: terraform not found in PATH"; exit 1)
	$(GO) generate ./...

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: tidy-check
tidy-check:
	$(GO) mod tidy
	git diff --compact-summary --exit-code || (echo; echo "Unexpected difference after 'go mod tidy'. Run 'go mod tidy' and commit."; exit 1)

build: main.go
	$(GO) build -o terraform-provider-$(NAME) -ldflags "-X main.version=$(VERSION)" .

devinstall: build
	test -d $(DEST) || mkdir -p $(DEST)
	install -m 0755 terraform-provider-$(NAME) $(DEST)/terraform-provider-$(NAME)

# this needs goreleaser installed and the following env vars defined
# GITHUB_TOKEN
# GPG_FINGERPRINT
mirror:
	command -v goreleaser >/dev/null 2>&1 || (echo "ERROR: goreleaser not found"; exit 1)
	goreleaser release --skip=publish --clean
	test -d $(MIRROR) || mkdir -p $(MIRROR)
	cp dist/*.zip $(MIRROR)

.PHONY: cover
cover:
	$(GO) tool cover -html $(COVERPROFILE) -o $(COVEROUT).html
	if command -v xdg-open >/dev/null 2>&1; then xdg-open $(COVEROUT).html; \
	elif command -v open >/dev/null 2>&1; then open $(COVEROUT).html; \
	else echo "Wrote $(COVEROUT).html"; fi

.PHONY: clean
clean:
	rm -f $(COVEROUT).html $(COVERPROFILE)

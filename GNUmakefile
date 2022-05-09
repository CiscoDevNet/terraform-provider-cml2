default: testacc

# temporary, for convenience
NAME := cml2
ARCH := linux_amd64
VERSION := 0.0.1
ORG := cisco.com/dev
DEST := ~/.terraform.d/plugins/$(ORG)/$(NAME)/$(VERSION)/$(ARCH)

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

build: main.go
	go build -o terraform-provider-cml2 .

devinstall: build
	test -d $(DEST) || mkdir -p $(DEST)
	mv terraform-provider-cml2 $(DEST)

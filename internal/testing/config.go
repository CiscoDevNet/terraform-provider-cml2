// Package testing provides base configuration for (integration) testing
package testing

// Cfg is a base provider config snippet for tests.
var Cfg = `
variable "address" {
	description = "CML controller address"
	type        = string
}
variable "username" {
	description = "CML controller username"
	type        = string
	default     = ""
}
variable "password" {
	description = "CML controller password"
	type        = string
	default     = ""
}
variable "token" {
	description = "CML controller JWT"
	type        = string
	default     = ""
}
variable "request_headers" {
	description = "Static request headers for the provider"
	type        = map(string)
	default     = {}
}
provider "cml2" {
	address = var.address
	username = var.username
	password = var.password
	token = var.token
	request_headers = var.request_headers
	token_cache = true
	token_cache_file = "/tmp/terraform-provider-cml2-acc-token.json"
	skip_verify = true
	named_configs = false
}
`

// CfgNamedConfigs is a base provider config snippet with named configs enabled.
var CfgNamedConfigs = `
variable "address" {
	description = "CML controller address"
	type        = string
}
variable "username" {
	description = "CML controller username"
	type        = string
	default     = ""
}
variable "password" {
	description = "CML controller password"
	type        = string
	default     = ""
}
variable "token" {
	description = "CML controller JWT"
	type        = string
	default     = ""
}
variable "request_headers" {
	description = "Static request headers for the provider"
	type        = map(string)
	default     = {}
}
provider "cml2" {
	address = var.address
	username = var.username
	password = var.password
	token = var.token
	request_headers = var.request_headers
	token_cache = true
	token_cache_file = "/tmp/terraform-provider-cml2-acc-token.json"
	skip_verify = true
	named_configs = true
}
`

// CfgBroken is an intentionally broken provider config snippet for tests.
var CfgBroken = `
variable "address" {
	description = "CML controller address"
	type        = string
}
variable "username" {
	description = "CML controller username"
	type        = string
	default     = ""
}
variable "password" {
	description = "CML controller password"
	type        = string
	default     = ""
}
variable "token" {
	description = "CML controller JWT"
	type        = string
	default     = ""
}
variable "request_headers" {
	description = "Static request headers for the provider"
	type        = map(string)
	default     = {}
}
provider "cml2" {
	# something non-existent
	address = "https://127.0.0.1:5555"
	username = var.username
	password = var.password
	token = var.token
	request_headers = var.request_headers
	skip_verify = true
}
`

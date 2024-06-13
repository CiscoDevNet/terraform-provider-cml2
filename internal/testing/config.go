package testing

var Cfg string = `
variable "address" {
	description = "CML controller address"
	type        = string
}
variable "username" {
	description = "CML controller username"
	type        = string
}
variable "password" {
	description = "CML controller password"
	type        = string
}
variable "token" {
	description = "CML controller JWT"
	type        = string
}
provider "cml2" {
	address = var.address
	username = var.username
	password = var.password
	token = var.token
	skip_verify = true
	named_configs = false
}
`

var CfgNamedConfigs string = `
variable "address" {
	description = "CML controller address"
	type        = string
}
variable "username" {
	description = "CML controller username"
	type        = string
}
variable "password" {
	description = "CML controller password"
	type        = string
}
variable "token" {
	description = "CML controller JWT"
	type        = string
}
provider "cml2" {
	address = var.address
	username = var.username
	password = var.password
	token = var.token
	skip_verify = true
	named_configs = true
}
`

var CfgBroken string = `
variable "address" {
	description = "CML controller address"
	type        = string
}
variable "username" {
	description = "CML controller username"
	type        = string
}
variable "password" {
	description = "CML controller password"
	type        = string
}
provider "cml2" {
	# something non-existent
	address = "https://127.0.0.1:5555"
	username = var.username
	password = var.password
	skip_verify = true
}
`

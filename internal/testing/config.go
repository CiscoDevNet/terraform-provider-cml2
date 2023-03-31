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
provider "cml2" {
	address = var.address
	username = var.username
	password = var.password
	skip_verify = true
	use_cache = false
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
	use_cache = false
}
`

provider "cml2" {
  # for use of variables, see
  # https://developer.hashicorp.com/terraform/language/values/variables

  # address must use https://
  address = var.address

  # credentials, either a JWT or username/password are required
  # an error is raised if neither token or username / password are set
  # token   = var.token
  username = var.username
  password = var.password

  # read the CA certificate from file
  # if not specified, the system root CAs are used
  # cacert = file("ca.pem")

  # should the server certificate be verified?
  # (defaults to false, it will be verified)
  # skip_verify = true

  # this is deprecated!
  # use_cache = false
}

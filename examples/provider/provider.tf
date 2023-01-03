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

  # should the certificate be verified?
  # (defaults to true)
  # skip_verify = false

  # should the API client cache responses?
  # this will improve performance but isn't guaranteed
  # to work for all scenarios
  # (defaults to false)
  # use_cache = true
}

provider "cml2" {
  # address must use https://
  address = var.address

  # credentials, set either a JWT or username/password
  # token   = var.token
  username = var.username
  password = var.password

  # read the CA certificate from file
  # if not specified, he system root CAs are used
  cacert = file("ca.pem")

  # should the certificate be verified?
  # (defaults to true)
  # skip_verify = false

  # should the API client cache responses?
  # this will improve performance but isn't guaranteed
  # to work for all scenarios
  # (defaults to false)
  # use_cache = true
}

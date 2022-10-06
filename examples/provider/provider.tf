provider "cml2" {
  # address must use https://
  address = var.address
  # set either a JWT or username/password
  # token   = var.token
  username = var.username
  password = var.password
  # read the CA certificate from file
  # if not specified, he system root CAs are used
  cacert = file("ca.pem")
  # should the certificate be verified?
  # (defaults to true)
  # skip_verify = false
}

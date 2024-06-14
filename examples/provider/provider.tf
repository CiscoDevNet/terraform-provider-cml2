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

  # this configuration option is deprecated with 0.8.0
  # use_cache = false

  # dynamic_config allows to initiate a provider with an incomplete config. For
  # example, the address/URL might only be known later, coming from another
  # module. This is being used in cloud-cml where Terraform provisions a CML VM
  # in the cloud and uses the result of that provisioning process to configure
  # the CML Terraform provider to talk to that instance.
  dynamic_config = false

  # named configs was introduced w/ 0.8.0 and CML 2.7.0, the default is false
  # enable this to provide multiple day0 configurations, the Cat 9000v is a
  # device that supports this to provide a unique serial number per device.
  named_configs = false
}

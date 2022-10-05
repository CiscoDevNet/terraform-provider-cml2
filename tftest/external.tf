# data "external" "env" {
#   program = ["${path.module}/env.sh"]

#   For Windows (or Powershell core on MacOS and Linux),
#   run a Powershell script instead
#   program = ["${path.module}/env.ps1"]
# }

# Show the results of running the data source. This is a map of environment
# variable names to their values.
# output "env" {
#   value = data.external.env.result
# }

# This outputs "bar"
# output "foo" {
#   value = data.external.env.result["foo"]
# }

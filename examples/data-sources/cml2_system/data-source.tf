# ensure the system is ready to talk to by using this data source
# with a timeout like 5m (5 minutes)

data "cml2_system" "wait_for_ready" {
  timeout = "5m"
}
output "readiness" {
  value = data.cml2_system.wait_for_ready.version
}

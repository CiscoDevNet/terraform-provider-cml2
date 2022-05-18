resource "cml2_lab" "bananas" {
  topology = file("topology.yaml")
  # if wait is set then wait until the lab converged, it defaults to true
  # wait     = false
  # state can be STARTED or DEFINED_ON_CORE when creating
  # for running lab, it can be also set to STOPPED
  # if not set, it defaults to STARTED (e.g. the lab starts after creating)
  # state    = "STARTED"
}
resource "cml2_lab" "bananas" {
  topology = file("topology.yaml")
  # start    = true
  # wait     = false
  # state    = "STARTED"
}
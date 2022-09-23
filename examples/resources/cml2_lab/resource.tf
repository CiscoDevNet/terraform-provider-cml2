resource "cml2_lab" "bananas" {

  # simply load the content of the given file
  topology = file("topology.yaml")

  # alternatively, use a template and replace variables within
  # topology = templatefile("topology.yaml", { toponame = var.toponame })

  # snippet from topology.yaml:
  # lab:
  #   description: 'lengthy description'
  #   notes: 'some verbose notes'
  #   timestamp: 1606137179.2951126
  #   title: ${toponame}
  #   version: 0.0.4
  # nodes:
  # ...

  # if wait is set then wait until the lab converged, it defaults to true
  # wait     = false

  # state can be STARTED or DEFINED_ON_CORE when creating
  # for running lab, it can be also set to STOPPED
  # if not set, it defaults to DEFINED_ON_CORE (e.g. the lab is created but will
  # not be started after creating).
  # state    = "STARTED"
}

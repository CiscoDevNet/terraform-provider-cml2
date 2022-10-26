resource "cml2_lab" "example" {

  # simply load the content of the given file
  topology = file("example-topology.yaml")

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

  # if wait is set then wait until the lab converged, it defaults to true stages
  # have no effect if wait is false and this will produce a warning!
  # wait = false

  # state can be STARTED or DEFINED_ON_CORE when creating. For a running lab, it
  # can be also set to STOPPED. If not set, it defaults to DEFINED_ON_CORE (e.g.
  # the lab is created but will not be started after creating).
  # state = "STARTED"

  # dictionary, keyed with the node label and the text configuration which will
  # be injected into the node when creating the lab resource.
  configs = {
    "server-0" : "hostname server-0",
    "server-1" : "hostname server-1",
    "server-2" : "hostname server-2"
    "server-3" : "hostname server-3",
  }

  staging = {
    stages = [
      "infrastructure",
      "underlay",
      "overlay",
      "red-team",
      "blue-team"
    ],
    start_remaining = false
  }

  timeouts = {
    create = "20h"
    update = "1h30m"
    delete = "20m" # currently unused
  }

}

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

  # if wait is set then wait until the lab converged, it defaults to true.
  # staging has no effect if wait is set to false and this will produce a
  # warning!
  # wait = false

  # state can be STARTED or DEFINED_ON_CORE when creating. For a running lab, it
  # can be also set to STOPPED. If not set, it defaults to STARTED (e.g.  the
  # lab is created and will be started after creating).
  # state = "STARTED"

  # dictionary, keyed with the node label and the text configuration which will
  # be injected into the node when creating the lab resource.
  configs = {
    "server-0" : "hostname server-0",
    "server-1" : "hostname server-1",
    "server-2" : "hostname server-2"
    "server-3" : "hostname server-3",
  }

  # start the nodes in the order given by the list of node tags.
  # if there's any nodes remaining then leave them alone (don't start them).
  staging = {
    stages = [
      "infrastructure",
      "underlay",
      "overlay",
      "hq",
      "site1",
      "site2",
      "site3"
    ],
    start_remaining = false
  }

  timeouts = {
    create = "20h"
    update = "1h30m"
    delete = "20m" # currently unused
  }

}

# the below will output the first IPv4 address on the first interface of the 
# node with the label "server-0" when the lab state is STARTED.
output "server0ip" {
  value = (cml2_lifecycle.this.state == "STARTED") ? [for k, v in cml2_lifecycle.this.nodes : v.interfaces[0].ip4[0] if v.label == "server-0"][0] : ""
}

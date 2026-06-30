# update_triggers example
#
# Why this exists:
# A lifecycle resource only gets an Update() call when Terraform detects a diff
# on the lifecycle resource itself. If a dependent node is replaced in the same
# apply, its Terraform id changes even when its configuration does not.
#
# This example wires the lifecycle resource to both the node id and the
# generation hash so it notices both:
#   - node replacement (id changes)
#   - node config changes (generation changes)
#
# The composite trigger keeps the lifecycle resource in sync when the unmanaged
# switch is recreated in the same apply cycle.

resource "cml2_lab" "this" {}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  nodedefinition = "external_connector"
  configuration  = "virbr0"
}

resource "cml2_node" "ums" {
  lab_id         = cml2_lab.this.id
  label          = "Unmanaged Switch"
  nodedefinition = "unmanaged_switch"
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "nginx-1"
  nodedefinition = "nginx"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "nginx-2"
  nodedefinition = "nginx"
}

resource "cml2_link" "l0" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ext.id
  node_b = cml2_node.ums.id
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r1.id
}

resource "cml2_link" "l2" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r2.id
}

locals {
  lifecycle_nodes = {
    ext = cml2_node.ext
    ums = cml2_node.ums
    r1  = cml2_node.r1
    r2  = cml2_node.r2
  }
}

resource "cml2_lifecycle" "top" {
  lab_id = cml2_lab.this.id
  state  = "STARTED"

  # The key trick: include the node id as well as its generation. Generation
  # alone covers config changes; id covers recreate/replacement events.
  update_triggers = {
    for name, node in local.lifecycle_nodes : name => "${node.id}:${node.generation}"
  }

  # Links still need to be listed here so lifecycle waits for the full topology.
  depends_on = [
    cml2_node.ext,
    cml2_node.ums,
    cml2_node.r1,
    cml2_node.r2,
    cml2_link.l0,
    cml2_link.l1,
    cml2_link.l2,
  ]
}

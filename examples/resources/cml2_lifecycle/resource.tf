resource "cml2_lab" "this" {
  title       = "fancy lab title"
  description = "extra description"
  notes       = <<-EOT
  # Heading
  - topic one
  - topic two
  EOT
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  configuration  = "bridge0"
  nodedefinition = "external_connector"
  x              = -100
  y              = 0
  tags           = ["infra"]
}

resource "cml2_node" "ums1" {
  lab_id         = cml2_lab.this.id
  label          = "UMS1"
  nodedefinition = "unmanaged_switch"
  ram            = 512
  x              = 0
  y              = 0
  tags           = ["infra"]
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "alpine"
  ram            = 512
  x              = 100
  y              = 0
  tags           = ["group1"]
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "alpine"
  x              = 100
  y              = 130
  tags           = ["group2"]
}

resource "cml2_node" "r3" {
  lab_id         = cml2_lab.this.id
  label          = "R3"
  nodedefinition = "alpine"
  x              = -100
  y              = 130
}

resource "cml2_link" "l0" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ext.id
  node_b = cml2_node.ums1.id
  slot_b = 31
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.r1.id
  node_b = cml2_node.ums1.id
  # slot_a = 3
  # slot_b = 31
}

resource "cml2_link" "l2" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.r2.id
  node_b = cml2_node.ums1.id
}

resource "cml2_link" "l3" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.r3.id
  node_b = cml2_node.ums1.id
}

resource "cml2_link" "l4" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.r2.id
  node_b = cml2_node.r3.id
}

locals {
  lifecycle_nodes = {
    ext  = cml2_node.ext
    ums1 = cml2_node.ums1
    r1   = cml2_node.r1
    r2   = cml2_node.r2
    r3   = cml2_node.r3
  }
}

resource "cml2_lifecycle" "top" {
  lab_id = cml2_lab.this.id

  # Why update_triggers exists:
  # Terraform only calls cml2_lifecycle.Update() when the lifecycle resource
  # itself has a diff. If a dependent node is replaced in the same apply, the
  # new node may come up as DEFINED_ON_CORE and would otherwise miss lifecycle
  # reconciliation in that apply.
  #
  # Practical example:
  # external_connector nodes are commonly replaced when configuration changes
  # (for example "virbr0" -> "bridge0"). The composite trigger below combines
  # node id + generation so lifecycle notices both replacement and config-only
  # changes.
  update_triggers = {
    for name, node in local.lifecycle_nodes : name => "${node.id}:${node.generation}"
  }

  # Links are siblings of lifecycle in the graph, so list them explicitly.
  # This keeps the lifecycle update after the full topology is present.
  depends_on = [
    cml2_node.ext,
    cml2_node.ums1,
    cml2_node.r1,
    cml2_node.r2,
    cml2_node.r3,
    cml2_link.l0,
    cml2_link.l1,
    cml2_link.l2,
    cml2_link.l3,
    cml2_link.l4,
  ]
  staging = {
    stages          = ["infra", "group1"]
    start_remaining = false
  }
  # state = "STARTED"
}

output "r1_ip_address" {
  value = (
    cml2_lifecycle.top.nodes[cml2_node.r1.id].interfaces[0].ip4 == null ?
    "undefined" : (
      length(cml2_lifecycle.top.nodes[cml2_node.r1.id].interfaces[0].ip4) > 0 ?
      cml2_lifecycle.top.nodes[cml2_node.r1.id].interfaces[0].ip4[0] :
      "no ip"
    )
  )
}

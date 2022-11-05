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
  lab_id      = cml2_lab.this.id
  node_a      = cml2_node.ext.id
  node_b      = cml2_node.ums1.id
  node_b_slot = 31
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.r1.id
  node_b = cml2_node.ums1.id
  # node_a_slot = 3
  # node_b_slot = 31
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

resource "cml2_lifecycle" "top" {
  lab_id = cml2_lab.this.id
  # the elements list has the dependencies
  elements = [
    cml2_node.ext.id,
    cml2_node.ums1.id,
    cml2_node.r1.id,
    cml2_node.r2.id,
    cml2_node.r3.id,
    cml2_link.l0.id,
    cml2_link.l1.id,
    cml2_link.l2.id,
    cml2_link.l3.id,
    cml2_link.l4.id,
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

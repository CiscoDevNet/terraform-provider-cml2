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

resource "cml2_lifecycle" "top" {
  lab_id = cml2_lab.this.id
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

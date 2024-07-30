resource "cml2_lab" "this" {
  title       = "eleven nodes"
  description = "extra description"
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  configuration  = "bridge0"
  nodedefinition = "external_connector"
  x              = -250
  y              = 130
  tags           = ["infra"]
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "alpine"
  configuration  = "hostname alpine0"
  ram            = 512
  x              = -109
  y              = 130
  tags           = ["group1"]
}

resource "cml2_node" "ums1" {
  lab_id         = cml2_lab.this.id
  label          = "UMS1"
  nodedefinition = "unmanaged_switch"
  ram            = 512
  x              = -28
  y              = -7
  tags           = ["infra"]
}

resource "cml2_node" "r3" {
  lab_id         = cml2_lab.this.id
  label          = "R3"
  nodedefinition = "alpine"
  ram            = 512
  x              = 50
  y              = 130
  tags           = ["group1"]
}

resource "cml2_node" "r4" {
  lab_id         = cml2_lab.this.id
  label          = "R4"
  nodedefinition = "alpine"
  ram            = 512
  x              = 100
  y              = 130
  tags           = ["group1"]
}

resource "cml2_node" "r5" {
  lab_id         = cml2_lab.this.id
  label          = "R5"
  nodedefinition = "alpine"
  ram            = 512
  x              = 150
  y              = 130
  tags           = ["group2"]
}

resource "cml2_node" "r6" {
  lab_id         = cml2_lab.this.id
  label          = "R6"
  nodedefinition = "alpine"
  ram            = 512
  x              = 200
  y              = 130
  tags           = ["group2"]
}

resource "cml2_node" "r7" {
  lab_id         = cml2_lab.this.id
  label          = "R7"
  nodedefinition = "alpine"
  ram            = 512
  x              = 250
  y              = 130
  tags           = ["group3"]
}

resource "cml2_node" "r8" {
  lab_id         = cml2_lab.this.id
  label          = "R8"
  nodedefinition = "alpine"
  ram            = 512
  x              = 300
  y              = 130
  tags           = ["group3"]
}

resource "cml2_node" "r9" {
  lab_id         = cml2_lab.this.id
  label          = "R9"
  nodedefinition = "alpine"
  ram            = 512
  x              = 350
  y              = 130
}

resource "cml2_node" "ios" {
  lab_id         = cml2_lab.this.id
  label          = "R10"
  nodedefinition = "iosv"
  ram            = 512
  x              = -200
  y              = 130
  tags           = ["group4"]
}

resource "cml2_link" "l0" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ext.id
  node_b = cml2_node.ums1.id
  # slot_a = 0
  # slot_b = 31
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.r1.id
  node_b = cml2_node.ums1.id
  slot_a = 0
  slot_b = 31
}


resource "cml2_link" "l2" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.r1.id
  node_b = cml2_node.r3.id
  slot_a = 3
  # slot_b = 0
}


resource "cml2_link" "l3" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums1.id
  node_b = cml2_node.r3.id
  # slot_a = 0
  # slot_b = 0
}

resource "cml2_link" "l4" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums1.id
  node_b = cml2_node.r4.id
  # slot_a = 0
  # slot_b = 0
}
resource "cml2_link" "l5" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums1.id
  node_b = cml2_node.r5.id
  # slot_a = 0
  # slot_b = 0
}
resource "cml2_link" "l6" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums1.id
  node_b = cml2_node.r6.id
  # slot_a = 0
  # slot_b = 0
}
resource "cml2_link" "l7" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums1.id
  node_b = cml2_node.r7.id
  # slot_a = 0
  # slot_b = 0
}
resource "cml2_link" "l8" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums1.id
  node_b = cml2_node.r8.id
  # slot_a = 0
  # slot_b = 0
}
resource "cml2_link" "l9" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums1.id
  node_b = cml2_node.r9.id
  # slot_a = 0
  # slot_b = 0
}
resource "cml2_link" "l10" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums1.id
  node_b = cml2_node.ios.id
  # slot_a = 0
  # slot_b = 0
}

resource "cml2_lifecycle" "top" {
  lab_id = cml2_lab.this.id
  depends_on = [
    # Note that referencing all elements is not strictly
    # required. As an alternative, a single "sentinel"
    # node could be used which is started at the very end
    # based on its tag. In particular, links could be ommited.
    cml2_node.ext,
    cml2_node.ums1,
    cml2_node.ios,
    cml2_node.r1,
    cml2_node.r3,
    cml2_node.r4,
    cml2_node.r5,
    cml2_node.r6,
    cml2_node.r7,
    cml2_node.r8,
    cml2_node.r9,

    cml2_link.l0,
    cml2_link.l1,
    cml2_link.l2,
    cml2_link.l3,
    cml2_link.l4,
    cml2_link.l5,
    cml2_link.l6,
    cml2_link.l7,
    cml2_link.l8,
    cml2_link.l9,
    cml2_link.l10
  ]
  configs = {
    "R1" : "hostname injected-hostname"
  }
  staging = {
    stages          = ["infra", "group1", "group2"]
    start_remaining = false
  }
  # state = "DEFINED_ON_CORE"
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

# bash / fish examples
# a=$(terraform output -raw r1_ip_address)
# set a (terraform output -raw r1_ip_address)

# This shows the use of a lab with nodes and links the links are explicitly
# specified with slots Nothing will be started, just basic create, read update /
# delete.

resource "cml2_lab" "twonode" {
  title       = "two node lab"
  description = "nodes are connected with two links"
  notes       = <<-EOT
  # Heading
  - topic one
  - topic two
  EOT
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.twonode.id
  label          = "R1"
  nodedefinition = "alpine"
  ram            = 512
  x              = 200
  y              = 130
  tags           = ["group1"]
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.twonode.id
  label          = "R2"
  nodedefinition = "alpine"
  x              = 100
  y              = 130
}

resource "cml2_link" "l0" {
  lab_id      = cml2_lab.twonode.id
  node_a      = cml2_node.r1.id
  node_a_slot = 3
  node_b      = cml2_node.r2.id
  node_b_slot = 3
}

resource "cml2_link" "l1" {
  lab_id      = cml2_lab.twonode.id
  node_a      = cml2_node.r1.id
  node_a_slot = 2
  node_b      = cml2_node.r2.id
  node_b_slot = 2
}

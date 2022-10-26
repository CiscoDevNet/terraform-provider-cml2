# get a node by ID, lab_id is required!
data "cml2_node" "r1_by_id" {
  id     = "a604ca39-4a81-4743-b888-ca7b8a571b8e"
  lab_id = "a6c124ca-1268-4de1-8bb0-6bb01e7764af"
}

# the node data is in the node attribute
output "result" {
  value = data.cml2_node.r1_by_id.node
}

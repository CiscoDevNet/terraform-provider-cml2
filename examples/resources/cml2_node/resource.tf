resource "cml2_lab" "this" {}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "alpine"
}

# generation is a provider-computed hash over replacement-relevant inputs.
# Use it from cml2_lifecycle.update_triggers to force lifecycle reconciliation
# when this node is replaced.
output "r1_generation" {
  value = cml2_node.r1.generation
}


resource "cml2_lab" "this" {
  topology = templatefile("topology.yaml", { toponame = var.toponame })
  wait  = true
  state = "DEFINED_ON_CORE"
  # state = "STARTED"
  # state = "STOPPED"
  # configurations = var.configs
  # special        = var.special
}


# data "cml2_lab_details" "example" {
#   id           = cml2_lab.this.id
#   only_with_ip = true
# }

# output "bla" {
#   value = data.cml2_lab_details.example
# }

output "bla" {
  sensitive = false
  # value = [cml2_lab.this.state, cml2_lab.this.booted, cml2_lab.this.nodes]
  value = [ for n in cml2_lab.this.nodes : "${n.id} = ${n.label}" ]
}

# resource {
#   cml2_lab.bananas.state = "STARTED"
# }

# module "bla" {
#   source = "../../"
#   name = "baem"
# }

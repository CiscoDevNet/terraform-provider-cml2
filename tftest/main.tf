
resource "cml2_lab" "this" {

  topology = templatefile("topology.yaml", { toponame = var.toponame })

  # stages have no effect if wait is false
  # wait = false

  # labs are always started with create
  # labs are always stopped / wiped / removed with delete
  # state = "STOPPED"

  # lifecycle {
  #   # ignore_changes = [compute_resources[0].desired_vcpus]
  #   # ignore_changes = [nodes]
  #   ignore_changes = [state]
  # }

  # state = "DEFINED_ON_CORE"
  # state = "STOPPED"
  # state = "STARTED"
  # special        = var.special
  configs = var.configs

  staging = {
    # stages = var.stages
    stages = ["infra", "group1", "group2"]
    # start_remaining = false
  }

  # timeouts = {
  #   create = "20h"
  #   update = "1h30m"
  #   delete = "20m"
  # }
}


# data "cml2_lab_details" "example" {
#   id           = cml2_lab.this.id
#   only_with_ip = true
# }

# output "bla" {
#   value = data.cml2_lab_details.example
# }

output "bla" {
  # sensitive = false
  # value = [cml2_lab.this.state, cml2_lab.this.booted, cml2_lab.this.nodes]
  value = [for n in cml2_lab.this.nodes : "${n.id} = ${n.label}"]
}

output "nodes" {
  value = cml2_lab.this.nodes
}

# resource {
#   cml2_lab.bananas.state = "STARTED"
# }

# module "bla" {
#   source = "../../"
#   name = "baem"
# }

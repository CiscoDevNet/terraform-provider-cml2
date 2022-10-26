# get a lab by ID
data "cml2_lab" "lab_by_id" {
  lab_id = "a6c124ca-1268-4de1-8bb0-6bb01e7764af"
}

# get a lab by title
data "cml2_lab" "r1_by_title_name" {
  title = "fancy lab name"
}

# the actual data is in the lab attribute
output "result" {
  value = data.cml2_lab.r1_by_id.lab
}

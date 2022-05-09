data "cml2_lab_details" "example" {
  id           = cml2_lab.bananas.id
  only_with_ip = true
}
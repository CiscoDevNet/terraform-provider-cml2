data "cml2_images" "test" {
  # filter images for the Alpine node definition
  node_definition = "alpine"
}

locals {
  il = data.cml2_images.test.image_list
}

# this returns the image ID of the oldest image for the Alpine node definition
# (last in list...)
output "oldest_alpine" {
  value = element(local.il, length(local.il) - 1).id
}

# the first element of the list has the newest image (alphabetically sorted by
# the image ID)
output "newest_alpine" {
  value = element(data.cml2_images.test.image_list, 0).id
}

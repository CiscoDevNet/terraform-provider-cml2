---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cml2_images Data Source - terraform-provider-cml2"
subcategory: ""
description: |-
  A data source that retrieves image definitions from the controller. The optional node_definition ID can be provided to filter the list of image definitions for the specified node definition. If no node definition ID is provided, the complete image definition list known to the controller is returned.
---

# cml2_images (Data Source)

A data source that retrieves image definitions from the controller. The optional `node_definition` ID can be provided to filter the list of image definitions for the specified node definition. If no node definition ID is provided, the complete image definition list known to the controller is returned.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `node_definition` (String) A node definition ID to filter the image list.

### Read-Only

- `id` (String) A UUID. The attribute required by the framework.
- `image_list` (Attributes List) A list of all image definitions available on the controller, potentially filtered by the provided `node_definition` attribute. (see [below for nested schema](#nestedatt--image_list))

<a id="nestedatt--image_list"></a>
### Nested Schema for `image_list`

Optional:

- `node_definition_id` (String) ID of the node definition this image belongs to

Read-Only:

- `boot_disk_size` (Number) Image specific boot disk size, can be null
- `cpu_limit` (Number) Image specific CPU limit, can be null
- `cpus` (Number) Image specific amount of CPUs, can be null
- `data_volume` (Number) Image specific data volume size, can be null
- `description` (String) Description of this image definition
- `id` (String) ID to identifying the image
- `label` (String) Text label of this image definition
- `ram` (Number) Image specific RAM value, can be null
- `read_only` (Boolean) Is this image definition read only?
- `schema_version` (String) Version of the image definition schemage


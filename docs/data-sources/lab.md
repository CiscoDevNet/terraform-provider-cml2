---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cml2_lab Data Source - terraform-provider-cml2"
subcategory: ""
description: |-
  A lab data source. Either the lab id or the lab title must be provided to retrieve the lab data from the controller.
---

# cml2_lab (Data Source)

A lab data source. Either the lab `id` or the lab `title` must be provided to retrieve the `lab` data from the controller.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) Lab ID that identifies the lab
- `title` (String) Lab title. If not unique, it will return the first one that matches. Use ID for labs with non-unique titles.

### Read-Only

- `lab` (Attributes) lab data (see [below for nested schema](#nestedatt--lab))

<a id="nestedatt--lab"></a>
### Nested Schema for `lab`

Read-Only:

- `created` (String) Creation date/time string in ISO8601 format.
- `description` (String) Lab description.
- `groups` (Attributes Set) Groups assigned to the lab. (see [below for nested schema](#nestedatt--lab--groups))
- `id` (String) Lab identifier, a UUID.
- `link_count` (Number) Number of links in the lab.
- `modified` (String) Modification date/time string in ISO8601 format.
- `node_count` (Number) Number of nodes in the lab.
- `notes` (String) Lab notes.
- `owner` (String) Owner of the lab, a UUID4.
- `state` (String) Lab state, one of `DEFINED_ON_CORE`, `STARTED` or `STOPPED`.
- `title` (String) Title of the lab.

<a id="nestedatt--lab--groups"></a>
### Nested Schema for `lab.groups`

Read-Only:

- `id` (String) Group ID (UUID).
- `name` (String) Descriptive group name.
- `permission` (String) Permission, either `read_only` or `read_write`.

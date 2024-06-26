---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cml2_link Resource - terraform-provider-cml2"
subcategory: ""
description: |-
  A link resource represents a CML link. At create time, the lab ID, source and destination node ID are required.  Interface slots are optional.  By default, the next free interface slot is used.
---

# cml2_link (Resource)

A link resource represents a CML link. At create time, the lab ID, source and destination node ID are required.  Interface slots are optional.  By default, the next free interface slot is used.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `lab_id` (String) Lab ID containing the link (UUID).
- `node_a` (String) Node (A) attached to link.
- `node_b` (String) Node (B) attached to link.

### Optional

- `slot_a` (Number) Optional interface slot on node A (src), if not provided use next free.
- `slot_b` (Number) Optional interface slot on node B (dst), if not provided use next free.

### Read-Only

- `id` (String) Link ID (UUID).
- `interface_a` (String) Interface ID containing the node (UUID).
- `interface_b` (String) Interface ID containing the node (UUID).
- `label` (String) link label (auto generated).
- `link_capture_key` (String) link capture key (when running).
- `state` (String) Link state.

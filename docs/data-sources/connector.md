---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cml2_connector Data Source - terraform-provider-cml2"
subcategory: ""
description: |-
  A data source that retrieves external connectors information from the controller.
---

# cml2_connector (Data Source)

A data source that retrieves external connectors information from the controller.



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `label` (String) A connector label to filter the connector list returned by the controller. Connector labels must be unique, so it's either one group or no group at all if a name filter is provided.
- `tag` (String) A tag name to filter the groups list returned by the controller. Connector tags can be defined on multiple connectors, a list can be returned.

### Read-Only

- `connectors` (Attributes List) A list of all permission groups available on the controller. (see [below for nested schema](#nestedatt--connectors))
- `id` (String) A UUID. The presence of the ID attribute is mandated by the framework. The attribute is a random UUID and has no actual significance.

<a id="nestedatt--connectors"></a>
### Nested Schema for `connectors`

Read-Only:

- `device_name` (String) the actual (Linux network) device name of the external connector.
- `id` (String) External connector identifier, a UUID.
- `label` (String) The label of the external connector, like "NAT" or "System Bridge".
- `protected` (Boolean) Whether the connector is protected, e.g. BPDUs are filtered or not.
- `snooped` (Boolean) True if the IP address snooper listens on this connector.
- `tags` (Set of String) The external connector tag set.

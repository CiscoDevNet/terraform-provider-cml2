---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cml2_user Resource - terraform-provider-cml2"
subcategory: ""
description: |-
  A resource which handles users.
---

# cml2_user (Resource)

A resource which handles users.

## Example Usage

```terraform
# Create a user account

resource "cml2_user" "student1" {
  username    = "student1"
  password    = "secret"
  fullname    = "Joe Learner"
  email       = "student1@cml.lab"
  description = "This is the Student 1 account"
  is_admin    = false
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `password` (String, Sensitive) Password of the user.
- `username` (String) Login name of the user.

### Optional

- `description` (String) Description of the user.
- `email` (String) E-mail address of the user.
- `fullname` (String) Full name of the user.
- `groups` (Set of String) Set of group IDs where the user is member of.
- `is_admin` (Boolean) True if the user has admin rights.
- `resource_pool` (String) Resource pool ID, if any.

### Read-Only

- `directory_dn` (String) Directory DN of the user (when using LDAP).
- `id` (String) User ID (UUID).
- `labs` (Set of String) Set of lab IDs the user owns.
- `opt_in` (Boolean) True if has opted in to sending telemetry data.

# Cisco CML2 Terraform Provider Changelog

Lists the changes in the provider.

## Unreleased

- reverted `skip_verify` to pre 0.5.1 logic, fixed documentation for it
- framework 1.1.1, added better provider description

## Version 0.5.1

### Breaking changes

- link resources: change the attribute names for slots from `node_a_slot` to `slot_a`
  and `node_b_slot` to `slot_b`
- image definition data source: change the attribute name of the node
  definition filter from `node_definition_id` to `nodedefinition` for consistency

### Other changes

- refactor the node resource update logic, allow change of more properties when
  node hasn't been started yet
- make link node changes (e.g. changing the ID of either end of the link) requires
  a replace now.
- fix node tag handling (also a regression in cmlclient)
- improve test coverage
- documentation improvements
- dependency updates
- fix `skip_verify` flag default

## Version 0.5.0

- refactor to work with Terraform Plugin Framework 1.0.1
- added an image definition data source
- removed the GH secret tool from the repo
- dependency updates

## Version 0.4.2

- fixed node properties (compute ID and VNC key)
- added the combine workflow action (internal)
- bumped gocmlclient to 0.0.6
- dependency updates

## Version 0.4.1

- documentation consistency and small fixes
- improved integration test coverage
- fix `lifecycle` import-mode regression
- updated to latest gocmlclient 0.0.4

## Version 0.4.0

- adapted to the terraform-provider-framework v0.0.15 changes
- udpated documentation
- `cml2_node.label` is now a required attribute
- `cml2_lifecycle.lab_id` replaces the `id` for consistency.  This will break
  existing lifecycle resources but it's easy to fix: just rename the `id` to
  `lab_id`.
- `cml2_lifecycle.id` is now auto-generated as a UUIDv4
- `cml2_lifecycle` resource logic changes... produces more concise change sets
- added a `cml2_lifecycle` acceptance test
- renamed the `mkkey` tool to `ghsecret`, adapted the `tunnel.sh` script

## Version 0.3.2

This releases is a large refactor of the initial code base.  It provides suppport
for a few resources and data sources:

- resources
  - `cml2_lab` manages labs (the top level element)
  - `cml2_node` manages nodes as elements of labs
  - `cml2_link` manages links connecting nodes as elements of labs
  - `cml2_lifecycle` manages the lifecycle of labs either by importing existing topology files or by referencing created labs via the lab resource
- data sources
  - `cml2_lab` reads an existing lab
  - `cml2_node` reads an existing node

In addition, the layout of the code base has been significantly changed.

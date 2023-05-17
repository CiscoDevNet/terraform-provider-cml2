# Cisco CML2 Terraform Provider Changelog

Lists the changes in the provider.

## Version 0.6.2

- The provider (via gocmlclient) now honors proxy configuration provided via environment variables `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` (or the lowercase versions thereof). `HTTPS_PROXY` takes precedence over `HTTP_PROXY` for https requests.
- bump gocmlclient to 0.0.21 (better handling of error conditions / proxy use)
- added `ignore_errors` attribute to the system data source to be able to simply ignore errors when waiting for te controller to provide status.

## Version 0.6.1

- allow dynamic configuration of the provider by introducing a "dynamic_config" provider config flag. This defaults to `false`. When set to `true` then the provider configuration is only validated when actual resources are read or created. This is for specific use cases like AWS deployments where the CML2 instance IP is only known after the EC2 instance has been created.
- bump the gocmlclient to 0.0.18

## Version 0.6.0

- allow empty node configurations (fixes #50)
- new data sources
  - "system" for version and ready check, can be used with a timeout to wait until the system becomes ready
  - "groups" to retrieve a list of groups from the system
  - "users" to retrieve a list of users from the system
- new resources
  - "user" for user operations
  - "group" for group operations
- changed "lists" to "sets" where applicable. That said, IP addresses should likely also be treated as sets (unordered, unique values) and not as (ordered) lists
- set correct ID for node and lab data sources
- updated dependencies

## Version 0.5.3

- fixed #40, `skip_verify` documentation
- bumped golang to 1.19
- bumped gocmlclient to 0.0.14
- for integration tests, don't use cache (there's still a race)
- bumped all direct dependencies
- fixed `modify_plan` for node, to include cpu limit and imagedefinition

## Version 0.5.2

- reverted `skip_verify` to pre 0.5.1 logic, fixed documentation for it
- framework 1.1.1, added better provider description
- fixed #38 to allow adding of links without specifying slots
- dependency updates

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

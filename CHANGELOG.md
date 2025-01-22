# Cisco CML2 Terraform Provider Changelog

Lists the changes in the provider.

## Version 0.8.3

- only check the CML host address a) if not dynamic provider configuration and
  b) when initializing the CML client. This fixes an issue with the cloud-cml
  deployment tooling where dynamic provider configuration is used.
- GH release action: take go version from go.mod

## Version 0.8.2

- go version 1.22 used
- updated dependencies (plugin-framework 1.13.0, plugin-go 0.25.0, plugin-testing 1.10.0)
- use newer gocmlclient which supports CML 2.8.0
- added signing public key to repo
- provider configuration check: Ensure valid CML host address (HTTPS)
- moved package name from rschmied to ciscodevnet (cosmetic)
- updated documentation

## Version 0.8.1

- updated documentation, examples and tests to match code changes.

## Version 0.8.0

- update gocmlclient to 0.1.0, supporting CML 2.7.0
- tested w/ 2.7.1
- ability to use named configurations (added w/ 2.7.0) (partially addresses #100)
- deprecate `use_cache` capability, there's a flag in the provider configuration to turn this on (fixes #94)
- replace link resource when changing link slots (fixes #95)
- update all dependencies as of July 2024
- run acceptance tests with 1.7 and 1.8 instead of 1.4 and 1.6
- removed github.com/hashicorp/terraform-plugin-sdk/v2 as a direct dependency, use TF framework testing module instead of SDK v2 testing module
- configurations which only differ in line endings (CR/LF vs LF, DOS / Unix) are now equivalent (fixes #106)
- deprecate `elements` in the `lifecycle` resource, replaced by using standard `depends_on`
- properly formatted tunnel.sh script
- allow to use a token with the integration tests (less requests, faster)

## Version 0.7.0

- added support for "hide links" node resource property, fixes #80
- added external connector data source
- fixed integration test for groups data source
- return error for external connector configuration when device name is provided instead of "NAT" or "System Bridge", fixes #81
- updated all package dependencies
- fix CPULimit property for UMS and ExtConn (they are now always NULL starting with 2.6.0)
- some cosmetic and test changes
- added a add-to-booted-lab lifecycle test (addresses #75 but can't reproduce)
- formatted code base w/ gofumpt
- removed cmlclient go.mod local replace and updated cmlclient to 0.0.22 in go.mod
- added / updated docs and ran generate
- bumped go version to 1.21 in the workflows
- added an extconn schema test
- updated gh action components, only run one test suite in parallel

## Version 0.6.2

- The provider (via gocmlclient) now honors proxy configuration provided via environment variables `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` (or the lowercase versions thereof).  `HTTPS_PROXY` takes precedence over `HTTP_PROXY` for https requests.
- bump gocmlclient to 0.0.21 (better handling of error conditions / proxy use)
- added `ignore_errors` attribute to the system data source to be able to simply ignore errors when waiting for the controller to provide status.

## Version 0.6.1

- allow dynamic configuration of the provider by introducing a `dynamic_config` provider config flag.  This defaults to `false`.  When set to `true` then the provider configuration is only validated when actual resources are read or created.  This is for specific use cases like AWS deployments where the CML2 instance IP is only known after the EC2 instance has been created.
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
- changed "lists" to "sets" where applicable.  That said, IP addresses should likely also be treated as sets (unordered, unique values) and not as (ordered) lists
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

- reverted `skip_verify` to pre-0.5.1 logic, fixed documentation for it
- framework 1.1.1, added better provider description
- fixed #38 to allow adding of links without specifying slots
- dependency updates

## Version 0.5.1

### Breaking changes

- link resources: change the attribute names for slots from `node_a_slot` to `slot_a` and `node_b_slot` to `slot_b`
- image definition data source: change the attribute name of the node definition filter from `node_definition_id` to `nodedefinition` for consistency

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
- `cml2_lifecycle.lab_id` replaces the `id` for consistency.  This will break existing lifecycle resources but it's easy to fix: just rename the `id` to `lab_id`.
- `cml2_lifecycle.id` is now auto-generated as a UUIDv4
- `cml2_lifecycle` resource logic changes... produces more concise change sets
- added a `cml2_lifecycle` acceptance test
- renamed the `mkkey` tool to `ghsecret`, adapted the `tunnel.sh` script

## Version 0.3.2

This releases is a large refactor of the initial code base.  It provides support
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

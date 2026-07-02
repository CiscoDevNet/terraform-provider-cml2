# Examples

This directory contains examples that are mostly used for documentation, but can
also be run/tested manually via the Terraform CLI.

## Lifecycle examples

- `resources/cml2_lifecycle/resource.tf`: a full lab with staged start-up, configs, and `update_triggers`
- `resources/cml2_lifecycle/eleven-nodes.tf`: a larger staged-start example with several nodes and links
- `resources/cml2_lifecycle/topology_import.tf`: importing an existing topology into lifecycle management
- `resources/cml2_lifecycle/update_triggers.tf`: how to wire `update_triggers` so lifecycle restarts when a node is replaced or reconfigured

> **Note**: For lifecycle examples that build the topology in HCL, it is usually safer to make the lifecycle resource depend on both the nodes and the links. That keeps the example deterministic and helps ensure the full topology is present before the lab is started.

## Other examples

- `resources/cml2_annotation`: classic annotation examples (text/rectangle/line), including `null` line endings

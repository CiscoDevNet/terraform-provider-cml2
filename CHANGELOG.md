# Cisco CML2 Terraform Provider Changelog

Lists the changes in the provider.

## Version 0.0.3

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

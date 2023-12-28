# Get a list of all connectors on the system with either the provided tag (the
# result can have multiple elements) or with a specific name/label (the result
# has exactly one or zero elements if the label does not exist).

data "cml2_connector" "nat" {
  tag = "NAT"
  # Alternatively (or in combination, logical AND):
  # label = "System Bridge"
}

output "nat_connector" {
  # The label can be used as the configuration of an external connector node
  # The ID is mostly for internal use.
  value = data.cml2_connector.nat.connectors[0].label
}

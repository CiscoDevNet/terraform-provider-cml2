package common

import "github.com/rschmied/gocmlclient/pkg/models"

// NodeDefinition is the internal string identifier used for node types.
// Examples: "alpine", "iosv", "ioll2-xe", "external_connector".
type NodeDefinition = string

// IsBuiltInNodeDefinition reports whether nodeDef is a built-in definition ID.
func IsBuiltInNodeDefinition(nodeDef NodeDefinition) bool {
	switch nodeDef {
	case "external_connector", "unmanaged_switch":
		return true
	default:
		return false
	}
}

// NodeDefLibvirtDomainDriver returns the libvirt domain driver for a node
// definition. Empty means "not a libvirt/qemu VM" (e.g. docker-style nodes).
func NodeDefLibvirtDomainDriver(nd models.NodeDefinition) string {
	return nd.Sim.LinuxNative.LibvirtDomainDriver
}

// NodeDefLinuxDriver returns the simulator driver value (often matches the
// backend implementation, e.g. qemu for VM-backed nodes).
func NodeDefLinuxDriver(nd models.NodeDefinition) string {
	return nd.Sim.LinuxNative.Driver
}

// NodeDefIsLibvirtBacked reports whether the node definition is VM-backed
// (libvirt/qemu) based on the libvirt_domain_driver field.
func NodeDefIsLibvirtBacked(nd models.NodeDefinition) bool {
	return NodeDefLibvirtDomainDriver(nd) != ""
}

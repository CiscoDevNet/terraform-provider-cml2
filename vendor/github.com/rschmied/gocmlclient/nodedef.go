package cmlclient

import "context"

type definitionID string

type NodeDefinition struct {
	ID            definitionID   `json:"id"`
	Configuration map[string]any `json:"configuration"`
	Device        deviceData     `json:"device"`
	Inherited     map[string]any `json:"inherited"`
	SchemaVersion string         `json:"schema_version"`
	Sim           simData        `json:"sim"`
	Boot          map[string]any `json:"boot"`
	PyATS         map[string]any `json:"pyats"`
	General       map[string]any `json:"general"`
	UI            map[string]any `json:"ui"`
}

type simData struct {
	RAM      int  `json:"ram"`
	Console  bool `json:"console"`
	Simulate bool `json:"simulate"`
	VNC      bool `json:"vnc"`
}

type deviceData struct {
	Interfaces interfaceData `json:"interfaces"`
}

type interfaceData struct {
	HasLoopbackZero bool     `json:"has_loopback_zero"`
	DefaultCount    int      `json:"default_count"`
	Physical        []string `json:"physical"`
	SerialPorts     int      `json:"serial_ports"`
}

type NodeDefinitionMap map[definitionID]NodeDefinition

func (nd NodeDefinition) hasVNC() bool {
	return nd.Sim.VNC
}

func (nd NodeDefinition) hasSerial() bool {
	return nd.Device.Interfaces.SerialPorts > 0
}

func (nd NodeDefinition) serialPorts() int {
	return nd.Device.Interfaces.SerialPorts
}

// NodeDefinitions returns the list of node definitions available on the CML
// controller. The key of the map is the definition type name (e.g. "alpine" or
// "ios"). The node def data structure is incomplete, only essential fields are
// populated.
func (c *Client) NodeDefinitions(ctx context.Context) (NodeDefinitionMap, error) {
	nd := []NodeDefinition{}
	err := c.jsonGet(ctx, "simplified_node_definitions", &nd, 0)
	if err != nil {
		return nil, err
	}

	nodeDefMap := make(NodeDefinitionMap)
	for _, nodeDef := range nd {
		nodeDefMap[nodeDef.ID] = nodeDef
	}
	return nodeDefMap, nil
}

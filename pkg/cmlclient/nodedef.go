package cmlclient

import "context"

type definitionID string

type NodeDefinition struct {
	ID      definitionID           `json:"id"`
	General map[string]interface{} `json:"general"`
	Device  deviceData             `json:"device"`
	UI      map[string]interface{} `json:"ui"`
	Sim     simData                `json:"sim"`
	Images  []string               `json:"image_definitions"`
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

type simplifiedNodeDefinitionList map[definitionID]NodeDefinition

func (nd NodeDefinition) hasVNC() bool {
	return nd.Sim.VNC
}

func (nd NodeDefinition) hasSerial() bool {
	return nd.Device.Interfaces.SerialPorts > 0
}

func (nd NodeDefinition) serialPorts() int {
	return nd.Device.Interfaces.SerialPorts
}

func (c *Client) GetNodeDefs(ctx context.Context) ([]NodeDefinition, error) {
	nd := []NodeDefinition{}
	err := c.jsonGet(ctx, "simplified_node_definitions", &nd)
	if err != nil {
		return nil, err
	}
	return nd, nil
}

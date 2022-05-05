package cmlclient

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetLab(id string) (*Lab, error) {
	api := fmt.Sprintf("labs/%s", id)
	req, err := c.apiRequest(http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.doAPI(req)
	if err != nil {
		return nil, err
	}

	lab := &Lab{}
	err = json.Unmarshal(res, lab)
	if err != nil {
		return nil, err
	}

	nl, err := c.getNodesForLab(id)
	if err != nil {
		return nil, err
	}
	lab.Nodes = nl

	for _, node := range lab.Nodes {
		il, err := c.getInterfacesForNode(id, node.ID)
		if err != nil {
			return nil, err
		}
		node.Interfaces = il
	}

	l3info, err := c.getL3Info(id)
	if err != nil {
		return nil, err
	}

	// we could use maps to lookup/merge the data...
	for _, l3data := range *l3info {
		for _, l3i := range l3data.Interfaces {
			for _, node := range lab.Nodes {
				for _, iface := range node.Interfaces {
					if iface.ID == l3i.ID {
						iface.IP4 = l3i.IP4
						iface.IP6 = l3i.IP6
					}
				}
			}
		}
	}

	links, err := c.getLinksForLab(lab)
	if err != nil {
		return nil, err
	}
	lab.Links = links

	return lab, nil
}

func (c *Client) getNodesForLab(id string) ([]*Node, error) {
	api := fmt.Sprintf("labs/%s/nodes", id)
	req, err := c.apiRequest(http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.doAPI(req)
	if err != nil {
		return nil, err
	}
	nodeIDlist := &IDlist{}
	err = json.Unmarshal(res, nodeIDlist)
	if err != nil {
		return nil, err
	}

	nodeList := []*Node{}
	for _, nodeID := range *nodeIDlist {
		api = fmt.Sprintf("labs/%s/nodes/%s", id, nodeID)
		req, err := c.apiRequest(http.MethodGet, api, nil)
		if err != nil {
			return nil, err
		}
		res, err := c.doAPI(req)
		if err != nil {
			return nil, err
		}
		node := &Node{}
		err = json.Unmarshal(res, node)
		if err != nil {
			return nil, err
		}
		nodeList = append(nodeList, node)
	}
	return nodeList, nil
}

func (c *Client) getInterfacesForNode(id, nodeID string) ([]*Interface, error) {
	api := fmt.Sprintf("labs/%s/nodes/%s/interfaces", id, nodeID)
	req, err := c.apiRequest(http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.doAPI(req)
	if err != nil {
		return nil, err
	}
	interfaceIDlist := &IDlist{}
	err = json.Unmarshal(res, interfaceIDlist)
	if err != nil {
		return nil, err
	}

	interfaceList := []*Interface{}
	for _, nodeID := range *interfaceIDlist {
		api = fmt.Sprintf("labs/%s/interfaces/%s", id, nodeID)
		req, err := c.apiRequest(http.MethodGet, api, nil)
		if err != nil {
			return nil, err
		}
		res, err := c.doAPI(req)
		if err != nil {
			return nil, err
		}
		iface := &Interface{}
		err = json.Unmarshal(res, iface)
		if err != nil {
			return nil, err
		}
		interfaceList = append(interfaceList, iface)
	}
	return interfaceList, nil
}

func (c *Client) getLinksForLab(lab *Lab) ([]*Link, error) {
	api := fmt.Sprintf("labs/%s/links", lab.ID)
	req, err := c.apiRequest(http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.doAPI(req)
	if err != nil {
		return nil, err
	}
	linkIDlist := &IDlist{}
	err = json.Unmarshal(res, linkIDlist)
	if err != nil {
		return nil, err
	}

	linkList := []*Link{}
	for _, linkID := range *linkIDlist {
		api = fmt.Sprintf("labs/%s/links/%s", lab.ID, linkID)
		req, err := c.apiRequest(http.MethodGet, api, nil)
		if err != nil {
			return nil, err
		}
		res, err := c.doAPI(req)
		if err != nil {
			return nil, err
		}
		link := &link{}
		err = json.Unmarshal(res, link)
		if err != nil {
			return nil, err
		}
		realLink := &Link{
			ID:      link.ID,
			State:   link.State,
			Label:   link.Label,
			PCAPkey: link.PCAPkey,
			Src:     c.findInterface(lab.Nodes, link.SrcID),
			Dst:     c.findInterface(lab.Nodes, link.DstID),
		}
		linkList = append(linkList, realLink)
	}
	return linkList, nil
}

func (c *Client) getL3Info(id string) (*l3nodes, error) {
	api := fmt.Sprintf("labs/%s/layer3_addresses", id)
	req, err := c.apiRequest(http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.doAPI(req)
	if err != nil {
		return nil, err
	}
	l3n := &l3nodes{}
	err = json.Unmarshal(res, l3n)
	if err != nil {
		return nil, err
	}
	return l3n, nil
}

func (c *Client) findInterface(nodes []*Node, id string) *Interface {
	for _, node := range nodes {
		for _, iface := range node.Interfaces {
			if iface.ID == id {
				return iface
			}
		}
	}
	return nil
}

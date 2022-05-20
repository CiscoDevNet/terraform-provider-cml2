package cmlclient

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"
)

// {
// 	"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 	"state": "DEFINED_ON_CORE",
// 	"created": "2021-02-28T07:33:47+00:00",
// 	"modified": "2021-02-28T07:33:47+00:00",
// 	"lab_title": "Lab at Mon 17:27 PM",
// 	"owner": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 	"lab_description": "string",
// 	"node_count": 0,
// 	"link_count": 0,
// 	"lab_notes": "string",
// 	"groups": [
// 	  {
// 		"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 		"permission": "read_only"
// 	  }
// 	]
// }

const (
	LabStateDefined = "DEFINED_ON_CORE"
	LabStateStopped = "STOPPED"
	LabStateStarted = "STARTED"
)

type IDlist []string
type nodeMap map[string]*Node
type interfaceMap map[string]*Interface
type linkList []*Link

type labAlias struct {
	Lab
	OwnerID string `json:"owner"`
}

type Lab struct {
	ID          string   `json:"id"`
	State       string   `json:"state"`
	Created     string   `json:"created"`
	Modified    string   `json:"modified"`
	Title       string   `json:"lab_title"`
	Description string   `json:"lab_description"`
	Notes       string   `json:"lab_notes"`
	Owner       *User    `json:"owner"`
	NodeCount   int      `json:"node_count"`
	LinkCount   int      `json:"link_count"`
	Nodes       nodeMap  `json:"nodes"`
	Links       linkList `json:"links"`
}

func (l *Lab) CanBeWiped() bool {
	if len(l.Nodes) == 0 {
		return l.State != LabStateDefined
	}
	for _, node := range l.Nodes {
		if node.State != LabStateDefined {
			return true
		}
	}
	return false
}

func (l *Lab) Running() bool {
	for _, node := range l.Nodes {
		if node.State != LabStateDefined {
			return true
		}
	}
	return false
}

type LabImport struct {
	ID       string   `json:"id"`
	Warnings []string `json:"warnings"`
}

func (c *Client) ImportLab(ctx context.Context, topo string) (*Lab, error) {
	topoReader := strings.NewReader(topo)
	labImport := &LabImport{}
	err := c.jsonPost(ctx, "import", topoReader, labImport)
	if err != nil {
		return nil, err
	}
	lab, err := c.GetLab(ctx, labImport.ID, true)
	if err != nil {
		return nil, err
	}
	return lab, nil
}

func (c *Client) StartLab(ctx context.Context, id string) error {
	api := fmt.Sprintf("labs/%s/start", id)
	req, err := c.apiRequest(ctx, http.MethodPut, api, nil)
	if err != nil {
		return err
	}
	_, err = c.doAPI(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ConvergedLab(ctx context.Context, id string) (bool, error) {
	api := fmt.Sprintf("labs/%s/check_if_converged", id)
	converged := false
	err := c.jsonGet(ctx, api, &converged)
	if err != nil {
		return false, err
	}
	return converged, nil
}

func (c *Client) StopLab(ctx context.Context, id string) error {
	api := fmt.Sprintf("labs/%s/stop", id)
	req, err := c.apiRequest(ctx, http.MethodPut, api, nil)
	if err != nil {
		return err
	}
	_, err = c.doAPI(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) WipeLab(ctx context.Context, id string) error {
	api := fmt.Sprintf("labs/%s/wipe", id)
	req, err := c.apiRequest(ctx, http.MethodPut, api, nil)
	if err != nil {
		return err
	}
	_, err = c.doAPI(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DestroyLab(ctx context.Context, id string) error {
	api := fmt.Sprintf("labs/%s", id)
	req, err := c.apiRequest(ctx, http.MethodDelete, api, nil)
	if err != nil {
		return err
	}
	_, err = c.doAPI(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetLab(ctx context.Context, id string, shallow bool) (*Lab, error) {
	api := fmt.Sprintf("labs/%s", id)
	la := &labAlias{}
	err := c.jsonGet(ctx, api, la)
	if err != nil {
		return nil, err
	}

	if shallow {
		return &la.Lab, nil
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer log.Printf("user done")
		la.Owner, err = c.getUser(ctx, la.OwnerID)
		if err != nil {
			return err
		}
		return nil
	})

	lab := &la.Lab

	// need to ensure that this block finishes before the others run
	ch := make(chan struct{})
	g.Go(func() error {
		defer func() {
			log.Printf("nodes/interfaces done")
			// two sync points, we can run the API endpoints but we need to
			// wait for the node data to be read until we can add the layer3
			// info (1) and the link info (2)
			ch <- struct{}{}
			ch <- struct{}{}
		}()
		err := c.getNodesForLab(ctx, lab)
		if err != nil {
			return err
		}
		for _, node := range lab.Nodes {
			err = c.getInterfacesForNode(ctx, id, node)
			if err != nil {
				return err
			}
		}
		return nil
	})

	g.Go(func() error {
		defer log.Printf("l3info done")
		l3info, err := c.getL3Info(ctx, id)
		if err != nil {
			return err
		}
		log.Printf("l3info read")
		// wait for node data read complete
		<-ch
		// map and merge the l3 data...
		for nid, l3data := range *l3info {
			for mac, l3i := range l3data.Interfaces {
				if node, found := lab.Nodes[nid]; found {
					if iface, found := node.Interfaces[l3i.ID]; found {
						if iface.MACaddress == mac {
							iface.IP4 = l3i.IP4
							iface.IP6 = l3i.IP6
						}
					}
				}
			}
		}
		return nil
	})

	g.Go(func() error {
		defer log.Printf("links done")
		idlist, err := c.getLinkIDsForLab(ctx, lab)
		if err != nil {
			return err
		}
		log.Printf("linkidlist read")
		// wait for node data read complete
		<-ch
		return c.getLinksForLab(ctx, lab, idlist)
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	log.Printf("wait done")

	return lab, nil
}

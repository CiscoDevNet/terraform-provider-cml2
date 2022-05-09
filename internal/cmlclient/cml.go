package cmlclient

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type Client struct {
	HttpClient *http.Client
	APIkey     string
	Host       string
	Base       string
	ctx        context.Context
}

func NewClientWithContext(ctx context.Context, host, apiKey string, insecure bool) *Client {
	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecure,
	}

	return &Client{
		HttpClient: &http.Client{
			Timeout:   15 * time.Second,
			Transport: tr,
		},
		Host:   host,
		APIkey: apiKey,
		Base:   DefaultAPIBase,
		ctx:    ctx,
	}
}

func NewClient(host, apiKey string, insecure bool) *Client {
	return NewClientWithContext(context.Background(), host, apiKey, insecure)
}

func (c *Client) ImportLab(topo string) (*Lab, error) {
	api := "import"
	topoReader := strings.NewReader(topo)
	req, err := c.apiRequest(c.ctx, http.MethodPost, api, topoReader)
	if err != nil {
		return nil, err
	}
	res, err := c.doAPI(req)
	if err != nil {
		return nil, err
	}
	labImport := &LabImport{}
	err = json.Unmarshal(res, labImport)
	if err != nil {
		return nil, err
	}

	lab, err := c.GetLab(labImport.ID, true)
	if err != nil {
		return nil, err
	}

	return lab, nil
}

func (c *Client) StartLab(id string) error {
	api := fmt.Sprintf("labs/%s/start", id)
	req, err := c.apiRequest(c.ctx, http.MethodPut, api, nil)
	if err != nil {
		return err
	}
	_, err = c.doAPI(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ConvergedLab(id string) (bool, error) {
	api := fmt.Sprintf("labs/%s/check_if_converged", id)
	converged := false
	err := c.jsonGet(c.ctx, api, &converged)
	if err != nil {
		return false, err
	}
	return converged, nil
}

func (c *Client) StopLab(id string) error {
	api := fmt.Sprintf("labs/%s/stop", id)
	req, err := c.apiRequest(c.ctx, http.MethodPut, api, nil)
	if err != nil {
		return err
	}
	_, err = c.doAPI(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) WipeLab(id string) error {
	api := fmt.Sprintf("labs/%s/wipe", id)
	req, err := c.apiRequest(c.ctx, http.MethodPut, api, nil)
	if err != nil {
		return err
	}
	_, err = c.doAPI(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DestroyLab(id string) error {
	api := fmt.Sprintf("labs/%s", id)
	req, err := c.apiRequest(c.ctx, http.MethodDelete, api, nil)
	if err != nil {
		return err
	}
	_, err = c.doAPI(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetLab(id string, shallow bool) (*Lab, error) {
	api := fmt.Sprintf("labs/%s", id)
	la := &labAlias{}
	err := c.jsonGet(c.ctx, api, la)
	if err != nil {
		return nil, err
	}

	if shallow {
		return &la.Lab, nil
	}

	g, ctx := errgroup.WithContext(c.ctx)

	g.Go(func() error {
		defer fmt.Fprintln(os.Stderr, "### user done")
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
			fmt.Fprintln(os.Stderr, "### nodes/interfaces done")
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
		defer fmt.Fprintln(os.Stderr, "### l3info done")
		l3info, err := c.getL3Info(ctx, id)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "### l3info read")
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
		defer fmt.Fprintln(os.Stderr, "### links done")
		idlist, err := c.getLinkIDsForLab(ctx, lab)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "### linkidlist read")
		// wait for node data read complete
		<-ch
		return c.getLinksForLab(ctx, lab, idlist)
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	fmt.Fprintln(os.Stderr, "### wait done")

	return lab, nil
}

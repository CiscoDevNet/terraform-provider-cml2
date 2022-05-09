package cmlclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	ContentType           = "application/json"
	DefaultAPIBase string = "/api/v0/"
)

func (c *Client) apiRequest(ctx context.Context, method string, path string, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s%s%s", c.Host, c.Base, path),
		data,
	)
	if err != nil {
		return nil, err
	}

	// set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIkey))
	req.Header.Set("Accept", ContentType)
	if data != nil {
		req.Header.Set("Content-Type", ContentType)
	}

	return req, nil
}

func (c *Client) doAPI(req *http.Request) ([]byte, error) {
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNoContent {
		return body, err
	} else {
		return nil, fmt.Errorf("status: %d, %s", res.StatusCode, body)
	}
}

func (c *Client) jsonGet(ctx context.Context, api string, data interface{}) error {
	req, err := c.apiRequest(ctx, http.MethodGet, api, nil)
	if err != nil {
		return err
	}
	res, err := c.doAPI(req)
	if err != nil {
		return err
	}
	err = json.Unmarshal(res, data)
	if err != nil {
		return err
	}
	return nil
}

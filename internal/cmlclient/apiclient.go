package cmlclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	contentType string = "application/json"
	apiBase     string = "/api/v0/"
	authAPI     string = "auth_extended"
)

func setTokenHeader(req *http.Request, token string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

func (c *Client) apiRequest(ctx context.Context, method string, path string, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s%s%s", c.host, apiBase, path),
		data,
	)
	if err != nil {
		return nil, err
	}

	// set headers (this avoids a loop when actually authenticating)
	if path != authAPI {
		setTokenHeader(req, c.apiToken)
	}
	req.Header.Set("Accept", contentType)
	if data != nil {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

func (c *Client) doAPI(ctx context.Context, req *http.Request) ([]byte, error) {
	retrying := false

	if !c.versionChecked {
		c.versionChecked = true
		c.compatible = c.versionCheck(ctx)
	}
	if c.compatible != nil {
		return nil, c.compatible
	}

retry:
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// no authorization and not retrying already
	if res.StatusCode == http.StatusUnauthorized {
		if retrying || !c.userpass.valid() {
			return nil, errors.New("can't authorize")
		}
		retrying = true
		log.Println("need to authenticate")
		err = c.authenticate(ctx, c.userpass)
		if err != nil {
			return nil, err
		}
		setTokenHeader(req, c.apiToken)
		goto retry
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNoContent {
		return body, err
	} else {
		return nil, fmt.Errorf("status: %d, %s", res.StatusCode, body)
	}
}

func (c *Client) jsonGet(ctx context.Context, api string, result interface{}) error {
	return c.jsonReq(ctx, http.MethodGet, api, nil, result)
}

func (c *Client) jsonPost(ctx context.Context, api string, data io.Reader, result interface{}) error {
	return c.jsonReq(ctx, http.MethodPost, api, data, result)
}

func (c *Client) jsonReq(ctx context.Context, method, api string, data io.Reader, result interface{}) error {
	req, err := c.apiRequest(ctx, method, api, data)
	if err != nil {
		return err
	}
	res, err := c.doAPI(ctx, req)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	err = json.Unmarshal(res, result)
	if err != nil {
		return err
	}
	return nil
}

package cmlclient

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	ContentType           = "application/json"
	DefaultAPIBase string = "/api/v0/"
)

type Client struct {
	HttpClient *http.Client
	ApiKey     string
	Host       string
	Base       string
}

func NewCMLClient(host, apiKey string, insecure bool) *Client {

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
		ApiKey: apiKey,
		Base:   DefaultAPIBase,
	}
}

func (c *Client) apiRequest(method string, path string, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s%s%s", c.Host, c.Base, path),
		data,
	)
	if err != nil {
		return nil, err
	}

	// set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiKey))
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
		return nil, fmt.Errorf("status: %d", res.StatusCode)
	}
}

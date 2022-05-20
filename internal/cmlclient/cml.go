package cmlclient

import (
	"crypto/tls"
	"net/http"
	"time"
)

type apiClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient     apiClient
	apiToken       string
	host           string
	userpass       userPass
	versionChecked bool
	compatible     error
	authChecked    bool
}

func NewClient(host string, insecure bool) *Client {
	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecure,
	}

	return &Client{
		host:     host,
		apiToken: "",
		userpass: userPass{},
		httpClient: &http.Client{
			Timeout:   15 * time.Second,
			Transport: tr,
		},
		versionChecked: false,
	}
}

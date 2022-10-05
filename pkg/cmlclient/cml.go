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
	host             string
	apiToken         string
	userpass         userPass
	httpClient       apiClient
	authChecked      bool
	versionChecked   bool
	compatibilityErr error
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
		authChecked:      false,
		versionChecked:   false,
		compatibilityErr: nil,
	}
}

package testing

import (
	"fmt"
	"os"

	gocml "github.com/rschmied/gocmlclient"
	"github.com/rschmied/gocmlclient/pkg/client"
)

// NewCMLClientFromTFEnv builds a gocmlclient from the same TF_VAR_* env vars
// used by the acceptance test configs.
//
// Required:
// - TF_VAR_address
// - either TF_VAR_token OR (TF_VAR_username + TF_VAR_password)
func NewCMLClientFromTFEnv() (*client.Client, error) {
	addr := os.Getenv("TF_VAR_address")
	username := os.Getenv("TF_VAR_username")
	password := os.Getenv("TF_VAR_password")
	token := os.Getenv("TF_VAR_token")

	if addr == "" {
		return nil, fmt.Errorf("TF_VAR_address must be set")
	}
	if token == "" && (username == "" || password == "") {
		return nil, fmt.Errorf("either TF_VAR_token or TF_VAR_username+TF_VAR_password must be set")
	}

	opts := []gocml.Option{gocml.SkipReadyCheck(), gocml.WithInsecureTLS()}
	if token != "" {
		opts = append(opts, gocml.WithStaticToken(token))
	} else {
		opts = append(opts, gocml.WithUsernamePassword(username, password))
	}

	client, err := gocml.New(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating gocmlclient: %w", err)
	}
	return client, nil
}

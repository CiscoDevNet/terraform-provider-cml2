package cmlclient

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

// the 2.4.0.dev is likely wrong, should be -dev (dash, not dot):
// {
// 	"version": "2.4.0.dev0+build.f904bdf8",
// 	"ready": true
// }
// 2.5.0-dev0+build.3.2f7875762

type systemVersion struct {
	Version string `json:"version"`
	Ready   bool   `json:"ready"`
}

const versionConstraint = ">=2.4.0,<3.0.0"

func versionError(got string) error {
	return fmt.Errorf(
		"server not compatible, want %s, got %s (%w)",
		versionConstraint, got, ErrSystemNotReady,
	)
}

func (c *Client) versionCheck(ctx context.Context, depth int32) error {

	sv := systemVersion{}
	if err := c.jsonGet(ctx, systeminfoAPI, &sv, depth); err != nil {
		return err
	}

	if !sv.Ready {
		return ErrSystemNotReady
	}

	constraint, err := semver.NewConstraint(versionConstraint)
	if err != nil {
		panic("unparsable semver version constant")
	}

	re := regexp.MustCompile(`^(\d\.\d\.\d)((-dev0)?\+build.*)?$`)
	m := re.FindStringSubmatch(sv.Version)
	if m == nil {
		return versionError(sv.Version)
	}
	log.Printf("controller version: %s", sv.Version)
	if len(m[3]) > 0 {
		log.Printf("Warning, this is a DEV version %s", sv.Version)
	}
	stem := m[1]
	v, err := semver.NewVersion(stem)
	if err != nil {
		return err
	}
	// Check if the version meets the constraints
	ok := constraint.Check(v)
	if !ok {
		return versionError(sv.Version)
	}
	c.version = sv.Version
	return nil
}

// Version returns the CML controller version
func (c *Client) Version() string {
	return c.version
}

// Ready returns nil if the system is compatible and ready
func (c *Client) Ready(ctx context.Context) error {
	// we can safely assume depth 0 as the API endpoint does not
	// require authentication
	return c.versionCheck(ctx, 0)
}

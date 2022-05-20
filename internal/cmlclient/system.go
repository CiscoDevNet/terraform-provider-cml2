package cmlclient

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

// {
// 	"version": "2.4.0.dev0+build.f904bdf8",
// 	"ready": true
// }

type systemVersion struct {
	Version string `json:"version"`
	Ready   bool   `json:"ready"`
}

const versionConstraint = ">=2.3.0,<2.5.0"

func versionError(got string) error {
	return fmt.Errorf("server not compatible, want %s, got %s", versionConstraint, got)
}

func (c *Client) versionCheck(ctx context.Context) error {

	sv := systemVersion{}
	if err := c.jsonGet(ctx, systeminfoAPI, &sv); err != nil {
		return err
	}
	constraint, err := semver.NewConstraint(versionConstraint)
	if err != nil {
		panic("unparsable semver version constant")
	}

	re, err := regexp.Compile(`^(\d\.\d\.\d)(.*$)`)
	if err != nil {
		panic("can't compile regex")
	}
	m := re.FindStringSubmatch(sv.Version)
	if m == nil {
		return versionError(sv.Version)
	}
	log.Printf("controller version: %s", sv.Version)
	if len(m[2]) > 0 {
		log.Printf("Warning, this is a DEV version %s", sv.Version)
	}
	stem := m[1]
	v, err := semver.NewVersion(stem)
	if err != nil {
		return err
	}
	// Check if the version meets the constraints. The a variable will be true.
	ok := constraint.Check(v)
	if !ok {
		return versionError(sv.Version)
	}
	return nil
}

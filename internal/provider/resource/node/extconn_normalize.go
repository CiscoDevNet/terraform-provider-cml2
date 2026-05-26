package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// normalizeExtConnConfig maps legacy external connector configuration values
// (device names like "virbr0") to connector labels (like "NAT").
//
// Returns:
// - normalized: the value to use
// - changed: true if a mapping was performed
// - warning: optional warning message to show to users
func normalizeExtConnConfig(ctx context.Context, cfg *common.ProviderConfig, in string) (normalized string, changed bool, warning string, err error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return "", false, "", nil
	}

	// Defensive nil checks to avoid panics when provider/client isn't
	// initialized. Return a descriptive error so callers can surface it.
	// Using a concrete *common.ProviderConfig (not an interface) makes cfg == nil
	// a reliable check — there is no typed-nil ambiguity with a concrete pointer.
	if cfg == nil {
		return in, false, "", fmt.Errorf("list external connectors: provider config is nil")
	}
	cli := cfg.Client()
	if cli == nil {
		return in, false, "", fmt.Errorf("list external connectors: client not initialized")
	}
	if cli.ExtConn == nil {
		return in, false, "", fmt.Errorf("list external connectors: extconn service unavailable")
	}

	connectors, err := cli.ExtConn.List(ctx)
	if err != nil {
		return in, false, "", fmt.Errorf("list external connectors: %w", err)
	}

	// Prefer exact label match (no change).
	for _, c := range connectors {
		if strings.EqualFold(c.Label, in) {
			return c.Label, false, "", nil
		}
	}

	// Map device name -> label.
	for _, c := range connectors {
		if strings.EqualFold(c.DeviceName, in) {
			return c.Label, true, fmt.Sprintf("External connector configuration %q is a device name and deprecated; normalized to label %q.", in, c.Label), nil
		}
	}

	// Unknown; assume user provided a label.
	return in, false, "", nil
}

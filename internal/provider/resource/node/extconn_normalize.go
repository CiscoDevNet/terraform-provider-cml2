package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// normalizeExtConnConfig validates that external connector configuration uses
// a device name (e.g. "virbr0"). Labels (e.g. "NAT") are rejected with a
// corrective error message that includes the matching device name.
func normalizeExtConnConfig(ctx context.Context, cfg *common.ProviderConfig, in string) (string, error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return "", nil
	}

	if cfg == nil {
		return "", fmt.Errorf("list external connectors: provider config is nil")
	}
	cli := cfg.Client()
	if cli == nil {
		return "", fmt.Errorf("list external connectors: client not initialized")
	}
	if cli.ExtConn == nil {
		return "", fmt.Errorf("list external connectors: extconn service unavailable")
	}

	connectors, err := cli.ExtConn.List(ctx)
	if err != nil {
		return "", fmt.Errorf("list external connectors: %w", err)
	}

	for _, c := range connectors {
		if strings.EqualFold(c.DeviceName, in) {
			tflog.Debug(ctx, "extconn config validated as device name", map[string]any{"input": in, "device_name": c.DeviceName})
			return c.DeviceName, nil
		}
	}

	for _, c := range connectors {
		if strings.EqualFold(c.Label, in) {
			tflog.Debug(ctx, "extconn config rejected label", map[string]any{"input": in, "label": c.Label, "device_name": c.DeviceName})
			return "", fmt.Errorf("external connector configuration %q is a label; use device name %q", in, c.DeviceName)
		}
	}

	tflog.Debug(ctx, "extconn config unknown", map[string]any{"input": in})
	return "", fmt.Errorf("external connector configuration %q does not exist", in)
}

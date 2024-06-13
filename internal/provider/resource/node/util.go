package node

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

func setNamedConfigsFromData(ctx context.Context, diag diag.Diagnostics, nm cmlschema.NodeModel) []cmlclient.NodeConfig {
	var configurations []cmlclient.NodeConfig
	if len(nm.Configurations.Elements()) > 0 {
		var nc cmlschema.NamedConfigModel
		for _, el := range nm.Configurations.Elements() {
			diag.Append(tfsdk.ValueAs(ctx, el, &nc)...)
			if diag.HasError() {
				return nil
			}
			cfg := cmlclient.NodeConfig{
				Name:    nc.Name.ValueString(),
				Content: nc.Content.ValueString(),
			}
			configurations = append(configurations, cfg)
		}
	}
	return configurations
}

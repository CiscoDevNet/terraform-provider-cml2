package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func getTimeouts(ctx context.Context, config tfsdk.Config, diags *diag.Diagnostics) ResourceTimeouts {

	timeouts := ResourceTimeouts{}
	diags.Append(config.GetAttribute(ctx, path.Root("timeouts"), &timeouts)...)
	tflog.Info(ctx, fmt.Sprintf("timeouts: %+v", timeouts))
	return timeouts
}

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func getTimeouts(ctx context.Context, config tfsdk.Config, diags *diag.Diagnostics) *ResourceTimeouts {
	// timeouts is optional, if ommitted it will result in a nil pointer
	var timeouts *ResourceTimeouts
	diags.Append(config.GetAttribute(ctx, path.Root("timeouts"), &timeouts)...)
	if diags.HasError() || timeouts == nil {
		tflog.Warn(ctx, "timeouts undefined, using defaults")
		return &ResourceTimeouts{
			Create: types.String{Value: "2h"},
			Delete: types.String{Value: "2h"},
			Update: types.String{Value: "2h"},
		}
	}
	tflog.Info(ctx, fmt.Sprintf("timeouts: %+v", timeouts))
	return timeouts
}

func getStaging(ctx context.Context, config tfsdk.Config, diags *diag.Diagnostics) *ResourceStaging {
	var staging *ResourceStaging
	diags.Append(config.GetAttribute(ctx, path.Root("staging"), &staging)...)
	tflog.Info(ctx, fmt.Sprintf("staging: %+v", staging))
	// default for this is true
	if staging != nil && staging.StartRemaining.IsNull() {
		tflog.Info(ctx, "setting start remaining to true, default value")
		staging.StartRemaining.Null = false
		staging.StartRemaining.Value = true
	}
	return staging
}

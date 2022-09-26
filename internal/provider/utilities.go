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

	dg := config.GetAttribute(ctx, path.Root("timeouts"), &timeouts)
	// if dg.Contains(diag.ErrorDiagnostic{}) {

	// }
	diags.Append(dg...)
	tflog.Info(ctx, fmt.Sprintf("timeouts: %+v", timeouts))
	return timeouts
}

func getStaging(ctx context.Context, config tfsdk.Config, diags *diag.Diagnostics) ResourceStaging {
	staging := ResourceStaging{}
	diags.Append(config.GetAttribute(ctx, path.Root("staging"), &staging)...)
	tflog.Info(ctx, fmt.Sprintf("staging: %+v", staging))
	return staging
}

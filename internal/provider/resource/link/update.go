package link

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

func (r LinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var data cmlschema.LinkModel

	tflog.Info(ctx, "Resource Link UPDATE")
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Warn(ctx, "Resource Link UPDATE: not implemented")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Resource Link UPDATE done")
}

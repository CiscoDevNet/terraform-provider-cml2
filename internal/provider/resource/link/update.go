package link

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r LinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var planData *schema.LinkModel

	tflog.Info(ctx, "Resource Link UPDATE")
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Warn(ctx, "Resource Link UPDATE: not implemented")
	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
	tflog.Info(ctx, "Resource Link UPDATE: done")
}

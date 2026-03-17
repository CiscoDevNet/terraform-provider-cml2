package link

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func (r *LinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cmlschema.LinkModel

	tflog.Info(ctx, "Resource Link READ")

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	link, err := r.cfg.Client().Link.GetByID(ctx, models.UUID(data.LabID.ValueString()), models.UUID(data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get link, got error: %s", err),
		)
		return
	}

	// Preserve explicit slot selections from state; the API does not reliably
	// report slot numbers for all node types.
	if !data.SlotA.IsNull() && !data.SlotA.IsUnknown() {
		link.SrcSlot = int(data.SlotA.ValueInt64())
	}
	if !data.SlotB.IsNull() && !data.SlotB.IsUnknown() {
		link.DstSlot = int(data.SlotB.ValueInt64())
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewLink(ctx, &link, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.LinkAttrType},
			&data,
		)...,
	)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Link READ done")
}

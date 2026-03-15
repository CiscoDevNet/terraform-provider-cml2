package node

import (
	"context"
	"fmt"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *NodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cmlschema.NodeModel

	tflog.Info(ctx, "Resource Node READ")

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node, err := r.cfg.Client().NodeGet(ctx, data.LabID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get node, got error: %s", err),
		)
		return
	}

	// Prefer the format that was used in config/state to avoid drift between
	// `configuration` (single) and `configurations` (named).
	if !r.cfg.UseNamedConfigs() && len(node.Configurations) > 0 {
		if node.Configuration == nil {
			node.Configuration = node.Configurations[0].Content
		}
		node.Configurations = nil
	} else if !data.Configurations.IsNull() {
		node.Configuration = nil
		// keep node.Configurations as-is
	} else if !data.Configuration.IsNull() && len(node.Configurations) > 0 {
		node.Configuration = node.Configurations[0].Content
		node.Configurations = nil
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewNode(ctx, node, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.NodeAttrType},
			&data,
		)...,
	)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Node READ done")
}

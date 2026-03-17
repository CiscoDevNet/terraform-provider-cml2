package node

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

func (r *NodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cmlschema.NodeModel

	tflog.Info(ctx, "Resource Node READ")

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node, err := r.cfg.Client().Node.GetByID(ctx, models.UUID(data.LabID.ValueString()), models.UUID(data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get node, got error: %s", err),
		)
		return
	}

	// Prefer the format that was used in config/state to avoid drift between
	// `configuration` (single) and `configurations` (named).
	//
	// External connector back-compat: if configuration was set in state, keep it
	// in state even if the controller returns the label form. This preserves
	// deprecated device-name configs (e.g. "virbr0") and prevents perpetual diffs.
	if node.NodeDefinition == "external_connector" && !data.Configuration.IsNull() && !data.Configuration.IsUnknown() {
		node.Configuration = data.Configuration.ValueString()
		node.Configurations = nil
	}

	switch {
	case !r.cfg.UseNamedConfigs() && len(node.Configurations) > 0:
		if node.Configuration == nil {
			node.Configuration = node.Configurations[0].Content
		}
		node.Configurations = nil
	case !data.Configurations.IsNull():
		node.Configuration = nil
		// keep node.Configurations as-is
	case !data.Configuration.IsNull() && len(node.Configurations) > 0:
		node.Configuration = node.Configurations[0].Content
		node.Configurations = nil
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewNode(ctx, &node, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.NodeAttrType},
			&data,
		)...,
	)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Node READ done")
}

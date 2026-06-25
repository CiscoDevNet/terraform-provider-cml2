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
	savedGeneration := data.Generation

	node, err := r.cfg.Client().Node.GetByID(ctx, models.UUID(data.LabID.ValueString()), models.UUID(data.ID.ValueString()))
	if err != nil {
		// If the node was deleted outside Terraform, treat it as gone and
		// remove it from the Terraform state. The next plan should recreate it.
		if common.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get node, got error: %s", err),
		)
		return
	}

	// Prefer the format that was used in config/state to avoid drift between
	// `configuration` (single) and `configurations` (named).
	if node.NodeDefinition == "external_connector" {
		if !data.Configuration.IsNull() && !data.Configuration.IsUnknown() {
			node.Configuration = data.Configuration.ValueString()
			node.Configurations = nil
		}
		if !data.Configurations.IsNull() && !data.Configurations.IsUnknown() {
			node.Configuration = nil
			node.Configurations = cmlschema.GetNamedConfigs(ctx, resp.Diagnostics, data.Configurations)
		}
		tflog.Debug(ctx, "extconn read state alignment", map[string]any{
			"saved_configuration": data.Configuration.ValueString(),
			"saved_named_count":   len(data.Configurations.Elements()),
			"api_configuration":   fmt.Sprintf("%v", node.Configuration),
		})
	}
	switch {
	case !r.cfg.UseNamedConfigs() && len(node.Configurations) > 0:
		if node.Configuration == nil {
			node.Configuration = node.Configurations[0].Content
		}
		node.Configurations = nil
	case !data.Configurations.IsNull() && !data.Configurations.IsUnknown():
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

	data.Generation = savedGeneration

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Node READ done")
}

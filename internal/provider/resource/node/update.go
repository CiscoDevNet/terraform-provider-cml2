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
	"github.com/rschmied/gocmlclient/pkg/models"
)

func (r *NodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		stateData, planData cmlschema.NodeModel
		err                 error
	)

	tflog.Info(ctx, "Resource Node UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := &models.Node{
		ID:             models.UUID(planData.ID.ValueString()),
		LabID:          models.UUID(planData.LabID.ValueString()),
		State:          models.NodeState(planData.State.ValueString()),
		NodeDefinition: planData.NodeDefinition.ValueString(),
	}

	if !planData.X.IsNull() {
		node.X = int(planData.X.ValueInt64())
	}
	if !planData.Y.IsNull() {
		node.Y = int(planData.Y.ValueInt64())
	}
	if !planData.HideLinks.IsNull() {
		v := planData.HideLinks.ValueBool()
		node.HideLinks = &v
	}
	if !planData.Priority.IsUnknown() && !planData.Priority.IsNull() {
		v := int(planData.Priority.ValueInt64())
		node.Priority = &v
	}
	if !planData.Label.IsNull() {
		node.Label = planData.Label.ValueString()
	}
	if !planData.Tags.IsNull() && !planData.Tags.IsUnknown() {
		var tag types.String
		tags := []string{}
		for _, elem := range planData.Tags.Elements() {
			resp.Diagnostics.Append(tfsdk.ValueAs(ctx, elem, &tag)...)
			if resp.Diagnostics.HasError() {
				return
			}
			tags = append(tags, tag.ValueString())
		}
		node.Tags = tags
	}

	// these can only be changed when the node is DEFINED_ON_CORE
	if stateData.State.ValueString() == string(models.NodeStateDefined) {
		if !planData.Configuration.IsUnknown() && !planData.Configuration.IsNull() {
			cfgVal := planData.Configuration.ValueString()
			if node.NodeDefinition == "external_connector" {
				normalized, changed, warn, nerr := normalizeExtConnConfig(ctx, r.cfg, cfgVal)
				if nerr != nil {
					resp.Diagnostics.AddError(common.ErrorLabel, nerr.Error())
					return
				}
				if changed {
					resp.Diagnostics.AddWarning("External connector configuration normalized", warn)
					cfgVal = normalized
				}
			}
			node.Configuration = cfgVal
		}
		if !planData.Configurations.IsUnknown() && !planData.Configurations.IsNull() {
			node.Configurations = cmlschema.GetNamedConfigs(ctx, resp.Diagnostics, planData.Configurations)
		}
		if !planData.RAM.IsUnknown() && !planData.RAM.IsNull() {
			v := int(planData.RAM.ValueInt64())
			node.RAM = &v
		}
		if !planData.CPUs.IsUnknown() && !planData.CPUs.IsNull() {
			node.CPUs = int(planData.CPUs.ValueInt64())
		}
		if !planData.CPUlimit.IsUnknown() && !planData.CPUlimit.IsNull() {
			v := int(planData.CPUlimit.ValueInt64())
			node.CPUlimit = &v
		}
		if !planData.BootDiskSize.IsUnknown() && !planData.BootDiskSize.IsNull() {
			v := int(planData.BootDiskSize.ValueInt64())
			node.BootDiskSize = &v
		}
		if !planData.DataVolume.IsUnknown() && !planData.DataVolume.IsNull() {
			v := int(planData.DataVolume.ValueInt64())
			node.DataVolume = &v
		}
		if !planData.ImageDefinition.IsUnknown() && !planData.ImageDefinition.IsNull() {
			v := planData.ImageDefinition.ValueString()
			if v != "" {
				node.ImageDefinition = &v
			}
		}
	}

	// TODO: groups are missing!

	newNode, err := r.cfg.Client().NodeUpdate(ctx, node)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to update node, got error: %s", err),
		)
		return
	}

	// When named configs are disabled provider-side, normalize any server-returned
	// named configs back into the single configuration field to avoid state drift.
	if !r.cfg.UseNamedConfigs() && len(newNode.Configurations) > 0 {
		newNode.Configuration = newNode.Configurations[0].Content
		newNode.Configurations = nil
	}

	// External connector back-compat: if the user provided a device name (e.g.
	// "virbr0"), keep the config value in state to match the user's config.
	// We still sent the normalized label to the API.
	if node.NodeDefinition == "external_connector" && !planData.Configuration.IsUnknown() && !planData.Configuration.IsNull() {
		inCfg := planData.Configuration.ValueString()
		_, changed, _, _ := normalizeExtConnConfig(ctx, r.cfg, inCfg)
		if changed {
			newNode.Configuration = inCfg
			newNode.Configurations = nil
		}
	}

	// External connector: device-name inputs (e.g. virbr0) are normalized to
	// labels (e.g. NAT) during planning for back-compat. Do not hard-fail here.

	// When updating with named configs on, we need to move over the returned
	// named config into the single configuration if it was previously used.
	// tflog.Warn(ctx, "###u", map[string]any{"null": stateData.Configuration.IsNull(), "unknown": stateData.Configuration.IsUnknown(), "len": len(node.Configurations)})
	if !stateData.Configuration.IsUnknown() && len(newNode.Configurations) > 0 {
		newNode.Configuration = newNode.Configurations[0].Content
		newNode.Configurations = nil
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewNode(ctx, newNode, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.NodeAttrType},
			&planData,
		)...,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)

	tflog.Info(ctx, "Resource Node UPDATE done")
}

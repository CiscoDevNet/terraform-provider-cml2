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
			node.Configuration = planData.Configuration.ValueString()
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

	// Work around the fact that updating an external connector can "resolve" the
	// device name (if given, worked in previous versions" with the label... e.g.
	// virbr0 -> NAT, bridge0 -> System Bridge. We want to keep the original
	// value in this case, otherwise we run into inconsistent state!
	if node.NodeDefinition == "external_connector" {
		// working with single string configuration or named configurations?
		if cfg, ok := newNode.Configuration.(string); ok && len(cfg) > 0 {
			nnc := cmlschema.NewConfigValue(cfg)
			if !planData.Configuration.Equal(nnc) {
				resp.Diagnostics.AddError(
					"External connector configuration (single)",
					fmt.Sprintf("Provide proper external connector configuration, not a device name (deprecated)."),
				)
				return
			}
		} else {
			nnc := cmlschema.NewNamedConfigs(ctx, newNode, &resp.Diagnostics)
			if !planData.Configurations.Equal(nnc) {
				oldCfg, _ := node.Configuration.(string)
				newCfg, _ := newNode.Configuration.(string)
				resp.Diagnostics.AddError(
					"External connector configurations (named)",
					fmt.Sprintf("Provide proper external connector configuration, not a device name (deprecated). Was: %q, is: %q", oldCfg, newCfg),
				)
				return
			}
		}
	}

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

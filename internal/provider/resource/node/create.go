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

func (r *NodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		data            cmlschema.NodeModel
		err             error
		extConnStateCfg string
	)

	tflog.Info(ctx, "Resource Node CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// ensure named configs are only used when configured!
	if len(data.Configurations.Elements()) > 0 && !r.cfg.UseNamedConfigs() {
		resp.Diagnostics.AddError(
			"Node config conflict",
			"Provider option \"named_configs\" required to use named configurations!",
		)
		return
	}

	// tflog.Info(ctx, "CFG", map[string]any{"v": fmt.Sprintf("%+v", data.Configuration.IsUnknown())})

	// can't configure both at the same time!
	if len(data.Configurations.Elements()) > 0 && !data.Configuration.IsUnknown() {
		resp.Diagnostics.AddError(
			"Node config conflict",
			"Can't provide both, configuration and configurations",
		)
		return
	}

	node := models.Node{}

	node.LabID = models.UUID(data.LabID.ValueString())

	if !data.Label.IsUnknown() {
		node.Label = data.Label.ValueString()
	}
	if !data.NodeDefinition.IsUnknown() {
		node.NodeDefinition = data.NodeDefinition.ValueString()
	}

	// External connector back-compat: accept device name (e.g. "virbr0") and
	// map to connector label (e.g. "NAT"). Keep the original config value in the
	// plan, but send the normalized label to the API.
	if node.NodeDefinition == "external_connector" && !data.Configuration.IsUnknown() && !data.Configuration.IsNull() {
		inCfg := data.Configuration.ValueString()
		normalized, changed, warn, nerr := normalizeExtConnConfig(ctx, r.cfg, inCfg)
		if nerr != nil {
			resp.Diagnostics.AddError(common.ErrorLabel, nerr.Error())
			return
		}
		if changed {
			resp.Diagnostics.AddWarning("External connector configuration normalized", warn)
			node.Configuration = normalized
			extConnStateCfg = inCfg
		}
	}

	// We always need to create a tag list as the API always returns a list of
	// tags even if none are set... e.g. no tags --> [] (instead of null).
	tags := []string{}
	for _, tag := range data.Tags.Elements() {
		tags = append(tags, tag.(types.String).ValueString())
	}
	node.Tags = tags

	if !data.Configuration.IsUnknown() && !data.Configuration.IsNull() {
		// Only set configuration from plan if we did not already normalize it for
		// external connectors.
		if node.NodeDefinition != "external_connector" || extConnStateCfg == "" {
			node.Configuration = data.Configuration.ValueString()
		}
	}

	if !data.Configurations.IsUnknown() && !data.Configurations.IsNull() {
		node.Configurations = cmlschema.GetNamedConfigs(ctx, resp.Diagnostics, data.Configurations)
	}

	if !data.X.IsUnknown() {
		node.X = int(data.X.ValueInt64())
	}
	if !data.Y.IsUnknown() {
		node.Y = int(data.Y.ValueInt64())
	}
	if !data.HideLinks.IsUnknown() {
		v := data.HideLinks.ValueBool()
		node.HideLinks = &v
	}
	if !data.Priority.IsUnknown() && !data.Priority.IsNull() {
		v := int(data.Priority.ValueInt64())
		node.Priority = &v
	}
	if !data.RAM.IsUnknown() && !data.RAM.IsNull() {
		v := int(data.RAM.ValueInt64())
		node.RAM = &v
	}
	if !data.CPUs.IsUnknown() && !data.CPUs.IsNull() {
		node.CPUs = int(data.CPUs.ValueInt64())
	}
	if !data.CPUlimit.IsUnknown() && !data.CPUlimit.IsNull() {
		v := int(data.CPUlimit.ValueInt64())
		node.CPUlimit = &v
	}
	if !data.BootDiskSize.IsUnknown() && !data.BootDiskSize.IsNull() {
		v := int(data.BootDiskSize.ValueInt64())
		node.BootDiskSize = &v
	}
	if !data.DataVolume.IsUnknown() && !data.DataVolume.IsNull() {
		v := int(data.DataVolume.ValueInt64())
		node.DataVolume = &v
	}
	if !data.ImageDefinition.IsUnknown() && !data.ImageDefinition.IsNull() {
		v := data.ImageDefinition.ValueString()
		if v != "" {
			node.ImageDefinition = &v
		}
	}

	// can't set a configuration for an unmanaged switch
	// tflog.Warn(ctx, "##UMS", map[string]any{"has_config": data.HasConfig(), "unknown": data.Configuration.IsUnknown(), "len": len(node.Configurations)})
	if node.NodeDefinition == "unmanaged_switch" && data.HasConfig() {
		resp.Diagnostics.AddError(
			"Unmanaged switch configuration",
			"Can't provide UMS configuration",
		)
		return
	}

	// tflog.Info(ctx, "NODE", map[string]any{"v": fmt.Sprintf("%+v", node)})

	newNode, err := r.cfg.Client().Node.Create(ctx, node)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to create node, got error: %s", err),
		)
		return
	}

	// When named configs are disabled provider-side, normalize any server-returned
	// named configs back into the single configuration field to avoid state drift.
	if !r.cfg.UseNamedConfigs() && len(newNode.Configurations) > 0 {
		if newNode.Configuration == nil {
			newNode.Configuration = newNode.Configurations[0].Content
		}
		newNode.Configurations = nil
	}

	// If we accepted a deprecated device name (e.g. "virbr0"), keep the device
	// name in state to match the user's config value and avoid Terraform drift.
	if node.NodeDefinition == "external_connector" && extConnStateCfg != "" {
		newNode.Configuration = extConnStateCfg
		newNode.Configurations = nil
	}

	// WAS UNKNOWN??
	// tflog.Warn(ctx, "###2", map[string]any{"null": data.Configuration.IsNull(), "unknown": data.Configuration.IsUnknown(), "len": len(node.Configurations)})
	if !data.Configuration.IsUnknown() && len(newNode.Configurations) > 0 {
		newNode.Configuration = newNode.Configurations[0].Content
		newNode.Configurations = nil
	}

	// External connector: device-name inputs (e.g. virbr0) are normalized to
	// labels (e.g. NAT) during planning for back-compat. Do not hard-fail here.

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewNode(ctx, &newNode, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.NodeAttrType},
			&data,
		)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Node CREATE done")
}

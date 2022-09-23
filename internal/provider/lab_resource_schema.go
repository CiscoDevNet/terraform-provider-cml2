package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (t *LabResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {

	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CML Lab resource",

		Blocks: map[string]tfsdk.Block{
			"timeouts": {
				Attributes: map[string]tfsdk.Attribute{
					"create": {
						Optional:    true,
						Description: "create timeout",
						Type:        types.StringType,
						Validators: []tfsdk.AttributeValidator{
							durationValidator{},
						},
					},
					"update": {
						Optional:    true,
						Description: "update timeout",
						Type:        types.StringType,
						Validators: []tfsdk.AttributeValidator{
							durationValidator{},
						},
					},
					"delete": {
						Optional:    true,
						Description: "delete timeout (currently unused)",
						Type:        types.StringType,
						Validators: []tfsdk.AttributeValidator{
							durationValidator{},
						},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSingle,
			},
		},

		Attributes: map[string]tfsdk.Attribute{
			// topology is marked as sensitive mostly b/c lengthy topology
			// YAML clutters the output.
			"topology": {
				MarkdownDescription: "topology to start",
				Required:            true,
				Type:                types.StringType,
				Sensitive:           true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"wait": {
				MarkdownDescription: "wait until topology is BOOTED if true",
				Optional:            true,
				Type:                types.BoolType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "CML lab identifier, a UUID",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"booted": {
				Computed:            true,
				MarkdownDescription: "All nodes in the lab have booted",
				Type:                types.BoolType,
			},
			"state": {
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "CML lab state, one of `DEFINED_ON_CORE`, `STARTED` or `STOPPED`",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
				Validators: []tfsdk.AttributeValidator{
					labStateValidator{},
				},
			},
			"nodes": {
				Description: "List of nodes and their interfaces with IP addresses",
				Computed:    true,
				Attributes: tfsdk.MapNestedAttributes(
					nodeSchema(),
				),
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"configs": {
				Description: "Map of node configurations to store into nodes, the key is the label of the node",
				Optional:    true,
				Type: types.MapType{
					ElemType: types.StringType,
				},
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"stages": {
				MarkdownDescription: "Ordered list of tags, controls node launch",
				Optional:            true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
		},
	}, nil
}

func interfaceSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			MarkdownDescription: "Interface ID (UUID)",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"mac_address": {
			MarkdownDescription: "MAC address",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"is_connected": {
			MarkdownDescription: "connection status",
			Type:                types.BoolType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"ip4": {
			MarkdownDescription: "IPv4 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"ip6": {
			MarkdownDescription: "IPv6 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
	}
}

func nodeSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			MarkdownDescription: "Node ID (UUID)",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"nodedefinition": {
			MarkdownDescription: "Node Definition",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"interfaces": {
			MarkdownDescription: "interfaces on the node",
			Computed:            true,
			// Sensitive:           false,
			Attributes: tfsdk.ListNestedAttributes(
				interfaceSchema(),
			),
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"tags": {
			MarkdownDescription: "Tags of the node",
			Computed:            true,
			// Sensitive:           false,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"configuration": {
			MarkdownDescription: "device configuration",
			Type:                types.StringType,
			Computed:            true,
			// Sensitive:           true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
	}
}

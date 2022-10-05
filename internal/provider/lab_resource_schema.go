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
		Description: "A CML lab resource represents a complete CML lab lifecyle, including configuration injection and staged node launches.  Resulting state also includes IP addresses of nodes which have external connectivity.",

		// Attributes are preferred over Blocks. Blocks should typically be used
		// for configuration compatibility with previously existing schemas from
		// an older Terraform Plugin SDK. Efforts should be made to convert from
		// Blocks to Attributes as a breaking change for practitioners.

		Attributes: map[string]tfsdk.Attribute{
			// topology is marked as sensitive mostly b/c lengthy topology
			// YAML clutters the output.
			"topology": {
				Description: "topology to start",
				Required:    true,
				Type:        types.StringType,
				Sensitive:   true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"wait": {
				Description:         "wait until topology is BOOTED if true",
				MarkdownDescription: "wait until topology is `BOOTED` if true",
				Optional:            true,
				Type:                types.BoolType,
			},
			"id": {
				Computed:    true,
				Description: "CML lab identifier, a UUID",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"booted": {
				Computed:    true,
				Description: "all nodes in the lab have booted",
				Type:        types.BoolType,
			},
			"state": {
				Computed:            true,
				Optional:            true,
				Description:         "CML lab state, one of DEFINED_ON_CORE, STARTED or STOPPED",
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
				Description: "ist of nodes and their interfaces with IP addresses",
				Computed:    true,
				Attributes: tfsdk.MapNestedAttributes(
					nodeSchema(),
				),
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"configs": {
				Description: "map of node configurations to store into nodes, the key is the label of the node",
				Optional:    true,
				Type: types.MapType{
					ElemType: types.StringType,
				},
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"timeouts": {
				Description:         "timeouts for operations, given as a parsable string as in 60m or 2h",
				MarkdownDescription: "timeouts for operations, given as a parsable string as in `60m` or `2h`",
				Optional:            true,
				Attributes: tfsdk.SingleNestedAttributes(
					map[string]tfsdk.Attribute{
						"create": {
							Required:    true,
							Description: "create timeout",
							Type:        types.StringType,
							Validators: []tfsdk.AttributeValidator{
								durationValidator{},
							},
						},
						"update": {
							Required:    true,
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
				),
			},
			"staging": {
				Description: "defines in what sequence nodes are launched",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(
					map[string]tfsdk.Attribute{
						"stages": {
							Description: "ordered list of node tags, controls node launch. Nodes currently not launched will be launched in the stage with the matching tag. Tags must match exactly.",
							Required:    true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
							PlanModifiers: []tfsdk.AttributePlanModifier{
								resource.RequiresReplace(),
							},
						},
						"start_remaining": {
							Optional:    true,
							Description: "if true (which is the default) then all nodes which are not matched by the stages list and which are still unstarted at the end of the stages list will be started",
							Type:        types.BoolType,
						},
					},
				),
			},
		},
	}, nil
}

func interfaceSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Description: "interface ID (UUID)",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"label": {
			Description: "label",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"mac_address": {
			Description: "MAC address",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"is_connected": {
			Description: "connection status",
			Type:        types.BoolType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"state": {
			Description: "interface state (UP / DOWN)",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"ip4": {
			Description: "IPv4 address list",
			Computed:    true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"ip6": {
			Description: "IPv6 address list",
			Computed:    true,
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
			Description: "node ID (UUID)",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"label": {
			Description: "label",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"state": {
			Description: "state",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"nodedefinition": {
			Description: "node definition / type",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"interfaces": {
			Description: "list of interfaces on the node",
			Computed:    true,
			// Sensitive:           false,
			Attributes: tfsdk.ListNestedAttributes(
				interfaceSchema(),
			),
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"tags": {
			Description: "tags of the node",
			Computed:    true,
			// Sensitive:           false,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"configuration": {
			Description: "node configuration",
			Type:        types.StringType,
			Computed:    true,
			// Sensitive:           true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
	}
}

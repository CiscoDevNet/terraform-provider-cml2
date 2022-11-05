package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/rschmied/terraform-provider-cml2/internal/validator"
)

type LabLifecycleModel struct {
	ID       types.String `tfsdk:"id"`
	LabID    types.String `tfsdk:"lab_id"`
	Topology types.String `tfsdk:"topology"`
	Wait     types.Bool   `tfsdk:"wait"`
	State    types.String `tfsdk:"state"`
	Booted   types.Bool   `tfsdk:"booted"`
	Nodes    types.Map    `tfsdk:"nodes"`
	Configs  types.Map    `tfsdk:"configs"`
	Staging  types.Object `tfsdk:"staging"`
	Timeouts types.Object `tfsdk:"timeouts"`
	Elements types.List   `tfsdk:"elements"`
}

func Lifecycle() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Computed:    true,
			Description: "Resource identifier, a UUID.",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		"lab_id": {
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Lab identifier, a UUID. If set, `elements` must be configured as well.",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.RequiresReplace(),
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		// topology is marked as sensitive mostly b/c lengthy topology
		// YAML clutters the output.
		"topology": {
			MarkdownDescription: "The topology to start, must be valid YAML. Can't be configured if the lab `id` is configured.",
			Optional:            true,
			Type:                types.StringType,
			Sensitive:           true,
			PlanModifiers: []tfsdk.AttributePlanModifier{
				resource.RequiresReplace(),
			},
		},
		"wait": {
			MarkdownDescription: "If set to `true` then wait until the lab has completely `BOOTED`.",
			Optional:            true,
			Type:                types.BoolType,
		},
		"booted": {
			Computed:            true,
			MarkdownDescription: "Set to `true` when all nodes in the lab have booted.",
			Type:                types.BoolType,
		},
		"state": {
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "Lab state, one of `DEFINED_ON_CORE`, `STARTED` or `STOPPED`.",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
			Validators: []tfsdk.AttributeValidator{
				validator.LabState{},
			},
		},
		"nodes": {
			Description: "List of nodes and their interfaces with IP addresses.",
			Computed:    true,
			Attributes: tfsdk.MapNestedAttributes(
				Node(),
			),
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"configs": {
			Description: "Map of node configurations to store into nodes, the key is the label of the node, the value is the node configuration.",
			Optional:    true,
			Type: types.MapType{
				ElemType: types.StringType,
			},
			PlanModifiers: []tfsdk.AttributePlanModifier{
				resource.RequiresReplace(),
			},
		},
		"timeouts": {
			MarkdownDescription: "Timeouts for operations, given as a parsable string as in `60m` or `2h`.",
			Optional:            true,
			Attributes: tfsdk.SingleNestedAttributes(
				map[string]tfsdk.Attribute{
					"create": {
						Required:    true,
						Description: "Create timeout.",
						Type:        types.StringType,
						Validators: []tfsdk.AttributeValidator{
							validator.Duration{},
						},
					},
					"update": {
						Required:    true,
						Description: "Update timeout.",
						Type:        types.StringType,
						Validators: []tfsdk.AttributeValidator{
							validator.Duration{},
						},
					},
					"delete": {
						Optional:    true,
						Description: "Delete timeout (currently unused).",
						Type:        types.StringType,
						Validators: []tfsdk.AttributeValidator{
							validator.Duration{},
						},
					},
				},
			),
		},
		"staging": {
			Description: "Defines in what sequence nodes are launched.",
			Optional:    true,
			Attributes: tfsdk.SingleNestedAttributes(
				map[string]tfsdk.Attribute{
					"stages": {
						Description: "Ordered list of node tags, controls node launch. Nodes currently not launched will be launched in the stage with the matching tag. Tags must match exactly.",
						Required:    true,
						Type: types.ListType{
							ElemType: types.StringType,
						},
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
						},
					},
					"start_remaining": {
						Optional:            true,
						MarkdownDescription: "If set to `true` (which is the default) then all nodes which are not matched by the stages list and which are still unstarted after running all stages will be started.",
						Type:                types.BoolType,
					},
				},
			),
		},
		"elements": {
			Description: "List of node and link IDs the lab consists of.  Works only when a (lab) ID is provided and no topology is configured.",
			Optional:    true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: []tfsdk.AttributePlanModifier{
				resource.RequiresReplace(),
			},
		},
	}
}

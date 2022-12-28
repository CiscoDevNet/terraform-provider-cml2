package cmlschema

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlvalidator"
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

func Lifecycle() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Resource identifier, a UUID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"lab_id": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Lab identifier, a UUID. If set, `elements` must be configured as well.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		// topology is marked as sensitive mostly b/c lengthy topology
		// YAML clutters the output.
		"topology": schema.StringAttribute{
			MarkdownDescription: "The topology to start, must be valid YAML. Can't be configured if the lab `id` is configured.",
			Optional:            true,
			Sensitive:           true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"wait": schema.BoolAttribute{
			MarkdownDescription: "If set to `true` then wait until the lab has completely `BOOTED`.",
			Optional:            true,
		},
		"booted": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Set to `true` when all nodes in the lab have booted.",
		},
		"state": schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "Lab state, one of `DEFINED_ON_CORE`, `STARTED` or `STOPPED`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				cmlvalidator.LabState{},
			},
		},
		"nodes": schema.MapNestedAttribute{
			Description: "List of nodes and their interfaces with IP addresses.",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: Node(),
			},
			PlanModifiers: []planmodifier.Map{
				mapplanmodifier.UseStateForUnknown(),
			},
		},
		"configs": schema.MapAttribute{
			Description: "Map of node configurations to store into nodes, the key is the label of the node, the value is the node configuration.",
			Optional:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.Map{
				mapplanmodifier.RequiresReplace(),
			},
		},
		"timeouts": schema.SingleNestedAttribute{
			MarkdownDescription: "Timeouts for operations, given as a parsable string as in `60m` or `2h`.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"create": schema.StringAttribute{
					Required:    true,
					Description: "Create timeout.",
					Validators: []validator.String{
						cmlvalidator.Duration{},
					},
				},
				"update": schema.StringAttribute{
					Required:    true,
					Description: "Update timeout.",
					Validators: []validator.String{
						cmlvalidator.Duration{},
					},
				},
				"delete": schema.StringAttribute{
					Optional:    true,
					Description: "Delete timeout (currently unused).",
					Validators: []validator.String{
						cmlvalidator.Duration{},
					},
				},
			},
		},
		"staging": schema.SingleNestedAttribute{
			Description: "Defines in what sequence nodes are launched.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"stages": schema.ListAttribute{
					Description: "Ordered list of node tags, controls node launch. Nodes currently not launched will be launched in the stage with the matching tag. Tags must match exactly.",
					Required:    true,
					ElementType: types.StringType,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
				"start_remaining": schema.BoolAttribute{
					Optional:            true,
					MarkdownDescription: "If set to `true` (which is the default) then all nodes which are not matched by the stages list and which are still unstarted after running all stages will be started.",
				},
			},
		},
		"elements": schema.ListAttribute{
			Description: "List of node and link IDs the lab consists of. Works only when a (lab) ID is provided and no topology is configured.",
			Optional:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
	}
}

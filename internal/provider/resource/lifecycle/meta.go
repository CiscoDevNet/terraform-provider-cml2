package lifecycle

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

const CML2ErrorLabel = "CML2 Provider Error"

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &LabLifecycleResource{}
var _ resource.ResourceWithImportState = &LabLifecycleResource{}
var _ resource.ResourceWithValidateConfig = &LabLifecycleResource{}

type LabLifecycleResource struct {
	cfg *common.ProviderConfig
}

type labLifecycleStaging struct {
	Stages         types.List `tfsdk:"stages"`
	StartRemaining types.Bool `tfsdk:"start_remaining"`
}

type labLifecycleTimeouts struct {
	Create types.String `tfsdk:"create"`
	Update types.String `tfsdk:"update"`
	Delete types.String `tfsdk:"delete"`
}

type startData struct {
	wait     bool
	lab      *cmlclient.Lab
	staging  *labLifecycleStaging
	timeouts *labLifecycleTimeouts
}

func (r *LabLifecycleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// This description is used by the documentation generator and the language
	// server.
	resp.Schema.Description = "A lifecycle resource represents a complete CML lab lifecyle, including configuration injection and staged node launches.  Resulting state also includes IP addresses of nodes which have external connectivity. This is a synthetic resource which \"glues\" other actual resources like labs, nodes and links together."
	resp.Schema.Attributes = cmlschema.Lifecycle()
	resp.Diagnostics = nil
}

func (r *LabLifecycleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cfg = common.ResourceConfigure(ctx, req, resp)
}

func (r *LabLifecycleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lifecycle"
}

func (r *LabLifecycleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data cmlschema.LabLifecycleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.LabID.IsNull() && !data.Topology.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("topology"),
			"Conflicting configuration",
			"Can't set \"LabID\" and \"topology\" at the same time.",
		)
		return
	}
	// id and elements are mutually exclusive with topology
	if !data.LabID.IsNull() && data.Elements.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("elements"),
			"Required configuration",
			"When \"LabID\" is set, \"elements\" is a required attribue.",
		)
		return
	}

	// If staging is not configured, return without warning.
	// (I think it never can be unknown as it's configuration data)
	if data.Staging.IsNull() || data.Staging.IsUnknown() {
		return
	}

	// If wait is set (true), return without warning
	// if it is null, then the default is "true" (e.g. wait)
	if data.Wait.IsNull() || data.Wait.ValueBool() {
		return
	}

	resp.Diagnostics.AddAttributeWarning(
		path.Root("staging"),
		"Conflicting configuration",
		"\"wait\" is set to false while staging is configured. "+
			"The resource may return unexpected results.",
	)
}

func NewResource() resource.Resource {
	return &LabLifecycleResource{}
}

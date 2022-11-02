package lifecycle

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/common"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
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

func (t *LabLifecycleResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {

	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "A lifecycle resource represents a complete CML lab lifecyle, including configuration injection and staged node launches.  Resulting state also includes IP addresses of nodes which have external connectivity.",

		// Attributes are preferred over Blocks. Blocks should typically be used
		// for configuration compatibility with previously existing schemas from
		// an older Terraform Plugin SDK. Efforts should be made to convert from
		// Blocks to Attributes as a breaking change for practitioners.

		Attributes: schema.Lifecycle(),
	}, nil
}

func (r *LabLifecycleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cfg = common.ResourceConfigure(ctx, req, resp)
}

func (r *LabLifecycleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lifecycle"
}

func (r *LabLifecycleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data schema.LabLifecycleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// id and elements are mutually exclusive with topology
	if !data.ID.Null && data.Elements.Null {
		resp.Diagnostics.AddAttributeError(
			path.Root("elements"),
			"Required configuration",
			"When \"ID\" is set, \"elements\" is a required attribue.",
		)
		return
	}
	if !data.ID.Null && !data.Topology.Null {
		resp.Diagnostics.AddAttributeError(
			path.Root("topology"),
			"Conflicting configuration",
			"Can't set \"ID\" and \"topology\" at the same time.",
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
	if data.Wait.IsNull() || data.Wait.Value {
		return
	}

	resp.Diagnostics.AddAttributeWarning(
		path.Root("staging"),
		"Conflicting configuration",
		"Expected \"wait\" to be true with when staging is configured. "+
			"The resource may return unexpected results.",
	)
}

func NewResource() resource.Resource {
	return &LabLifecycleResource{}
}

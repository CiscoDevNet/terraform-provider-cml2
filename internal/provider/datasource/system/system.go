package system

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlvalidator"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SystemDataSource{}

type SystemDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Version      types.String `tfsdk:"version"`
	Ready        types.Bool   `tfsdk:"ready"`
	Timeout      types.String `tfsdk:"timeout"`
	IgnoreErrors types.Bool   `tfsdk:"ignore_errors"`
}

func NewDataSource() datasource.DataSource {
	return &SystemDataSource{}
}

// SystemDataSource defines the data source implementation.
type SystemDataSource struct {
	cfg *common.ProviderConfig
}

func (d *SystemDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system"
}

func (d *SystemDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cfg = common.DatasourceConfigure(ctx, req, resp)
}

func (d *SystemDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "A UUID. The presence of the ID attribute is mandated by the framework. The attribute is a random UUID and has no actual significance.",
			Computed:    true,
		},
		"version": schema.StringAttribute{
			Description: "The system software version.",
			Computed:    true,
		},
		"ready": schema.BoolAttribute{
			Description: "Shows if the CML system API is ready.",
			Optional:    true,
		},
		"timeout": schema.StringAttribute{
			MarkdownDescription: "Wait timeout, like `5m`, defaults to 0.",
			Validators: []validator.String{
				cmlvalidator.Duration{},
			},
			Optional: true,
		},
		"ignore_errors": schema.BoolAttribute{
			MarkdownDescription: "If set to `true`, then errors will be ignored during the ready check. This can help when using proxies which might return intermediate errors especially during the initial phase where gateway timeouts or proxy errors might be returned because of initial connectivity issues towards the CML2 instance. Will default to `false`.",
			Optional:            true,
		},
	}

	resp.Schema.MarkdownDescription = "A data source that retrieves system state information from the controller. If a `timeout` is set then this will only return when the system responds."
	resp.Diagnostics = nil
}

func (d *SystemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SystemDataSourceModel

	tflog.Info(ctx, "Datasource System READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tov time.Duration

	timeout := "0s"
	if !data.Timeout.IsNull() {
		timeout = data.Timeout.ValueString()
	}

	ignoreErrors := false
	if !data.IgnoreErrors.IsNull() {
		ignoreErrors = data.IgnoreErrors.ValueBool()
	}

	tov, err := time.ParseDuration(timeout)
	if err != nil {
		panic("can't parse timeout -- should be validated")
	}

	snoozeFor := 5 * time.Second
	endTime := time.Now().Add(tov)
	waited := time.Duration(0)

	for {
		err = d.cfg.Client().Ready(ctx)
		if err == nil {
			break
		}
		if !(errors.Is(err, cmlclient.ErrSystemNotReady) || ignoreErrors) {
			resp.Diagnostics.AddError("CML client error", err.Error())
			return
		}

		// if no timeout was specified, break immediately
		if time.Now().After(endTime) {
			break
		}

		select {
		case <-time.After(snoozeFor):
		case <-ctx.Done():
			return
		}
		if time.Now().After(endTime) {
			resp.Diagnostics.AddError(
				common.ErrorLabel,
				fmt.Sprintf("ran into timeout (max %s)", timeout),
			)
			return
		}
		waited++
		tflog.Info(
			ctx, "wait for system ready",
			map[string]any{"seconds": waited * snoozeFor},
		)
	}

	if data.ID.IsNull() {
		data.ID = types.StringValue(uuid.NewString())
	}
	if err != nil {
		resp.Diagnostics.AddWarning("system ready", fmt.Sprintf("err %s", err))
	}
	data.Ready = types.BoolValue(err == nil)
	data.Version = types.StringValue(d.cfg.Client().Version())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Datasource System READ: done")
}

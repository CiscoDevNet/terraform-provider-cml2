// Package system implements the CML2 system datasource.
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

	cmlerrors "github.com/rschmied/gocmlclient/pkg/errors"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlvalidator"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SystemDataSource{}

// SystemDataSourceModel describes the data source data model.
type SystemDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Version      types.String `tfsdk:"version"`
	Ready        types.Bool   `tfsdk:"ready"`
	Timeout      types.String `tfsdk:"timeout"`
	IgnoreErrors types.Bool   `tfsdk:"ignore_errors"`
}

// NewDataSource returns a new system data source.
func NewDataSource() datasource.DataSource {
	return &SystemDataSource{}
}

// SystemDataSource defines the data source implementation.
type SystemDataSource struct {
	cfg *common.ProviderConfig
}

// Metadata sets the data source type name.
func (d *SystemDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system"
}

// Configure stores provider configuration for the data source.
func (d *SystemDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cfg = common.DatasourceConfigure(ctx, req, resp)
}

// Schema defines the schema for the data source.
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
		resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("invalid timeout %q: %s", timeout, err))
		return
	}

	snoozeFor := 5 * time.Second
	endTime := time.Now().Add(tov)
	attempts := 0

	for {
		err = d.cfg.Client().System.Ready(ctx)
		if err == nil {
			break
		}
		if !errors.Is(err, cmlerrors.ErrSystemNotReady) && !ignoreErrors {
			resp.Diagnostics.AddError("CML client error", err.Error())
			return
		}

		// if no timeout was specified, break immediately after the first check
		if tov <= 0 {
			break
		}

		if time.Now().After(endTime) {
			if !ignoreErrors {
				resp.Diagnostics.AddError(
					common.ErrorLabel,
					fmt.Sprintf("ran into timeout (max %s)", timeout),
				)
				return
			}
			break
		}

		waitFor := snoozeFor
		if remaining := time.Until(endTime); remaining < waitFor {
			waitFor = remaining
		}

		select {
		case <-time.After(waitFor):
		case <-ctx.Done():
			return
		}
		if time.Now().After(endTime) {
			if ignoreErrors {
				break
			}
			resp.Diagnostics.AddError(
				common.ErrorLabel,
				fmt.Sprintf("ran into timeout (max %s)", timeout),
			)
			return
		}
		attempts++
		tflog.Info(
			ctx, "wait for system ready",
			map[string]any{"seconds": time.Duration(attempts) * snoozeFor},
		)
	}

	if data.ID.IsNull() {
		data.ID = types.StringValue(uuid.NewString())
	}
	if err != nil {
		resp.Diagnostics.AddWarning("system ready", fmt.Sprintf("err %s", err))
	}
	data.Ready = types.BoolValue(err == nil)
	data.Version = types.StringValue(d.cfg.Client().System.Version())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Datasource System READ: done")
}

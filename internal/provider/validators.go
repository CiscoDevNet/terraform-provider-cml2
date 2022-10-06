package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/terraform-provider-cml2/m/v2/pkg/cmlclient"
)

type labStateValidator struct{}

func (v labStateValidator) Description(ctx context.Context) string {
	return "valid states are DEFINED_ON_CORE, STOPPED and STARTED"
}

// MarkdownDescription returns a markdown formatted description of the
// validator's behavior, suitable for a practitioner to understand its impact.
func (v labStateValidator) MarkdownDescription(ctx context.Context) string {
	return "valid states are `DEFINED_ON_CORE`, `STOPPED` and `STARTED`"
}

// Validate runs the main validation logic of the validator, reading
// configuration data out of `req` and updating `resp` with diagnostics.
func (v labStateValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var labState types.String

	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.AttributeConfig, &labState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if labState.Unknown || labState.Null {
		return
	}

	if labState.Value != cmlclient.LabStateDefined &&
		labState.Value != cmlclient.LabStateStopped &&
		labState.Value != cmlclient.LabStateStarted {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid lab state",
			"valid states are DEFINED_ON_CORE, STOPPED and STARTED.",
		)
		return
	}
}

type durationValidator struct{}

func (v durationValidator) Description(ctx context.Context) string {
	return "a duration given as a parsable string as in 60m or 2h"
}

// MarkdownDescription returns a markdown formatted description of the
// validator's behavior, suitable for a practitioner to understand its impact.
func (v durationValidator) MarkdownDescription(ctx context.Context) string {
	return "a duration given as a parsable string as in `60m` or `2h`"
}

// Validate runs the main validation logic of the validator, reading
// configuration data out of `req` and updating `resp` with diagnostics.
func (v durationValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {

	var duration types.String
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.AttributeConfig, &duration)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if duration.Unknown || duration.Null {
		return
	}

	_, err := time.ParseDuration(duration.Value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid duration",
			err.Error(),
		)
		return
	}
}

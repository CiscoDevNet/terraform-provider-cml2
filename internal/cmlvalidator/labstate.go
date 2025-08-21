// Package cmlvalidator provides functions and types which validate CML related
// types
package cmlvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

var _ validator.String = LabState{}

type LabState struct{}

func (v LabState) Description(ctx context.Context) string {
	return "valid states are DEFINED_ON_CORE, STOPPED and STARTED"
}

// MarkdownDescription returns a markdown formatted description of the
// validator's behavior, suitable for a practitioner to understand its impact.
func (v LabState) MarkdownDescription(ctx context.Context) string {
	return "valid states are `DEFINED_ON_CORE`, `STOPPED` and `STARTED`"
}

// ValidateString runs the main validation logic of the validator, reading
// configuration data out of `req` and updating `resp` with diagnostics.
func (v LabState) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var labState types.String

	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.ConfigValue, &labState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if labState.IsUnknown() || labState.IsNull() {
		return
	}

	if labState.ValueString() != cmlclient.LabStateDefined &&
		labState.ValueString() != cmlclient.LabStateStopped &&
		labState.ValueString() != cmlclient.LabStateStarted {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid lab state",
			"valid states are DEFINED_ON_CORE, STOPPED and STARTED.",
		)
		return
	}
}

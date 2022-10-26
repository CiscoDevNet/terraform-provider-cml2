package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

var _ tfsdk.AttributeValidator = LabState{}

type LabState struct{}

func (v LabState) Description(ctx context.Context) string {
	return "valid states are DEFINED_ON_CORE, STOPPED and STARTED"
}

// MarkdownDescription returns a markdown formatted description of the
// validator's behavior, suitable for a practitioner to understand its impact.
func (v LabState) MarkdownDescription(ctx context.Context) string {
	return "valid states are `DEFINED_ON_CORE`, `STOPPED` and `STARTED`"
}

// Validate runs the main validation logic of the validator, reading
// configuration data out of `req` and updating `resp` with diagnostics.
func (v LabState) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
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

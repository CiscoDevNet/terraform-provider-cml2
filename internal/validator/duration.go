package validator

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ tfsdk.AttributeValidator = Duration{}

type Duration struct{}

func (v Duration) Description(ctx context.Context) string {
	return "a duration given as a parsable string as in 60m or 2h"
}

// MarkdownDescription returns a markdown formatted description of the
// validator's behavior, suitable for a practitioner to understand its impact.
func (v Duration) MarkdownDescription(ctx context.Context) string {
	return "a duration given as a parsable string as in `60m` or `2h`"
}

// Validate runs the main validation logic of the validator, reading
// configuration data out of `req` and updating `resp` with diagnostics.
func (v Duration) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {

	var duration types.String
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.AttributeConfig, &duration)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if duration.IsUnknown() || duration.IsNull() {
		return
	}

	_, err := time.ParseDuration(duration.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid duration",
			err.Error(),
		)
		return
	}
}

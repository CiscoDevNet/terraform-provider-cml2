package cmlvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = GroupPermission{}

type GroupPermission struct{}

func (v GroupPermission) Description(ctx context.Context) string {
	return "valid states are \"read_write\" and \"read_only\""
}

// MarkdownDescription returns a markdown formatted description of the
// validator's behavior, suitable for a practitioner to understand its impact.
func (v GroupPermission) MarkdownDescription(ctx context.Context) string {
	return "valid states are `read_write` and `read_only`"
}

// ValidateString runs the main validation logic of the validator, reading
// configuration data out of `req` and updating `resp` with diagnostics.
func (v GroupPermission) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var permission types.String

	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.ConfigValue, &permission)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if permission.IsUnknown() || permission.IsNull() {
		return
	}

	if permission.ValueString() != "read_write" &&
		permission.ValueString() != "read_only" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid group permission value",
			"valid states are read_write and read_only.",
		)
		return
	}
}

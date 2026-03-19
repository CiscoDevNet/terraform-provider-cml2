package user

import "github.com/hashicorp/terraform-plugin-framework/types"

func resourcePoolTemplateChanged(plan, state types.String) bool {
	if plan.IsUnknown() || state.IsUnknown() {
		return false
	}
	if plan.IsNull() && state.IsNull() {
		return false
	}
	if plan.IsNull() != state.IsNull() {
		return true
	}
	return plan.ValueString() != state.ValueString()
}

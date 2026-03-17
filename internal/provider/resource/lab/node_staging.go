package lab

import (
	"context"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func expandNodeStaging(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *models.NodeStaging {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	var ns cmlschema.LabNodeStagingModel
	diags.Append(tfsdk.ValueAs(ctx, obj, &ns)...)
	if diags.HasError() {
		return nil
	}

	startRemaining := true
	if !ns.StartRemaining.IsNull() {
		startRemaining = ns.StartRemaining.ValueBool()
	}
	abortOnFailure := false
	if !ns.AbortOnFailure.IsNull() {
		abortOnFailure = ns.AbortOnFailure.ValueBool()
	}

	return &models.NodeStaging{
		Enabled:        ns.Enabled.ValueBool(),
		StartRemaining: startRemaining,
		AbortOnFailure: abortOnFailure,
	}
}

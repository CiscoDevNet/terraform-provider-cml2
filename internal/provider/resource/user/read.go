package user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cmlschema.UserModel

	tflog.Info(ctx, "Resource User READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.cfg.Client().User.GetByID(ctx, models.UUID(data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get user, got error: %s", err),
		)
		return
	}

	// need to preserve "write once" values
	user.Password = data.Password.ValueString()
	value := cmlschema.NewUser(ctx, &user, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	newModel := cmlschema.UserModel{}
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, value, &newModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve config-only input in state.
	newModel.ResourcePoolTemplate = data.ResourcePoolTemplate

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newModel)...)

	tflog.Info(ctx, "Resource User READ: done")
}

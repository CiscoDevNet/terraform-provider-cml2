package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

func (r *LabResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cmlschema.LabModel

	tflog.Info(ctx, "Resource Lab READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		lab *cmlclient.Lab
		err error
	)

	lab, err = r.cfg.Client().LabGet(ctx, data.ID.ValueString(), false)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	// Save data into Terraform state
	value := cmlschema.NewLab(ctx, lab, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &value)...)

	tflog.Info(ctx, "Resource Lab READ done")
}

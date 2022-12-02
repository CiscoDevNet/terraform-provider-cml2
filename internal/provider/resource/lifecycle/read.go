package lifecycle

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

func (r *LabLifecycleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *cmlschema.LabLifecycleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// tflog.Info(ctx, "state:", map[string]any{"data": data})

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Read: errors!")
		return
	}

	tflog.Info(ctx, "Read: start")

	lab, err := r.cfg.Client().LabGet(ctx, data.LabID.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to fetch lab, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Read: lab state: %s", lab.State))

	data.LabID = types.StringValue(lab.ID)
	data.State = types.StringValue(lab.State)
	data.Nodes = r.populateNodes(ctx, lab, &resp.Diagnostics)
	data.Booted = types.BoolValue(lab.Booted())

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Read: errors!")
		return
	}
	tflog.Info(ctx, "Read: done")
}

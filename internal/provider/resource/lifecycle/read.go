package lifecycle

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r *LabLifecycleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *schema.LabLifecycleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// tflog.Info(ctx, "state:", map[string]any{"data": data})

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Read: errors!")
		return
	}

	tflog.Info(ctx, "Read: start")

	lab, err := r.cfg.Client().LabGet(ctx, data.ID.Value, true)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to fetch lab, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Read: lab state: %s", lab.State))

	data.ID = types.String{Value: lab.ID}
	data.State = types.String{Value: lab.State}
	data.Nodes = r.populateNodes(ctx, lab, &resp.Diagnostics)
	data.Booted = types.Bool{Value: lab.Booted()}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Read: errors!")
		return
	}
	tflog.Info(ctx, "Read: done")
}

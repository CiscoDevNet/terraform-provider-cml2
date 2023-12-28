package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

func (r LabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		data cmlschema.LabModel
		err  error
	)

	tflog.Info(ctx, "Resource Lab UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lab := cmlclient.Lab{
		ID:          data.ID.ValueString(),
		Notes:       data.Notes.ValueString(),
		Description: data.Description.ValueString(),
		Title:       data.Title.ValueString(),
	}

	groupList := make([]*cmlclient.LabGroup, 0)
	if !data.Groups.IsUnknown() {
		var model cmlschema.LabGroupModel
		for _, elem := range data.Groups.Elements() {
			tfsdk.ValueAs(ctx, elem, &model)
			el := cmlclient.LabGroup{
				ID:         model.ID.ValueString(),
				Permission: model.Permission.ValueString(),
			}
			groupList = append(groupList, &el)
		}
	}
	lab.Groups = groupList

	newLab, err := r.cfg.Client().LabUpdate(ctx, lab)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to update lab, got error: %s", err),
		)
		return
	}

	value := cmlschema.NewLab(ctx, newLab, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &value)...)

	tflog.Info(ctx, "Resource Lab UPDATE done")
}

package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func (r *LabResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		labModel cmlschema.LabModel
		err      error
	)

	tflog.Info(ctx, "Resource Lab CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &labModel)...)
	if resp.Diagnostics.HasError() {
		return
	}
	planNodeStaging := labModel.NodeStaging

	createReq := models.LabCreateRequest{}
	if !labModel.Notes.IsNull() {
		createReq.Notes = labModel.Notes.ValueString()
	}
	if !labModel.Description.IsNull() {
		createReq.Description = labModel.Description.ValueString()
	}
	if !labModel.Title.IsNull() {
		createReq.Title = labModel.Title.ValueString()
	}

	desiredNodeStaging := expandNodeStaging(ctx, labModel.NodeStaging, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if desiredNodeStaging != nil {
		// Ensure client knows server version (SkipReadyCheck is used at init).
		if err := r.cfg.Client().Ready(ctx); err != nil {
			resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("Unable to check CML readiness/version, got error: %s", err))
			return
		}
		ok, err := r.cfg.Client().VersionCheck(ctx, ">=2.10.0")
		if err != nil {
			resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("Unable to check CML version, got error: %s", err))
			return
		}
		if !ok {
			resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("node_staging requires CML >= 2.10.0 (detected %s)", r.cfg.Client().Version()))
			return
		}
		// Prefer fail-fast create semantics when supported.
		createReq.NodeStaging = desiredNodeStaging
	}

	if !labModel.Groups.IsUnknown() && !labModel.Groups.IsNull() {
		groups := make([]models.LabGroup, 0)
		var g cmlschema.LabGroupModel
		for _, elem := range labModel.Groups.Elements() {
			resp.Diagnostics.Append(tfsdk.ValueAs(ctx, elem, &g)...)
			if resp.Diagnostics.HasError() {
				return
			}
			perm := models.OldPermissionReadOnly
			if g.Permission.ValueString() == string(models.OldPermissionReadWrite) {
				perm = models.OldPermissionReadWrite
			}
			groups = append(groups, models.LabGroup{ID: models.UUID(g.ID.ValueString()), Permission: perm})
		}
		createReq.Groups = groups
	}

	newLab, err := r.cfg.Client().Lab.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to create lab, got error: %s", err),
		)
		return
	}

	// Refresh to get populated groups from API.
	fullLab, err := r.cfg.Client().Lab.GetByID(ctx, newLab.ID, false)
	if err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("Unable to get lab, got error: %s", err))
		return
	}

	// Defensive enforcement: if create did not apply node_staging, enforce via PATCH.
	if desiredNodeStaging != nil {
		if fullLab.NodeStaging == nil ||
			fullLab.NodeStaging.Enabled != desiredNodeStaging.Enabled ||
			fullLab.NodeStaging.StartRemaining != desiredNodeStaging.StartRemaining ||
			fullLab.NodeStaging.AbortOnFailure != desiredNodeStaging.AbortOnFailure {
			_, err = r.cfg.Client().Lab.Update(ctx, newLab.ID, models.LabUpdateRequest{NodeStaging: desiredNodeStaging})
			if err != nil {
				resp.Diagnostics.AddError(
					common.ErrorLabel,
					fmt.Sprintf("Unable to enforce lab node_staging (CML %s), got error: %s", r.cfg.Client().Version(), err),
				)
				return
			}
			fullLab, err = r.cfg.Client().Lab.GetByID(ctx, newLab.ID, false)
			if err != nil {
				resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("Unable to get lab after node_staging update, got error: %s", err))
				return
			}
		}
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(ctx, cmlschema.NewLab(ctx, &fullLab, &resp.Diagnostics), types.ObjectType{AttrTypes: cmlschema.LabAttrType}, &labModel)...,
	)
	keepNodeStagingNullWhenUnmanaged(planNodeStaging, &labModel)
	resp.Diagnostics.Append(resp.State.Set(ctx, &labModel)...)

	tflog.Info(ctx, "Resource Lab CREATE done")
}

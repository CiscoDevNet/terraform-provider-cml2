package user

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func userGroupIDsFromSet(_ context.Context, _ *diag.Diagnostics, set types.Set) []models.UUID {
	if set.IsUnknown() || set.IsNull() {
		return nil
	}

	groupIDs := make([]models.UUID, 0, len(set.Elements()))
	for _, v := range set.Elements() {
		id := v.(types.String).ValueString()
		if id == "" {
			continue
		}
		groupIDs = append(groupIDs, models.UUID(id))
	}
	return groupIDs
}

func stateGroupIDsFromSet(ctx context.Context, diags *diag.Diagnostics, set types.Set) []models.UUID {
	return userGroupIDsFromSet(ctx, diags, set)
}

func uuidSet(ids []models.UUID) map[models.UUID]struct{} {
	m := make(map[models.UUID]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return m
}

func removeUUID(ids []models.UUID, target models.UUID) []models.UUID {
	out := ids[:0]
	for _, id := range ids {
		if id == target {
			continue
		}
		out = append(out, id)
	}
	return out
}

func (r *UserResource) reconcileGroupMembership(ctx context.Context, diags *diag.Diagnostics, userID models.UUID, current, desired []models.UUID) {
	cur := uuidSet(current)
	des := uuidSet(desired)

	for gid := range cur {
		if _, ok := des[gid]; ok {
			continue
		}
		g, err := r.cfg.Client().Group.GetByID(ctx, gid)
		if err != nil {
			diags.AddError(common.ErrorLabel, fmt.Sprintf("unable to get group %s: %s", gid, err))
			return
		}
		update := models.Group{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Members:     removeUUID(g.Members, userID),
		}
		if _, err := r.cfg.Client().Group.Update(ctx, update); err != nil {
			diags.AddError(common.ErrorLabel, fmt.Sprintf("unable to update group %s: %s", gid, err))
			return
		}
	}

	for gid := range des {
		if _, ok := cur[gid]; ok {
			continue
		}
		g, err := r.cfg.Client().Group.GetByID(ctx, gid)
		if err != nil {
			diags.AddError(common.ErrorLabel, fmt.Sprintf("unable to get group %s: %s", gid, err))
			return
		}
		members := g.Members
		if !slices.Contains(members, userID) {
			members = append(members, userID)
		}
		update := models.Group{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Members:     members,
		}
		if _, err := r.cfg.Client().Group.Update(ctx, update); err != nil {
			diags.AddError(common.ErrorLabel, fmt.Sprintf("unable to update group %s: %s", gid, err))
			return
		}
	}
}

package lab

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
)

func expandGroupAssociations(ctx context.Context, groups types.Set, diags *diag.Diagnostics) models.AssociationUsersGroups {
	associations := models.AssociationUsersGroups{}
	if groups.IsUnknown() || groups.IsNull() {
		return associations
	}

	groupAssocs := make([]models.Association, 0, len(groups.Elements()))
	var g cmlschema.LabGroupModel
	for _, elem := range groups.Elements() {
		diags.Append(tfsdk.ValueAs(ctx, elem, &g)...)
		if diags.HasError() {
			return associations
		}
		groupAssocs = append(groupAssocs, models.Association{
			ID:          models.UUID(g.ID.ValueString()),
			Permissions: cmlschema.AssociationPermissionsFromTFGroupPermission(g.Permission.ValueString()),
		})
	}

	if len(groupAssocs) > 0 {
		associations.Groups = groupAssocs
	}

	return associations
}

func flattenLabGroupsFromAssociations(groups []models.Group, labID models.UUID) []models.LabGroup { //nolint:staticcheck
	labGroups := make([]models.LabGroup, 0) //nolint:staticcheck
	for _, group := range groups {
		for _, assoc := range group.Associations {
			if assoc.ID != labID {
				continue
			}
			labGroups = append(labGroups, models.LabGroup{ //nolint:staticcheck
				ID:         group.ID,
				Name:       group.Name,
				Permission: models.OldPermission(cmlschema.TFGroupPermissionFromAssociationPermissions(assoc.Permissions)), //nolint:staticcheck
			})
			break
		}
	}

	sort.Slice(labGroups, func(i, j int) bool {
		return labGroups[i].ID < labGroups[j].ID
	})

	return labGroups
}

func (r *LabResource) hydrateGroups(ctx context.Context, lab *models.Lab) error {
	groups, err := r.cfg.Client().Group.Groups(ctx)
	if err != nil {
		return err
	}
	lab.Groups = flattenLabGroupsFromAssociations(groups, lab.ID) //nolint:staticcheck
	return nil
}

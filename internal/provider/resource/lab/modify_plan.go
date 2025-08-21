package lab

import (
	"context"
	"reflect"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *LabResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var stateData, planData cmlschema.LabModel

	tflog.Info(ctx, "Resource Lab MODIFYPLAN")

	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	// Read Terraform plan and state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// tflog.Info(ctx, fmt.Sprintf("XXX  plan: %v", planData.Groups))
	// tflog.Info(ctx, fmt.Sprintf("XXX state: %v", stateData.Groups))

	// this makes TF crash
	// planData.Groups = types.SetUnknown(
	// 	types.ObjectType{AttrTypes: cmlschema.LabGroupAttrType},
	// )

	// maybe related to
	// https://github.com/hashicorp/terraform-plugin-framework/issues/628
	// this should go into the unequal block below...!
	// resp.Diagnostics.Append(
	// 	resp.Plan.SetAttribute(
	// 		ctx, path.Root("groups"),
	// 		types.SetUnknown(
	// 			types.ObjectType{AttrTypes: cmlschema.LabGroupAttrType},
	// 		),
	// 	)...,
	// )

	// if state and plan are NOT identical -> modified date has changed
	// this gets auto-updated when we change something
	if !reflect.DeepEqual(stateData, planData) {
		planData.Modified = types.StringUnknown()
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, planData)...)

	// maybe related to
	// https://github.com/hashicorp/terraform-plugin-framework/issues/628
	// this should go into the unequal block below...!
	// resp.Diagnostics.Append(
	// 	resp.Plan.SetAttribute(
	// 		ctx, path.Root("groups"),
	// 		types.SetUnknown(
	// 			types.ObjectType{
	// 				AttrTypes: cmlschema.LabGroupAttrType,
	// 			},
	// 			// types.SetType{
	// 			// 	ElemType: types.ObjectType{
	// 			// 		AttrTypes: cmlschema.LabGroupAttrType,
	// 			// 	},
	// 			// },
	// 			// types.ObjectType.WithAttributeTypes(
	// 			// 	types.ObjectType{}, cmlschema.LabGroupAttrType,
	// 			// ),
	// 		),
	// 	)...,
	// )

	tflog.Info(ctx, "Resource Lab MODIFYPLAN done")
}

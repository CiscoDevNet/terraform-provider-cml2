package link

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *LinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	// var (
	// 	data *schema.NodeModel
	// 	err  error
	// )

	tflog.Info(ctx, "Resource Link DELETE")

	// resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // if data.State.Value != cmlclient.NodeStateDefined {
	// // 	resp.Diagnostics.AddError(CML2ErrorLabel, "node is not in DEFINED_ON_CORE state")
	// // 	return
	// // }

	// node := &cmlclient.Node{ID: data.ID.Value, LabID: data.LabID.Value}

	// err = r.client.NodeStop(ctx, node)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		CML2ErrorLabel,
	// 		fmt.Sprintf("Unable to stop node, got error: %s", err),
	// 	)
	// 	return
	// }

	// err = r.client.NodeWipe(ctx, node)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		CML2ErrorLabel,
	// 		fmt.Sprintf("Unable to wipe node, got error: %s", err),
	// 	)
	// 	return
	// }

	// err = r.client.NodeDestroy(ctx, node)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		CML2ErrorLabel,
	// 		fmt.Sprintf("Unable to destroy node, got error: %s", err),
	// 	)
	// 	return
	// }
	tflog.Info(ctx, "Resource Link DELETE: done")
}

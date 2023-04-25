package link

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *LinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Resource Link DELETE")
	// this is a no-op at this point as links are removed automatically
	// when nodes and their interfaces are deleted
	tflog.Info(ctx, "Resource Link DELETE done")
}

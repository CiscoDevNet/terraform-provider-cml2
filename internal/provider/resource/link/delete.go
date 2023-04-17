package link

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *LinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Resource Link DELETE")
	// this is a no-op at this point as links are removed automatically
	// when nodes and their interfaces are deleted

	// when nodes are deleted then this apparently causes a race of some sort.
	// wait for 100ms to allow the controller to settle
	// attempt to work around SIMPLE-5368 until the root cause is fixed!
	<-time.After(100 * time.Millisecond)

	tflog.Info(ctx, "Resource Link DELETE done")
}

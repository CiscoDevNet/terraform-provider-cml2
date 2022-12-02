package link

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

func (r *LinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var (
		data cmlschema.LinkModel
		err  error
	)

	// can only create one link at a time... this is because parallel link
	// creation can allocate the same interface twice. E.g. when link1 is
	// created and grabs e.g. i0 on a node, link2 might grab the same interface
	// on that node as there's no way for the client to tell that there's
	// something going on in parallel.
	//
	// This becomes even more complicated if we assume parallel access to the
	// same lab/resource by multiple clients.  Unless we implement some more
	// complex logic/caching in the client itself.
	//
	// Right now, this is a trade-off between blunt sequential creation and a
	// complex implementation.

	r.cfg.Lock()
	defer r.cfg.Unlock()

	tflog.Info(ctx, "Resource Link CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	link := cmlclient.Link{}
	link.LabID = data.LabID.ValueString()
	link.SrcNode = data.NodeA.ValueString()
	link.DstNode = data.NodeB.ValueString()
	if !data.NodeAslot.IsUnknown() {
		slot := int(data.NodeAslot.ValueInt64())
		link.SrcSlot = &slot
	}
	if !data.NodeBslot.IsUnknown() {
		slot := int(data.NodeBslot.ValueInt64())
		link.DstSlot = &slot
	}

	newLink, err := r.cfg.Client().LinkCreate(ctx, &link)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to create link, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("src slot %d", *newLink.SrcSlot))
	tflog.Info(ctx, fmt.Sprintf("dst slot %d", *newLink.DstSlot))

	data.NodeAslot = types.Int64Value(int64(*newLink.SrcSlot))
	data.NodeBslot = types.Int64Value(int64(*newLink.DstSlot))

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewLink(ctx, newLink, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.LinkAttrType},
			&data,
		)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Link CREATE: done")
}

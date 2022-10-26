package lifecycle

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/common"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func getTimeouts(ctx context.Context, config tfsdk.Config, diags *diag.Diagnostics) *labLifecycleTimeouts {
	// timeouts is optional, if ommitted it will result in a nil pointer
	var timeouts *labLifecycleTimeouts
	diags.Append(config.GetAttribute(ctx, path.Root("timeouts"), &timeouts)...)
	if diags.HasError() || timeouts == nil {
		tflog.Warn(ctx, "timeouts undefined, using defaults")
		return &labLifecycleTimeouts{
			Create: types.String{Value: "2h"},
			Delete: types.String{Value: "2h"},
			Update: types.String{Value: "2h"},
		}
	}
	tflog.Info(ctx, fmt.Sprintf("timeouts: %+v", timeouts))
	return timeouts
}

func getStaging(ctx context.Context, config tfsdk.Config, diags *diag.Diagnostics) *labLifecycleStaging {
	var staging *labLifecycleStaging
	diags.Append(config.GetAttribute(ctx, path.Root("staging"), &staging)...)
	tflog.Info(ctx, fmt.Sprintf("staging: %+v", staging))
	// default for this is true
	if staging != nil && staging.StartRemaining.IsNull() {
		tflog.Info(ctx, "setting start remaining to true, default value")
		staging.StartRemaining.Null = false
		staging.StartRemaining.Value = true
	}
	return staging
}

func (r *LabLifecycleResource) stop(ctx context.Context, diags diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab stop")
	err := r.cfg.Client().LabStop(ctx, id)
	if err != nil {
		diags.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to stop CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab stop done")
}

func (r *LabLifecycleResource) wipe(ctx context.Context, diags diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab wipe")
	err := r.cfg.Client().LabWipe(ctx, id)
	if err != nil {
		diags.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab wipe done")
}

func (r *LabLifecycleResource) startNodesAll(ctx context.Context, diags *diag.Diagnostics, start startData) {
	tflog.Info(ctx, "lab start")
	err := r.cfg.Client().LabStart(ctx, start.lab.ID)
	if err != nil {
		diags.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to start lab, got error: %s", err),
		)
	}
	tflog.Info(ctx, "lab start done")
	if start.wait {
		timeout := start.timeouts.Create.Value
		common.Converge(ctx, r.cfg.Client(), diags, start.lab.ID, timeout)
	}
}

func (r *LabLifecycleResource) startNodes(ctx context.Context, diags *diag.Diagnostics, start startData) {

	// start all nodes at once, no staging
	if start.staging == nil {
		r.startNodesAll(ctx, diags, start)
		return
	}

	// start nodes in stages
	for _, stage_elem := range start.staging.Stages.Elems {
		stage := stage_elem.(types.String).Value
		for _, node := range start.lab.Nodes {
			for _, tag := range node.Tags {
				if tag == stage {
					tflog.Info(ctx, fmt.Sprintf("starting node %s", node.Label))
					err := r.cfg.Client().NodeStart(ctx, node)
					if err != nil {
						diags.AddError(
							CML2ErrorLabel,
							fmt.Sprintf("Unable to start node %s, got error: %s", node.Label, err),
						)
					}
				}
			}
		}
		// this is not 100% correct as the timeout is applied to each stage
		// should be: timeout applied to all stages combined
		timeout := start.timeouts.Create.Value
		common.Converge(ctx, r.cfg.Client(), diags, start.lab.ID, timeout)
	}

	// start remaining nodes, if indicated
	if start.staging.StartRemaining.Value {
		tflog.Info(ctx, "starting remaining nodes")
		r.startNodesAll(ctx, diags, start)
	}
}

func (r *LabLifecycleResource) injectConfigs(ctx context.Context, lab *cmlclient.Lab, data *schema.LabLifecycleModel, diags *diag.Diagnostics) {
	tflog.Info(ctx, "injectConfigs")

	if data.Configs.IsNull() {
		tflog.Info(ctx, "injectConfigs: no configs")
		return
	}

	for nodeID, config := range data.Configs.Elems {
		node, err := lab.NodeByLabel(ctx, nodeID)
		if err == cmlclient.ErrElementNotFound {
			node = lab.Nodes[nodeID]
		}
		if node == nil {
			diags.AddError(CML2ErrorLabel, fmt.Sprintf("node with label %s not found", nodeID))
			continue
		}
		if node.State != cmlclient.NodeStateDefined {
			diags.AddError(CML2ErrorLabel, fmt.Sprintf("unexpected node state %s", node.State))
			continue
		}
		config_string := config.(types.String).Value
		err = r.cfg.Client().NodeSetConfig(ctx, node, config_string)
		if err != nil {
			diags.AddError("set node config failed",
				fmt.Sprintf("setting the new node configuration failed: %s", err),
			)
		}
	}
	tflog.Info(ctx, "injectConfigs: done")
}

func (r *LabLifecycleResource) populateNodes(ctx context.Context, lab *cmlclient.Lab, diags *diag.Diagnostics) types.Map {
	// we want this as a stable sort by node UUID
	nodeList := []*cmlclient.Node{}
	for _, node := range lab.Nodes {
		nodeList = append(nodeList, node)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})
	nodes := types.Map{
		ElemType: types.ObjectType{AttrTypes: schema.NodeAttrType},
		Elems:    make(map[string]attr.Value),
	}
	for _, node := range nodeList {
		nodes.Elems[node.ID] = schema.NewNode(ctx, node, diags)
	}
	return nodes
}

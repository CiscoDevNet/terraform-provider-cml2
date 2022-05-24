package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = cmlNodeDefDataSourceType{}
var _ tfsdk.DataSource = cml2NodeDefDataSource{}

type cmlNodeDefDataSourceType struct{}

func (t cmlNodeDefDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the
		// language server.
		MarkdownDescription: "CML2 IP lab addresses data source",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Lab ID (UUID)",
				Type:                types.StringType,
				Required:            true,
			},
			"state": {
				MarkdownDescription: "Lab state (DEFINED_ON_CORE, STOPPED or STARTED)",
				Type:                types.StringType,
				Computed:            true,
			},
			"filter": {
				MarkdownDescription: "node type filter (regex)",
				Optional:            true,
				Type:                types.StringType,
			},
			"only_with_ip": {
				MarkdownDescription: "only consider nodes with an IP address",
				Optional:            true,
				Type:                types.BoolType,
			},
			"nodes": {
				MarkdownDescription: "List of nodes and their interfaces with IP addresses",
				Computed:            true,
				// Type: types.ListType{
				// 	ElemType: resultNode{},
				// },
				Attributes: tfsdk.ListNestedAttributes(
					nodeSchema(),
					tfsdk.ListNestedAttributesOptions{},
				),
			},
		},
	}, nil
}

func interfaceSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			MarkdownDescription: "Interface ID (UUID)",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"mac_address": {
			MarkdownDescription: "MAC address",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"is_connected": {
			MarkdownDescription: "connection status",
			Type:                types.BoolType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"ip4": {
			MarkdownDescription: "IPv4 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"ip6": {
			MarkdownDescription: "IPv6 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
	}
}

func nodeSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			MarkdownDescription: "Node ID (UUID)",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"nodetype": {
			MarkdownDescription: "Node Type / Definition",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"interfaces": {
			MarkdownDescription: "interfaces on the node",
			Computed:            true,
			Attributes: tfsdk.ListNestedAttributes(
				interfaceSchema(),
				tfsdk.ListNestedAttributesOptions{},
			),
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
	}
}

func (t cmlNodeDefDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return cml2NodeDefDataSource{
		provider: provider,
	}, diags
}

// type cml2Node struct {
// 	Id         types.String    `tfsdk:"id"`
// 	Label      types.String    `tfsdk:"label"`
// 	State      types.String    `tfsdk:"state"`
// 	NodeType   types.String    `tfsdk:"nodetype"`
// 	Interfaces []cml2Interface `tfsdk:"interfaces"`
// }

// type cml2Interface struct {
// 	Id          types.String   `tfsdk:"id"`
// 	Label       types.String   `tfsdk:"label"`
// 	State       types.String   `tfsdk:"state"`
// 	MACaddress  types.String   `tfsdk:"mac_address"`
// 	IsConnected types.Bool     `tfsdk:"is_connected"`
// 	IP4         []types.String `tfsdk:"ip4"`
// 	IP6         []types.String `tfsdk:"ip6"`
// }

type cml2NodeDefDataSourceData struct {
	Id         types.String `tfsdk:"id"`
	State      types.String `tfsdk:"state"`
	Filter     types.String `tfsdk:"filter"`
	OnlyWithIP types.Bool   `tfsdk:"only_with_ip"`
	// Nodes      []cml2Node   `tfsdk:"nodes"`
}

type cml2NodeDefDataSource struct {
	provider cml2
}

func (d cml2NodeDefDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data cml2NodeDefDataSourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	lab, err := d.provider.client.GetLab(ctx, data.Id.Value, false)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to fetch lab, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("!!!! state: %s", lab.State))

	data.State = types.String{Value: lab.State}

	// needsIP := false
	if !data.OnlyWithIP.Null && data.OnlyWithIP.Value {
		tflog.Info(ctx, "nodes need IP addresses to be considered!")
		// needsIP = true
	}

	// we want this as a stable sort by node UUID
	nodeList := []*cmlclient.Node{}
	for _, node := range lab.Nodes {
		nodeList = append(nodeList, node)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})

	// data.Nodes = types.List{
	// 	Elems:    make([]attr.Value, 0),
	// 	ElemType: nodeSchema().AttributeType(),
	// }

	for _, node := range nodeList {
		rnode := cml2Node{
			Id:       types.String{Value: node.ID},
			Label:    types.String{Value: node.Label},
			State:    types.String{Value: node.State},
			NodeType: types.String{Value: node.NodeDefinition},
		}
		_ = rnode
		// hasIP := false

		// we want this as a stable sort by interface UUID
		ilist := []*cmlclient.Interface{}
		for _, iface := range node.Interfaces {
			ilist = append(ilist, iface)
		}
		sort.Slice(ilist, func(i, j int) bool {
			return ilist[i].ID < ilist[j].ID
		})

		// data.Nodes = types.List{
		// 	ElemType: resultNode{},

		// 	// types.ObjectType{
		// 	// 	AttrTypes: map[string]attr.Type{
		// 	// 		"id": nodeList[0],
		// 	// 	},
		// 	// Id         types.String      `tfsdk:"id"`
		// 	// Label      types.String      `tfsdk:"label"`
		// 	// State      types.String      `tfsdk:"state"`
		// 	// Type       types.String      `tfsdk:"type"`
		// 	// Interfaces []resultInterface `tfsdk:"interfaces"`
		// 	// },
		// 	// Elem.Type: types.ListType{
		// 	// 	ElemType: resultNode,
		// 	// },
		// }
		// tflog.Info(ctx, "node", map[string]interface{}{"node": rnode})

		// data.Nodes.Elems = append(data.Nodes.Elems, rnode)

		// for _, iface := range ilist {
		// 	if needsIP && len(iface.IP4) == 0 && len(iface.IP6) == 0 {
		// 		continue
		// 	}
		// 	hasIP = true
		// 	riface := resultInterface{
		// 		Id:          types.String{Value: iface.ID},
		// 		Label:       types.String{Value: iface.Label},
		// 		State:       types.String{Value: iface.State},
		// 		MACaddress:  types.String{Value: iface.MACaddress},
		// 		IsConnected: types.Bool{Value: iface.IsConnected},
		// 	}
		// 	for _, ip := range iface.IP4 {
		// 		riface.IP4 = append(riface.IP4, types.String{Value: ip})
		// 	}
		// 	for _, ip := range iface.IP6 {
		// 		riface.IP6 = append(riface.IP6, types.String{Value: ip})
		// 	}
		// 	rnode.Interfaces = append(rnode.Interfaces, riface)
		// }
		// if needsIP && !hasIP {
		// 	continue
		// }
		// data.Nodes = append(data.Nodes, rnode)
	}

	// tflog.Info(ctx, "$$$", map[string]interface{}{"id": data.Id.Value})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

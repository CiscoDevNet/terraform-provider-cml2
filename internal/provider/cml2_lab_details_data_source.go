package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = cmlLabDetailDataSourceType{}
var _ tfsdk.DataSource = cml2LabDetailDataSource{}

type cmlLabDetailDataSourceType struct{}

func (t cmlLabDetailDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
				Sensitive:           false,
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
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
		},
		"mac_address": {
			MarkdownDescription: "MAC address",
			Type:                types.StringType,
			Computed:            true,
		},
		"is_connected": {
			MarkdownDescription: "connection status",
			Type:                types.BoolType,
			Computed:            true,
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
		},
		"ip4": {
			MarkdownDescription: "IPv4 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Sensitive: false,
		},
		"ip6": {
			MarkdownDescription: "IPv6 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Sensitive: false,
		},
	}
}

func nodeSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			MarkdownDescription: "Node ID (UUID)",
			Type:                types.StringType,
			Computed:            true,
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
		},
		"nodetype": {
			MarkdownDescription: "Node Type / Definition",
			Type:                types.StringType,
			Computed:            true,
		},
		"interfaces": {
			MarkdownDescription: "interfaces on the node",
			Computed:            true,
			Sensitive:           false,
			Attributes: tfsdk.ListNestedAttributes(
				interfaceSchema(),
				tfsdk.ListNestedAttributesOptions{},
			),
		},
	}
}

func (t cmlLabDetailDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return cml2LabDetailDataSource{
		provider: provider,
	}, diags
}

type resultInterface struct {
	Id          types.String   `tfsdk:"id"`
	Label       types.String   `tfsdk:"label"`
	State       types.String   `tfsdk:"state"`
	MACaddress  types.String   `tfsdk:"mac_address"`
	IsConnected types.Bool     `tfsdk:"is_connected"`
	IP4         []types.String `tfsdk:"ip4"`
	IP6         []types.String `tfsdk:"ip6"`
}

type resultNode struct {
	Id         types.String      `tfsdk:"id"`
	Label      types.String      `tfsdk:"label"`
	State      types.String      `tfsdk:"state"`
	NodeType   types.String      `tfsdk:"nodetype"`
	Interfaces []resultInterface `tfsdk:"interfaces"`
}

// func (rn resultNode) Equal(v attr.Type) bool {
// 	return true
// }

// func (rn resultNode) String() string {
// 	return "as"
// }

// func (rn resultNode) TerraformType(ctx context.Context) tftypes.Type {
// 	return tftypes.DynamicPseudoType
// }

// func (rn resultNode) ApplyTerraform5AttributePathStep(ps tftypes.AttributePathStep) (interface{}, error) {
// 	return nil, nil
// }

// func (rn resultNode) ValueFromTerraform(ctx context.Context, v tftypes.Value) (attr.Value, error) {
// 	return nil, nil
// }

func (rn resultNode) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	v := tftypes.NewValue(
		tftypes.Tuple{},
		rn,
	)
	return v, nil
}

// func (rn resultNode) Type(ctx context.Context) attr.Type {
// 	return nodeSchema().AttributeType()
// }

type cml2DataSourceData struct {
	Id         types.String `tfsdk:"id"`
	State      types.String `tfsdk:"state"`
	Filter     types.String `tfsdk:"filter"`
	OnlyWithIP types.Bool   `tfsdk:"only_with_ip"`
	Nodes      []resultNode `tfsdk:"nodes"`
	// Nodes types.List `tfsdk:"nodes"`
}

type cml2LabDetailDataSource struct {
	provider cml2
}

func (d cml2LabDetailDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data cml2DataSourceData

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

	needsIP := false
	if !data.OnlyWithIP.Null && data.OnlyWithIP.Value {
		tflog.Info(ctx, "nodes need IP addresses to be considered!")
		needsIP = true
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
		rnode := resultNode{
			Id:       types.String{Value: node.ID},
			Label:    types.String{Value: node.Label},
			State:    types.String{Value: node.State},
			NodeType: types.String{Value: node.NodeDefinition},
		}
		hasIP := false

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

		for _, iface := range ilist {
			if needsIP && len(iface.IP4) == 0 && len(iface.IP6) == 0 {
				continue
			}
			hasIP = true
			riface := resultInterface{
				Id:          types.String{Value: iface.ID},
				Label:       types.String{Value: iface.Label},
				State:       types.String{Value: iface.State},
				MACaddress:  types.String{Value: iface.MACaddress},
				IsConnected: types.Bool{Value: iface.IsConnected},
			}
			for _, ip := range iface.IP4 {
				riface.IP4 = append(riface.IP4, types.String{Value: ip})
			}
			for _, ip := range iface.IP6 {
				riface.IP6 = append(riface.IP6, types.String{Value: ip})
			}
			rnode.Interfaces = append(rnode.Interfaces, riface)
		}
		if needsIP && !hasIP {
			continue
		}
		data.Nodes = append(data.Nodes, rnode)
	}

	// tflog.Info(ctx, "$$$", map[string]interface{}{"id": data.Id.Value})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

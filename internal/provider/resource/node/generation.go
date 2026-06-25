package node

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
)

type normalizedNamedConfig struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type generationPayload struct {
	Label           *string                 `json:"label"`
	NodeDefinition  *string                 `json:"nodedefinition"`
	Configuration   *string                 `json:"configuration"`
	Configurations  []normalizedNamedConfig `json:"configurations"`
	ImageDefinition *string                 `json:"imagedefinition"`
	RAM             *int64                  `json:"ram"`
	CPUs            *int64                  `json:"cpus"`
	CPUlimit        *int64                  `json:"cpu_limit"`
	BootDiskSize    *int64                  `json:"boot_disk_size"`
	DataVolume      *int64                  `json:"data_volume"`
}

func ptrString(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func ptrInt64(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

func normalizedConfigs(ctx context.Context, data cmlschema.NodeModel) ([]normalizedNamedConfig, error) {
	if data.Configurations.IsNull() || data.Configurations.IsUnknown() {
		return nil, nil
	}
	out := make([]normalizedNamedConfig, 0, len(data.Configurations.Elements()))
	for _, el := range data.Configurations.Elements() {
		var nc cmlschema.NamedConfigModel
		if diags := tfsdk.ValueAs(ctx, el, &nc); diags.HasError() {
			return nil, fmt.Errorf("convert named config for generation: %v", diags)
		}
		out = append(out, normalizedNamedConfig{Name: nc.Name.ValueString(), Content: nc.Content.ValueString()})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].Content < out[j].Content
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func generationFromNodeModel(ctx context.Context, data cmlschema.NodeModel) (types.String, error) {
	cfgs, err := normalizedConfigs(ctx, data)
	if err != nil {
		return types.StringNull(), err
	}
	payload := generationPayload{
		Label:           ptrString(data.Label),
		NodeDefinition:  ptrString(data.NodeDefinition),
		Configuration:   nil,
		Configurations:  cfgs,
		ImageDefinition: ptrString(data.ImageDefinition),
		RAM:             ptrInt64(data.RAM),
		CPUs:            ptrInt64(data.CPUs),
		CPUlimit:        ptrInt64(data.CPUlimit),
		BootDiskSize:    ptrInt64(data.BootDiskSize),
		DataVolume:      ptrInt64(data.DataVolume),
	}
	if !data.Configuration.IsNull() && !data.Configuration.IsUnknown() {
		cfg := data.Configuration.ValueString()
		payload.Configuration = &cfg
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return types.StringNull(), err
	}
	sum := sha256.Sum256(b)
	return types.StringValue(hex.EncodeToString(sum[:])), nil
}

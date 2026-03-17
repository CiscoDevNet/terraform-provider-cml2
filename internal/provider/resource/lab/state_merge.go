package lab

import (
	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func keepNodeStagingNullWhenUnmanaged(managedNS types.Object, m *cmlschema.LabModel) {
	if managedNS.IsNull() || managedNS.IsUnknown() {
		m.NodeStaging = types.ObjectNull(cmlschema.LabNodeStagingAttrType)
	}
}

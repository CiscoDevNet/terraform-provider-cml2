package lab

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
)

func keepNodeStagingNullWhenUnmanaged(managedNS types.Object, m *cmlschema.LabModel) {
	if managedNS.IsNull() || managedNS.IsUnknown() {
		m.NodeStaging = types.ObjectNull(cmlschema.LabNodeStagingAttrType)
	}
}

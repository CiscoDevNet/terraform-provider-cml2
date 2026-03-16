package lifecycle_test

import (
	"fmt"
	"testing"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// This test targets node definitions that are expected to be container-based
// ("docker" style) rather than VM-heavy images.
//
// Topology intent:
// external_connector -- ioll2-xe -- chrome
//
// Note: The exact node definition IDs must exist on the target controller.
// If your controller uses different IDs for these, adjust them in the config.

var testAccDockerProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cml2": providerserver.NewProtocol6WithError(cml.New("test")()),
}

func TestAccLifecycleDockerChain(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: testAccDockerProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleDockerChain(cfg.Cfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
				),
			},
		},
	})
}

func testAccLifecycleDockerChain(cfg string) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_lab" "this" {
	title = "acc docker chain"
}

resource "cml2_node" "ext" {
	lab_id         = cml2_lab.this.id
	label          = "Internet"
	nodedefinition = "external_connector"
	configuration  = "NAT"
	tags           = ["stage-1"]
}

resource "cml2_node" "rtr" {
	lab_id         = cml2_lab.this.id
	label          = "R1"
	nodedefinition = "ioll2-xe"
	tags           = ["stage-2"]
}

resource "cml2_node" "chrome" {
	lab_id         = cml2_lab.this.id
	label          = "Chrome"
	nodedefinition = "chrome"
	tags           = ["stage-3"]
}

resource "cml2_link" "l1" {
	lab_id = cml2_lab.this.id
	node_a = cml2_node.ext.id
	node_b = cml2_node.rtr.id
}

resource "cml2_link" "l2" {
	lab_id = cml2_lab.this.id
	node_a = cml2_node.rtr.id
	node_b = cml2_node.chrome.id
}

resource "cml2_lifecycle" "top" {
	lab_id = cml2_lab.this.id
	depends_on = [
		cml2_node.ext,
		cml2_node.rtr,
		cml2_node.chrome,
		cml2_link.l1,
		cml2_link.l2,
	]
	staging = {
		stages          = ["stage-1", "stage-2", "stage-3"]
		start_remaining = false
	}
	// wait until all nodes booted. for fast runs, flip this to false
	wait = true
}
`, cfg)
}

package node_test

import (
	"fmt"
	"testing"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cml2": providerserver.NewProtocol6WithError(cml.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for
	// example assertions about the appropriate environment variables being set
	// are common to see in a pre-check function.
}

func TestNodeDataSource(t *testing.T) {
	title := "thislab"
	label := "thetestnode"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testNodeDataSourceConfig(cfg.Cfg, title, label),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("theoutput", label),
				),
			},
		},
	})
}

func testNodeDataSourceConfig(cfg, title, label string) string {
	return fmt.Sprintf(`
	%[1]s
	resource "cml2_lab" "lab" {
			title = %[2]q
	}
	resource "cml2_node" "node" {
		  lab_id = cml2_lab.lab.id
		  nodedefinition = "alpine"
			label = %[3]q
	}
	data "cml2_node" "acc_test" {
		id = cml2_node.node.id
		lab_id = cml2_lab.lab.id
	}
	output "theoutput" {
		value = data.cml2_node.acc_test.node.label
	}
	`, cfg, title, label)
}

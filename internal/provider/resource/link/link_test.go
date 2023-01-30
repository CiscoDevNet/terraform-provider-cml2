package link_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	cml "github.com/rschmied/terraform-provider-cml2/internal/provider"
	cfg "github.com/rschmied/terraform-provider-cml2/internal/testing"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cml2": providerserver.NewProtocol6WithError(cml.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for
	// example assertions about the appropriate environment variables being set
	// are common to see in a pre-check function.
}

func TestAccLinkResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLinkResourceConfig(cfg.Cfg),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_link.l0", "label", "r1-eth3<->r2-eth2"),
				),
			},
			{
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.cml2_node.r1", "node.interfaces.#", "4"),
				),
				RefreshState: true,
			},
			// ImportState testing
			// {
			// 	ResourceName:      "cml2_link.test",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
			// Update and Read testing
			// {
			// 	Config: testAccLabResourceConfig(cfg.Cfg),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("cml2_lab.test", "title", "newtitle"),
			// 		resource.TestCheckResourceAttr("cml2_lab.test", "description", "newdesc"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccLifecycleResourceDaniel(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleDaniel(cfg.Cfg),
			},
		},
	})
}

func testAccLinkResourceConfig(cfg string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
}
resource "cml2_node" "r1" {
	lab_id         = cml2_lab.test.id
	label          = "r1"
	nodedefinition = "alpine"
}
resource "cml2_node" "r2" {
	lab_id         = cml2_lab.test.id
	label          = "r2"
	nodedefinition = "alpine"
}
resource "cml2_link" "l0" {
	lab_id = cml2_lab.test.id
	node_a = cml2_node.r1.id
	node_b = cml2_node.r2.id
	slot_a = 3
	slot_b = 2
}
data "cml2_node" "r1" {
	id = cml2_node.r1.id
	lab_id = cml2_lab.test.id
}
`, cfg)
}

// this specifically tests the omission of link interface slots which should
// result in "use next free slot" as defined by the CML client. This was broken
// in 0.5.1 and earlier.
func testAccLifecycleDaniel(cfg string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "devnet-expert" {
	title       = "DevNet Expert Lab"
	description = "This is the DevNet Expert Lab for study"
  }

resource "cml2_node" "ext" {
	lab_id         = cml2_lab.devnet-expert.id
	nodedefinition = "external_connector"
	label          = "Internet"
	configuration  = "bridge0"
  }

  resource "cml2_node" "nat1" {
	lab_id         = cml2_lab.devnet-expert.id
	label          = "NAT"
	nodedefinition = "iosv"
  }

  resource "cml2_node" "ums1" {
	lab_id         = cml2_lab.devnet-expert.id
	label          = "MGMT"
	nodedefinition = "unmanaged_switch"
  }

  resource "cml2_node" "cws1" {
	lab_id         = cml2_lab.devnet-expert.id
	label          = "CWS"
	nodedefinition = "alpine"
  }

  resource "cml2_link" "l0" {
	lab_id = cml2_lab.devnet-expert.id
	node_a = cml2_node.ext.id
	node_b = cml2_node.nat1.id
  }

  resource "cml2_link" "l1" {
    lab_id = cml2_lab.devnet-expert.id
    node_a = cml2_node.nat1.id
    node_b = cml2_node.ums1.id
  }

  resource "cml2_link" "l2" {
	lab_id = cml2_lab.devnet-expert.id
	node_a = cml2_node.ums1.id
	node_b = cml2_node.cws1.id
  }
  `, cfg)
}

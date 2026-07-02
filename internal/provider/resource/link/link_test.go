package link_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"

	"github.com/rschmied/gocmlclient/pkg/models"
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
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLinkResourceConfig(cfg.Cfg),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_link.l0", "label", "r1-eth0<->r2-eth0"),
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_a", "0"),
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_b", "0"),
				),
			},
			{
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.cml2_node.r1", "node.interfaces.#", "1"),
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

func TestAccLinkResourceRecreatesWhenDeletedExternally(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	config := testAccLinkResourceConfig(cfg.Cfg)
	var initialLinkID string
	var initialLabID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_link.l0", "label", "r1-eth0<->r2-eth0"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_link.l0"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_link.l0")
						}
						initialLinkID = rs.Primary.ID
						initialLabID = rs.Primary.Attributes["lab_id"]
						if initialLinkID == "" || initialLabID == "" {
							return fmt.Errorf("expected lab_id and id to be set in state")
						}
						return nil
					},
				),
			},
			{
				Config:             config,
				ExpectNonEmptyPlan: true,
				Check: func(s *terraform.State) error {
					client, err := cfg.NewCMLClientFromTFEnv()
					if err != nil {
						return err
					}
					return client.Link.Delete(context.Background(), models.UUID(initialLabID), models.UUID(initialLinkID))
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_link.l0", "label", "r1-eth0<->r2-eth0"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_link.l0"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_link.l0")
						}
						if rs.Primary.ID == initialLinkID {
							return fmt.Errorf("expected link to be recreated; id still %q", initialLinkID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccLifecycleResourceDaniel(t *testing.T) {
	cfg.SkipUnlessAcc(t)

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

func TestAccLifecycleResourceSlotChange(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create a link, not specifying any slots
			{
				Config: testAccLinkResourceConfigSlotChange(cfg.Cfg, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_a", "0"),
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_b", "0"),
				),
			},
			// modify the slots for this link, this needs to create a plan change
			{
				Config: testAccLinkResourceConfigSlotChange(cfg.Cfg, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_a", "1"),
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_b", "2"),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// apply this change
			{
				Config: testAccLinkResourceConfigSlotChange(cfg.Cfg, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_a", "1"),
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_b", "2"),
				),
			},
			// change the config back to not specifying any links, should still be
			// the same.
			{
				Config: testAccLinkResourceConfigSlotChange(cfg.Cfg, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_a", "1"),
					resource.TestCheckResourceAttr("cml2_link.l0", "slot_b", "2"),
				),
				PlanOnly: true,
			},
		},
	})
}

func testAccLinkResourceConfig(cfg string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
	title = "acc link resource"
}
resource "cml2_node" "r1" {
	lab_id         = cml2_lab.test.id
	label          = "r1"
	nodedefinition = "nginx"
}
resource "cml2_node" "r2" {
	lab_id         = cml2_lab.test.id
	label          = "r2"
	nodedefinition = "nginx"
}
resource "cml2_link" "l0" {
	lab_id = cml2_lab.test.id
	node_a = cml2_node.r1.id
	node_b = cml2_node.r2.id
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
	title       = "acc link devnet expert"
	description = "This is the DevNet Expert Lab for study"
  }

resource "cml2_node" "ext" {
	lab_id         = cml2_lab.devnet-expert.id
	nodedefinition = "external_connector"
	label          = "Internet"
	configuration  = "virbr0"
  }

  resource "cml2_node" "nat1" {
	lab_id         = cml2_lab.devnet-expert.id
	label          = "NAT"
	nodedefinition = "iol-xe"
  }

  resource "cml2_node" "ums1" {
	lab_id         = cml2_lab.devnet-expert.id
	label          = "MGMT"
	nodedefinition = "unmanaged_switch"
  }

  resource "cml2_node" "cws1" {
	lab_id         = cml2_lab.devnet-expert.id
	label          = "CWS"
	nodedefinition = "nginx"
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

func testAccLinkResourceConfigSlotChange(cfg string, step int) string {
	var slotCfg string
	if step == 0 {
		slotCfg = ""
	} else {
		slotCfg = `
          slot_a = 1
          slot_b = 2
		`
	}
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
	title = "acc link slot change"
}
resource "cml2_node" "r1" {
	lab_id         = cml2_lab.test.id
	label          = "r1"
	nodedefinition = "ioll2-xe"
}
resource "cml2_node" "r2" {
	lab_id         = cml2_lab.test.id
	label          = "r2"
	nodedefinition = "ioll2-xe"
}
resource "cml2_link" "l0" {
	lab_id = cml2_lab.test.id
	node_a = cml2_node.r1.id
	node_b = cml2_node.r2.id
	%[2]s
}
data "cml2_node" "r1" {
	id = cml2_node.r1.id
	lab_id = cml2_lab.test.id
}
`, cfg, slotCfg)
}

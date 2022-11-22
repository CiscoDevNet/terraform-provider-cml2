package node_test

import (
	"fmt"
	"regexp"
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

func TestAccNodeResource(t *testing.T) {
	re1 := regexp.MustCompile(`Node Definition not found:`)
	re2 := regexp.MustCompile(`expected "alpine", got "iosv"`)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:      testAccNodeResourceConfig(cfg.Cfg, "doesntexist"),
				ExpectError: re1,
			},
			{
				Config: testAccNodeResourceConfig(cfg.Cfg, "alpine"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "nodedefinition", "alpine"),
					resource.TestCheckResourceAttr("cml2_node.r1", "x", "98"),
					resource.TestCheckResourceAttr("cml2_node.r1", "y", "99"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "1"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.0", "test"),
				),
			},
			{
				Config: testAccNodeResourceConfig2(cfg.Cfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "nodedefinition", "alpine"),
					resource.TestCheckResourceAttr("cml2_node.r1", "x", "100"),
					resource.TestCheckResourceAttr("cml2_node.r1", "y", "200"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "2"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.0", "test"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.1", "someothertag"),
				),
			},
			{
				Config: testAccNodeResourceConfig(cfg.Cfg, "iosv"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "nodedefinition", "alpine"),
				),
				ExpectError: re2,
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

func TestAccLifecycleNodeProps(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResourceRam(cfg.Cfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "ram", "512"),
				),
			},
		},
	})
}

func testAccNodeResourceConfig(cfg, node_def string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
}
resource "cml2_node" "r1" {
	lab_id         = cml2_lab.test.id
	label          = "r1"
	nodedefinition = "%[2]s"
	x              = 98
	y              = 99
	tags           = ["test"]
}
`, cfg, node_def)
}

func testAccNodeResourceConfig2(cfg string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
}
resource "cml2_node" "r1" {
	lab_id         = cml2_lab.test.id
	label          = "r1"
	x              = 100
	y              = 200
	nodedefinition = "alpine"
	tags           = ["test", "someothertag"]
}
`, cfg)
}

func testAccNodeResourceRam(cfg string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.test.id
  label          = "R1"
  ram            = 512
  nodedefinition = "alpine"
}
`, cfg)
}

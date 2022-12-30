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
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNodeResourceConfigNodeDefInvalid(cfg.Cfg),
				ExpectError: re1,
			},
			{
				Config: testAccNodeResourceConfig(cfg.Cfg, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "nodedefinition", "alpine"),
					resource.TestCheckResourceAttr("cml2_node.r1", "label", "alpine-0"),
					resource.TestCheckNoResourceAttr("cml2_node.r1", "imagedefinition"),
					resource.TestCheckResourceAttr("cml2_node.r1", "x", "100"),
					resource.TestCheckResourceAttr("cml2_node.r1", "y", "100"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "1"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.0", "test"),
				),
			},
			{
				// ExpectNonEmptyPlan: true,
				Config: testAccNodeResourceConfig(cfg.Cfg, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "nodedefinition", "alpine"),
					resource.TestCheckResourceAttr("cml2_node.r1", "label", "alpine-99"),
					resource.TestCheckResourceAttr("cml2_node.r1", "x", "100"),
					resource.TestCheckResourceAttr("cml2_node.r1", "y", "200"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "2"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.0", "test"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.1", "tag2"),
				),
			},
			{
				Config: testAccNodeResourceConfig(cfg.Cfg, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "nodedefinition", "alpine"),
					resource.TestCheckResourceAttr("cml2_node.r1", "label", "alpine-99"),
					resource.TestCheckResourceAttrSet("cml2_node.r1", "imagedefinition"),
					resource.TestCheckResourceAttr("cml2_node.r1", "x", "100"),
					resource.TestCheckResourceAttr("cml2_node.r1", "y", "200"),
					resource.TestCheckResourceAttr("cml2_node.r1", "ram", "1024"),
					resource.TestCheckResourceAttr("cml2_node.r1", "cpus", "2"),
					resource.TestCheckResourceAttr("cml2_node.r1", "boot_disk_size", "64"),
					resource.TestCheckResourceAttr("cml2_node.r1", "data_volume", "64"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "2"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.0", "test"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.1", "tag2"),
				),
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

func testAccNodeResourceConfigNodeDefInvalid(cfg string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
}
resource "cml2_node" "r1" {
	lab_id         = cml2_lab.test.id
	label          = "r1"
	nodedefinition = "invalid"
}
`, cfg)
}

func testAccNodeResourceConfig(cfg string, step int) string {
	if step == 1 {
		return fmt.Sprintf(`
		%[1]s
		resource "cml2_lab" "test" {
		}
		resource "cml2_node" "r1" {
			lab_id          = cml2_lab.test.id
			label           = "alpine-0"
			x               = 100
			y               = 100
			nodedefinition  = "alpine"
			tags            = [ "test" ]
		}
		`, cfg)
	}
	if step == 2 {
		return fmt.Sprintf(`
		%[1]s
		resource "cml2_lab" "test" {
		}
		resource "cml2_node" "r1" {
			lab_id          = cml2_lab.test.id
			label           = "alpine-99"
			x               = 100
			y               = 200
			nodedefinition  = "alpine"
			tags            = [ "test", "tag2" ]
		}
		`, cfg)
	}
	if step == 3 {
		return fmt.Sprintf(`
		%[1]s
		data "cml2_images" "test" {
			nodedefinition = "alpine"
		}
		resource "cml2_lab" "test" {
		}
		resource "cml2_node" "r1" {
			lab_id          = cml2_lab.test.id
			label           = "alpine-99"
			x               = 100
			y               = 200
			nodedefinition  = "alpine"
			imagedefinition = element(data.cml2_images.test.image_list, 0).id
			ram             = 1024
			cpus            = 2
			boot_disk_size  = 64
			data_volume     = 64
			tags            = [ "test", "tag2" ]
		}
		`, cfg)
	}
	panic("undefined step!")
}

// 	tagline := ""
// 	if len(tags) > 0 {
// 		for idx, tag := range tags {
// 			tags[idx] = fmt.Sprintf("%q", tag)
// 		}
// 		tagline = fmt.Sprintf("tags = [%s]\n", strings.Join(tags, ","))
// 	}

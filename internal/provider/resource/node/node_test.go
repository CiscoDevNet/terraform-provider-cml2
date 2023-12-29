package node_test

import (
	"fmt"
	"regexp"
	"strings"
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

func TestAccNodeResourceCreateAllAttrs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResourceCreateAllAttrs(cfg.Cfg),
			},
		},
	})
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
					resource.TestCheckResourceAttr("cml2_node.r1", "hide_links", "false"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("cml2_node.r1", "tags.*", "test"),
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
					resource.TestCheckResourceAttr("cml2_node.r1", "hide_links", "true"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("cml2_node.r1", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("cml2_node.r1", "tags.*", "tag2"),
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
					resource.TestCheckResourceAttr("cml2_node.r1", "hide_links", "false"),
					resource.TestCheckResourceAttr("cml2_node.r1", "ram", "1024"),
					resource.TestCheckResourceAttr("cml2_node.r1", "cpus", "2"),
					resource.TestCheckResourceAttr("cml2_node.r1", "boot_disk_size", "64"),
					resource.TestCheckResourceAttr("cml2_node.r1", "data_volume", "64"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("cml2_node.r1", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("cml2_node.r1", "tags.*", "tag2"),
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

func TestAccNodeResourceTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResourceConfigTags(cfg.Cfg, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "0"),
				),
			},
			{
				Config: testAccNodeResourceConfigTags(cfg.Cfg, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("cml2_node.r1", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("cml2_node.r1", "tags.*", "tag2"),
				),
			},
			{
				Config: testAccNodeResourceConfigTags(cfg.Cfg, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "1"),
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.0", "tag2"),
				),
			},
			{
				Config: testAccNodeResourceConfigTags(cfg.Cfg, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "0"),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// need to re-run to apply the change
				Config: testAccNodeResourceConfigTags(cfg.Cfg, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "0"),
				),
			},
			{
				Config: testAccNodeResourceConfigTags(cfg.Cfg, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "0"),
				),
			},
			{
				Config: testAccNodeResourceConfigTags(cfg.Cfg, 6),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "tags.#", "0"),
				),
			},
		},
	})
}

func TestAccNodeResourceEmptyConfig(t *testing.T) {
	empty := ""
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResourceConfigEmpty(cfg.Cfg, &empty),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.r1", "configuration", ""),
				),
			},
		},
	})
}

func TestAccNodeResourceNullConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResourceConfigEmpty(cfg.Cfg, nil),
				Check: resource.TestCheckResourceAttrWith("cml2_node.r1", "configuration", func(value string) error {
					expected := "this is a shell script which"
					if strings.Contains(value, expected) {
						return nil
					}
					return fmt.Errorf("expected %q to contain %q", value, expected)
				}),
			},
		},
	})
}

func TestAccNodeResourceExtConn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeResourceConfigNodeDefExtConn(cfg.Cfg, ""),
				Check: resource.TestCheckResourceAttrWith("cml2_node.ext", "configuration", func(value string) error {
					expected := "NAT"
					if value == expected {
						return nil
					}
					return fmt.Errorf("expected %q to contain %q", value, expected)
				}),
				// Destroy: ,
			},
			{
				Config: testAccNodeResourceConfigNodeDefExtConn(cfg.Cfg, "NAT"),
				Check: resource.TestCheckResourceAttrWith("cml2_node.ext", "configuration", func(value string) error {
					expected := "NAT"
					if value == expected {
						return nil
					}
					return fmt.Errorf("expected %q to contain %q", value, expected)
				}),
			},
			// this tests the error condition in Update!
			{
				Config:      testAccNodeResourceConfigNodeDefExtConn(cfg.Cfg, "virbr0"),
				ExpectError: regexp.MustCompile("provide proper external connector configuration, not a device name"),
			},
		},
	})
	// same test, this time the error is raised in Create!
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNodeResourceConfigNodeDefExtConn(cfg.Cfg, "virbr0"),
				ExpectError: regexp.MustCompile("provide proper external connector configuration, not a device name"),
			},
		},
	})
}

func testAccNodeResourceConfigNodeDefExtConn(cfg, extconnname string) string {
	var config string
	if len(extconnname) > 0 {
		config = fmt.Sprintf("configuration = %q", extconnname)
	}
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
}
resource "cml2_node" "ext" {
	lab_id         = cml2_lab.test.id
	label          = "ext0"
	nodedefinition = "external_connector"
	%[2]s
}
`, cfg, config)
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

func testAccNodeResourceCreateAllAttrs(cfg string) string {
	return fmt.Sprintf(`
		%[1]s
		data "cml2_images" "test" {
			nodedefinition = "alpine"
		}
		resource "cml2_lab" "test" {
		}
		resource "cml2_node" "r1" {
			lab_id          = cml2_lab.test.id
			label           = "alpine-0"
			x               = 100
			y               = 100
			nodedefinition  = "alpine"
			tags            = [ "test" ]
			configuration   = "hostname bla"
			ram             = 2048
			cpus            = 2
			cpu_limit       = 90
			boot_disk_size  = 64
			data_volume     = 64
			imagedefinition = element(data.cml2_images.test.image_list, 0).id
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
			hide_links      = true
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
			hide_links      = false
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

func testAccNodeResourceConfigTags(cfg string, step int) string {
	var tags string
	switch step {
	case 1:
		tags = ""
	case 2:
		tags = "tags = [ \"test\", \"tag2\" ]"
	case 3:
		tags = "tags = [ \"tag2\" ]"
	case 4:
		tags = ""
	case 5:
		tags = "tags = [ ]"
	case 6:
		tags = ""
	default:
		panic("undefined step!")
	}
	return fmt.Sprintf(`
	%[1]s
	resource "cml2_lab" "test" {
	}
	resource "cml2_node" "r1" {
		lab_id          = cml2_lab.test.id
		label           = "alpine-0"
		nodedefinition  = "alpine"
		%[2]s
	}
	`, cfg, tags)
}

// 	tagline := ""
// 	if len(tags) > 0 {
// 		for idx, tag := range tags {
// 			tags[idx] = fmt.Sprintf("%q", tag)
// 		}
// 		tagline = fmt.Sprintf("tags = [%s]\n", strings.Join(tags, ","))
// 	}

func testAccNodeResourceConfigEmpty(cfg string, nodeCfg *string) string {
	if nodeCfg != nil {
		return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
}
resource "cml2_node" "r1" {
	lab_id         = cml2_lab.test.id
	label          = "r1"
	nodedefinition = "alpine"
	configuration  = %[2]q
}
`, cfg, *nodeCfg)
	}

	// no configuration when the node config is null
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "test" {
}
resource "cml2_node" "r1" {
	lab_id         = cml2_lab.test.id
	label          = "r1"
	nodedefinition = "alpine"
}
`, cfg)
}

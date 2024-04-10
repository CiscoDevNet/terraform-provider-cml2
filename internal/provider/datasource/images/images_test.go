package images_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	cml "github.com/rschmied/terraform-provider-cml2/internal/provider"
	cfg "github.com/rschmied/terraform-provider-cml2/internal/testing"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cml2": providerserver.NewProtocol6WithError(cml.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for
	// example assertions about the appropriate environment variables being set
	// are common to see in a pre-check function.
}

func TestImageDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testImageDataSourceConfig(cfg.Cfg, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.cml2_images.test", "image_list.#", func(value string) error {
						num, err := strconv.Atoi(value)
						if err == nil && num < 10 {
							return fmt.Errorf("expected at least 10 image definitions, got %d", num)
						}
						return err
					}),
				),
			},
			{
				Config: testImageDataSourceConfig(cfg.Cfg, "alpine"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("bla", "alpine"),
				),
			},
		},
	})
}

func testImageDataSourceConfig(cfg, nd string) string {
	ndCfg := ""
	if len(nd) > 0 {
		ndCfg = fmt.Sprintf("nodedefinition = %q", nd)
	}
	return fmt.Sprintf(`
	%[1]s
	data "cml2_images" "test" {
		%[2]s
	}
	locals {
		il = data.cml2_images.test.image_list
	}
	output "bla" {
		value = element(local.il, length(local.il)-1).nodedefinition
	}
	`, cfg, ndCfg)
}

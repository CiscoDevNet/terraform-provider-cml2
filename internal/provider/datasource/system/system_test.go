package system_test

import (
	"fmt"
	"regexp"
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

func TestSystemDataSource(t *testing.T) {
	re1 := regexp.MustCompile(`ran into timeout`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testSystemDataSourceConfig(cfg.CfgBroken, 8),
				ExpectError: re1,
			},
			{
				Config: testSystemDataSourceConfig(cfg.CfgBroken, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("bla", "false"),
				),
			},
			{
				Config: testSystemDataSourceConfig(cfg.Cfg, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("bla", "true"),
				),
			},
		},
	})
}

func testSystemDataSourceConfig(cfg string, timeout int) string {
	return fmt.Sprintf(`
	%[1]s
	data "cml2_system" "test" {
		timeout = "%[2]ds"
	}
	output "bla" {
		value = data.cml2_system.test.ready
	}
	`, cfg, timeout)
}

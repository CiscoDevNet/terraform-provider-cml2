package extconn_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cml2": providerserver.NewProtocol6WithError(cml.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for
	// example assertions about the appropriate environment variables being set
	// are common to see in a pre-check function.
}

func TestExtConnDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testExtConnDataSourceConfig(cfg.Cfg, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.cml2_connector.test", "connectors.#", func(value string) error {
						num, err := strconv.Atoi(value)
						if err == nil && num == 0 {
							return fmt.Errorf("expected at least one connector definition, got %d", num)
						}
						return err
					}),
				),
			},
			{
				Config: testExtConnDataSourceConfig(cfg.Cfg, "", "NAT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("bla", "virbr0"),
				),
			},
			{
				Config: testExtConnDataSourceConfig(cfg.Cfg, "NAT", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("bla", "virbr0"),
				),
			},
		},
	})
}

func testExtConnDataSourceConfig(cfg, label, tag string) string {
	connectorCfg := ""
	if len(label) > 0 {
		connectorCfg = fmt.Sprintf("label = %q", label)
	}
	if len(tag) > 0 {
		connectorCfg += fmt.Sprintf("\ntag = %q", tag)
	}
	return fmt.Sprintf(`
	%[1]s
	data "cml2_connector" "test" {
		%[2]s
	}
	locals {
		cl = data.cml2_connector.test.connectors
	}
	output "bla" {
		value = element(local.cl, 0).device_name
	}
	`, cfg, connectorCfg)
}

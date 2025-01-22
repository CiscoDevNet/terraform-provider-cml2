package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

func TestAccHTTPScheck(t *testing.T) {
	re := regexp.MustCompile(`A valid CML server URL using HTTPS must be provided.`)
	for _, url := range []string{"()!@*(#$&", "https://"} {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccHTTPScheckCfg(url),
					ExpectError: re,
				},
			},
		})
	}

	re = regexp.MustCompile(`Can't parse server address`)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccHTTPScheckCfg("http://cml.bla. com"),
				ExpectError: re,
			},
		},
	})
}

func testAccHTTPScheckCfg(address string) string {
	return fmt.Sprintf(`
provider "cml2" {
	address = "%[1]s"
	token = "something"
}
resource "cml2_lifecycle" "top" {
	topology = "{}"
}
`, address)
}

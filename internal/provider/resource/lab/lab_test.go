package lab_test

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

func TestAccLabResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLabResourceConfig(cfg.Cfg, "thistitle", "description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lab.test", "title", "thistitle"),
					resource.TestCheckResourceAttr("cml2_lab.test", "description", "description"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cml2_lab.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"title"},
			},
			// Update and Read testing
			{
				Config: testAccLabResourceConfig(cfg.Cfg, "newtitle", "newdesc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lab.test", "title", "newtitle"),
					resource.TestCheckResourceAttr("cml2_lab.test", "description", "newdesc"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLabResourceConfig(cfg, title, description string) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_group" "group1" {
	name       = "user_acc_lab_test_group1"
}

resource "cml2_group" "group2" {
	name       = "user_acc_lab_test_group2"
}

resource "cml2_lab" "test" {
	title       = %[2]q
	description = %[3]q
	notes       = <<-EOT
	# Heading
	- topic one
	- topic two
	This is where it's ending... PEBKAC
	EOT
	groups = [
		{
			id = cml2_group.group1.id
			permission = "read_only"
		},
		{
			id = cml2_group.group2.id
			permission = "read_only"
		}
	]
}
`, cfg, title, description)
}

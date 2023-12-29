package users_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func TestUsersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testGroupDataSourceConfig(cfg.Cfg, "admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("bla", "admin"),
				),
			},
		},
	})
}

func testGroupDataSourceConfig(cfg, username string) string {
	return fmt.Sprintf(`
	%[1]s
	data "cml2_users" "acc_test" {
		username = %[2]q
	}
	output "bla" {
		value = data.cml2_users.acc_test.users[0].username
	}
	`, cfg, username)
}

package user_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
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

// RandomString generates a simple hex suffix for tests
func RandomString(n int) string {
	b := make([]byte, n/2)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(b)
	return fmt.Sprintf("%x", b)
}

func TestAccUserResource(t *testing.T) {
	cfg.SkipUnlessAcc(t)
	suffix := RandomString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResourceConfig(cfg.Cfg, suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_user.acc_test", "description", "acc test user description"),
					resource.TestCheckResourceAttr("cml2_user.acc_test", "groups.#", "2"),
					resource.TestCheckResourceAttr("cml2_user.acc_test", "is_admin", "true"),
				),
			},
			// Update and Read testing
			{
				Config: testAccUserResourceConfigUpdate(cfg.Cfg, suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_user.acc_test", "description", "has changed"),
					resource.TestCheckResourceAttr("cml2_user.acc_test", "groups.#", "0"),
					resource.TestCheckResourceAttr("cml2_user.acc_test", "is_admin", "false"),
				),
				// PlanOnly:           true,
				// ExpectNonEmptyPlan: true,
			},
			// ImportState testing
			{
				ResourceName:      "cml2_user.acc_test",
				ImportState:       true,
				ImportStateVerify: true,
				// password will be unknown at import
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUserResourceConfig(cfg, suffix string) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_group" "group1" {
	name       = "user_acc_test1_group1_%[2]s"
}

resource "cml2_group" "group2" {
	name       = "user_acc_test1_group2_%[2]s"
}

resource "cml2_user" "acc_test" {
	username      = "acc_test_user_%[2]s"
	password      = "süpersücret"
	fullname      = "firstname, lastname"
	email         = "bla@cml.lab"
	description   = "acc test user description"
	is_admin      = true
	# resource_pool = "e0e18ef5-9d1f-4cbb-99e8-a6da60c20113"
	groups = [ cml2_group.group1.id, cml2_group.group2.id ]
}
`, cfg, suffix)
}

func testAccUserResourceConfigUpdate(cfg, suffix string) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_group" "group1" {
	name       = "user_acc_test2_group1_%[2]s"
}

resource "cml2_group" "group2" {
	name       = "user_acc_test2_group2_%[2]s"
}

resource "cml2_user" "acc_test" {
	username      = "acc_test_user_%[2]s"
	password      = "süpersücret"
	fullname      = "firstname, lastname"
	email         = "bla@cml.lab"
	description   = "has changed"
	is_admin      = false
	# resource_pool = "e0e18ef5-9d1f-4cbb-99e8-a6da60c20113"
	groups = []
}
`, cfg, suffix)
}

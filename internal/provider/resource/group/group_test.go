package group_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
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

func TestAccGroupResource(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	// Keep short to satisfy CML username length constraints.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	// user max length is 32 chars; to stay safely within CML validation,
	// keep the suffix at 9 digits.
	suffix := fmt.Sprintf("%09d", rng.Int63()%1_000_000_000)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGroupResourceConfig(cfg.Cfg, suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_group.test", "description", "description"),
					resource.TestCheckResourceAttr("cml2_group.test", "labs.#", "2"),
					resource.TestCheckResourceAttr("cml2_group.test", "members.#", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cml2_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccGroupResourceConfigUpdate(cfg.Cfg, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_group.test", "name", "new name"),
					resource.TestCheckResourceAttr("cml2_group.test", "description", "new description"),
					resource.TestCheckResourceAttr("cml2_group.test", "labs.#", "1"),
					resource.TestCheckResourceAttr("cml2_group.test", "members.#", "0"),
				),
			},
			{
				Config: testAccGroupResourceConfigUpdate2(cfg.Cfg, "read_write", suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_group.test", "labs.#", "1"),
					resource.TestCheckResourceAttr("cml2_group.test", "members.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("cml2_group.test", "labs.*", map[string]string{
						"permission": "read_write",
					}),
				),
			},
			{
				Config: testAccGroupResourceConfigUpdate2(cfg.Cfg, "read_only", suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_group.test", "labs.#", "1"),
					resource.TestCheckResourceAttr("cml2_group.test", "members.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("cml2_group.test", "labs.*", map[string]string{
						"permission": "read_only",
					}),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccGroupResourceRecreatesWhenDeletedExternally(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	// CML enforces max 32 chars for usernames; keep names short to avoid 400s.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	suffix := fmt.Sprintf("%09d", rng.Int63()%1_000_000_000)

	config := testAccGroupResourceConfig(cfg.Cfg, suffix)
	var initialGroupID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_group.test", "name", fmt.Sprintf("acc_test_group_%s", suffix)),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_group.test"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_group.test")
						}
						initialGroupID = rs.Primary.ID
						if initialGroupID == "" {
							return fmt.Errorf("expected cml2_group.test.id")
						}
						return nil
					},
				),
			},
			{
				Config:             config,
				ExpectNonEmptyPlan: true,
				Check: func(s *terraform.State) error {
					client, err := cfg.NewCMLClientFromTFEnv()
					if err != nil {
						return err
					}
					return client.Group.Delete(context.Background(), initialGroupID)
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_group.test", "name", fmt.Sprintf("acc_test_group_%s", suffix)),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_group.test"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_group.test")
						}
						if rs.Primary.ID == initialGroupID {
							return fmt.Errorf("expected group to be recreated; id still %q", initialGroupID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccGroupResourceNoLists(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGroupResourceConfigNoLists(cfg.Cfg),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_group.test", "description", "description"),
					resource.TestCheckResourceAttr("cml2_group.test", "labs.#", "0"),
					resource.TestCheckResourceAttr("cml2_group.test", "members.#", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccGroupResourceConfig(cfg, suffix string) string {
	groupUserName := fmt.Sprintf("acc_test_group_user_%s", suffix)
	groupName := fmt.Sprintf("acc_test_group_%s", suffix)
	lab1Title := "acc group lab1"
	lab2Title := "acc group lab2"

	return fmt.Sprintf(`
%[1]s

resource "cml2_user" "acc_test" {
	username      = %[2]q
	password      = "süpersücret"
	fullname      = "firstname, lastname"
	email         = "bla@cml.lab"
	description   = "acc test user description"
	is_admin      = false
}

resource "cml2_lab" "lab1" {
	title       = %[3]q
}

resource "cml2_lab" "lab2" {
	title       = %[4]q
}

resource "cml2_group" "test" {
	description = "description"
	name = %[5]q
	members = [ cml2_user.acc_test.id ]
	labs = [
	{
		id = cml2_lab.lab1.id
		permission = "read_only"
	},
	{
		id = cml2_lab.lab2.id
		permission = "read_only"
	}
]
}
`, cfg, groupUserName, lab1Title, lab2Title, groupName)
}

func testAccGroupResourceConfigUpdate(cfg, suffix string) string {
	groupUserName := fmt.Sprintf("acc_test_group_user_%s", suffix)
	lab1Title := "acc group lab1"
	lab2Title := "acc group lab2"

	return fmt.Sprintf(`
%[1]s

resource "cml2_user" "acc_test" {
	username      = %[2]q
	password      = "süpersücret"
	fullname      = "firstname, lastname"
	email         = "bla@cml.lab"
	description   = "acc test user description"
	is_admin      = false
}

resource "cml2_lab" "lab1" {
	title       = %[3]q
}

resource "cml2_lab" "lab2" {
	title       = %[4]q
}

resource "cml2_group" "test" {
	description = "new description"
	name = "new name"
	members = []
	labs = [
	{
		id = cml2_lab.lab1.id
		permission = "read_only"
	}
]
}
`, cfg, groupUserName, lab1Title, lab2Title)
}

func testAccGroupResourceConfigUpdate2(cfg, permission, suffix string) string {
	groupUserName := fmt.Sprintf("acc_test_group_user_%s", suffix)
	groupUser2Name := fmt.Sprintf("acc_test_group_user_2_%s", suffix)
	lab1Title := "acc group lab1"

	return fmt.Sprintf(`
%[1]s

resource "cml2_user" "acc_test" {
	username      = %[2]q
	password      = "süpersücret"
	fullname      = "firstname, lastname"
	email         = "bla@cml.lab"
	description   = "acc test user description"
	is_admin      = false
}

resource "cml2_user" "acc_test_2" {
	username      = %[3]q
	password      = "süpersücret"
	fullname      = "firstname, lastname"
	email         = "bla@cml.lab"
	description   = "acc test user description"
	is_admin      = false
}

resource "cml2_lab" "lab1" {
	title       = %[4]q
}

resource "cml2_group" "test" {
	description = "new description"
	name = "new name"
	members = [ cml2_user.acc_test.id, cml2_user.acc_test_2.id ]
	labs = [
	{
		id = cml2_lab.lab1.id
		permission = %[5]q
	},
]
}
`, cfg, groupUserName, groupUser2Name, lab1Title, permission)
}

func testAccGroupResourceConfigNoLists(cfg string) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_group" "test" {
	description = "description"
	name = "new name"
}
`, cfg)
}

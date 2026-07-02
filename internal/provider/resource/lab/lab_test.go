package lab_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/rschmied/gocmlclient/pkg/models"

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

func TestAccLabResource(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	const title = "acc lab resource"

	// Use unique group names to avoid 409 "already exists" errors across test runs.
	group1Name := fmt.Sprintf("user_acc_lab_test_group1_%d", time.Now().UnixNano())
	group2Name := fmt.Sprintf("user_acc_lab_test_group2_%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLabResourceConfig(cfg.Cfg, "description", 1, group1Name, group2Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lab.test", "title", title),
					resource.TestCheckResourceAttr("cml2_lab.test", "description", "description"),
					resource.TestCheckResourceAttr("cml2_lab.test", "groups.#", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cml2_lab.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"node_staging",
				},
			},
			// Update and Read testing
			{
				// disabled for now, raised
				// https://github.com/hashicorp/terraform-plugin-framework/issues/709
				// Update: workaround for now is to disable the UseStateForUnknown modifier
				// in the schema for the nested schema objects in the set.
				// SkipFunc: func() (bool, error) { return true, nil },
				Config: testAccLabResourceConfig(cfg.Cfg, "newdesc", 2, group1Name, group2Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lab.test", "title", title),
					resource.TestCheckResourceAttr("cml2_lab.test", "description", "newdesc"),
					resource.TestCheckResourceAttr("cml2_lab.test", "groups.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("cml2_lab.test", "groups.*", map[string]string{
						"permission": "read_write",
					}),
				),
			},
			{
				// should use the disabled one above and remove this.
				// using this to have some test for update.
				Config: testAccLabResourceConfig(cfg.Cfg, "newdesc", 1, group1Name, group2Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lab.test", "title", title),
					resource.TestCheckResourceAttr("cml2_lab.test", "description", "newdesc"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccLabResourceRecreatesWhenDeletedExternally(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	const title = "acc lab resource"
	config := "" // set below with unique group names
	var initialLabID string

	// Use unique group names to avoid 409 "already exists" errors across test runs.
	group1Name := fmt.Sprintf("user_acc_lab_test_group1_%d", time.Now().UnixNano())
	group2Name := fmt.Sprintf("user_acc_lab_test_group2_%d", time.Now().UnixNano())
	config = testAccLabResourceConfig(cfg.Cfg, "description", 1, group1Name, group2Name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lab.test", "title", title),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lab.test"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lab.test")
						}
						initialLabID = rs.Primary.ID
						if initialLabID == "" {
							return fmt.Errorf("expected cml2_lab.test.id")
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
					// delete the lab directly to simulate external drift
					return client.Lab.Delete(context.Background(), models.UUID(initialLabID))
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lab.test", "title", title),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lab.test"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lab.test")
						}
						if rs.Primary.ID == initialLabID {
							return fmt.Errorf("expected lab to be recreated; id still %q", initialLabID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccLabResourceConfig(cfg, description string, group int, group1Name, group2Name string) string {
	var groupCfg string
	if group == 1 {
		groupCfg = `
		{
			id = cml2_group.group1.id
			permission = "read_only"
		},
		{
			id = cml2_group.group2.id
			permission = "read_only"
		}
		`
	} else {
		groupCfg = `
		{
			id = cml2_group.group2.id
			permission = "read_write"
		}
		`
	}

	// Keep fmt placeholders sequential to avoid accidental index drift.
	return fmt.Sprintf(`
	%[1]s

	resource "cml2_group" "group1" {
		name = %q
	}

	resource "cml2_group" "group2" {
		name = %q
	}

	resource "cml2_lab" "test" {
		title       = "acc lab resource"
		description = %q
		notes       = <<-EOT
		# Heading
		- topic one
		- topic two
		This is where it's ending... PEBKAC
		EOT
		node_staging = {
			enabled = false
		}
		groups = [
			%s
		]
	}
`, cfg, group1Name, group2Name, description, groupCfg)
}

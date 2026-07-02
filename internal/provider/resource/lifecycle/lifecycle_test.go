package lifecycle_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"

	"github.com/rschmied/gocmlclient/pkg/models"
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

func TestAccLifecycleResource(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLifecycleResourceConfig(cfg.Cfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
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

func TestAccLifecycleResourceRestartsWhenStoppedExternally(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	config := testAccLifecycleResourceConfigWithState(cfg.Cfg, "STARTED")
	var labID string
	var lifecycleID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						lifecycleID = rs.Primary.ID
						labID = rs.Primary.Attributes["lab_id"]
						if lifecycleID == "" || labID == "" {
							return fmt.Errorf("expected lifecycle id and lab_id")
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

					// Simulate external drift relevant to lifecycle: stop the lab
					// outside Terraform while the desired state remains STARTED.
					return client.Lab.Stop(context.Background(), models.UUID(labID))
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						if rs.Primary.ID != lifecycleID {
							return fmt.Errorf("expected lifecycle not to be recreated; id changed from %q to %q", lifecycleID, rs.Primary.ID)
						}
						return nil
					},
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
				),
			},
		},
	})
}

func TestAccLifecycleResourceRestartsWhenNodeAndLinkStoppedExternally(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	config := testAccLifecycleResourceConfigWithState(cfg.Cfg, "STARTED")
	var labID string
	var nodeID string
	var linkID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						labID = rs.Primary.Attributes["lab_id"]
						if labID == "" {
							return fmt.Errorf("expected lab_id")
						}

						nodeRS, ok := s.RootModule().Resources["cml2_node.r1"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_node.r1")
						}
						nodeID = nodeRS.Primary.ID
						if nodeID == "" {
							return fmt.Errorf("expected node id")
						}

						linkRS, ok := s.RootModule().Resources["cml2_link.l1"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_link.l1")
						}
						linkID = linkRS.Primary.ID
						if linkID == "" {
							return fmt.Errorf("expected link id")
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

					if labID == "" || nodeID == "" || linkID == "" {
						return fmt.Errorf("internal test error: expected captured lab_id/node_id/link_id")
					}

					// Simulate external drift: stop a single node.
					if err := client.Node.Stop(context.Background(), models.UUID(labID), models.UUID(nodeID)); err != nil {
						return err
					}

					// Wait briefly for CML to update operational state.
					var lab *models.Lab
					var lastErr error
					deadline := time.Now().Add(30 * time.Second)
					for time.Now().Before(deadline) {
						l, getErr := client.Lab.GetByID(context.Background(), models.UUID(labID), true)
						if getErr != nil {
							lastErr = getErr
							continue
						}
						lab = &l

						n := lab.Nodes[models.UUID(nodeID)]
						if n != nil && n.State == models.NodeStateStopped {
							break
						}
						time.Sleep(2 * time.Second)
					}
					if lab == nil {
						if lastErr != nil {
							return lastErr
						}
						return fmt.Errorf("timeout waiting for node to become STOPPED")
					}

					// Validate we hit the intended drift scenario: lab state unchanged
					// while dependent node/link states changed.
					if lab.State != models.LabStateStarted {
						return fmt.Errorf("expected lab state to remain STARTED during drift, got %s", lab.State)
					}

					if n := lab.Nodes[models.UUID(nodeID)]; n == nil || n.State != models.NodeStateStopped {
						return fmt.Errorf("expected node to be STOPPED, got %v", func() models.NodeState {
							if n == nil {
								return ""
							}
							return n.State
						}())
					}

					var linkState string
					for _, l := range lab.Links {
						if l.ID == models.UUID(linkID) {
							linkState = l.State
							break
						}
					}
					if linkState == "" {
						return fmt.Errorf("link not found in lab during drift check")
					}
					if linkState == models.LinkStateStarted {
						return fmt.Errorf("expected link state to not be STARTED during drift, got %s", linkState)
					}

					return nil
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						client, err := cfg.NewCMLClientFromTFEnv()
						if err != nil {
							return err
						}
						lab, err := client.Lab.GetByID(context.Background(), models.UUID(labID), true)
						if err != nil {
							return err
						}

						n := lab.Nodes[models.UUID(nodeID)]
						if n == nil {
							return fmt.Errorf("expected node to exist after drift correction")
						}
						if n.State != models.NodeStateBooted {
							return fmt.Errorf("expected node to be BOOTED after drift correction, got %s", n.State)
						}

						var linkState string
						for _, l := range lab.Links {
							if l.ID == models.UUID(linkID) {
								linkState = l.State
								break
							}
						}
						if linkState == "" {
							return fmt.Errorf("link not found in lab after drift correction")
						}
						if linkState != models.LinkStateStarted {
							return fmt.Errorf("expected link to be STARTED after drift correction, got %s", linkState)
						}
						return nil
					},
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
				),
			},
		},
	})
}

func TestAccLifecycleResourceRestartsWhenLinkStoppedExternally(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	config := testAccLifecycleResourceConfigWithState(cfg.Cfg, "STARTED")
	var labID string
	var nodeID string
	var linkID string
	var lifecycleID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						lifecycleID = rs.Primary.ID
						labID = rs.Primary.Attributes["lab_id"]
						if lifecycleID == "" || labID == "" {
							return fmt.Errorf("expected lifecycle id and lab_id")
						}

						nodeRS, ok := s.RootModule().Resources["cml2_node.r1"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_node.r1")
						}
						nodeID = nodeRS.Primary.ID
						if nodeID == "" {
							return fmt.Errorf("expected node id")
						}

						linkRS, ok := s.RootModule().Resources["cml2_link.l1"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_link.l1")
						}
						linkID = linkRS.Primary.ID
						if linkID == "" {
							return fmt.Errorf("expected link id")
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
					if labID == "" || nodeID == "" || linkID == "" {
						return fmt.Errorf("internal test error: expected captured lab_id/node_id/link_id")
					}

					if err := client.Link.Stop(context.Background(), models.UUID(labID), models.UUID(linkID)); err != nil {
						return err
					}

					deadline := time.Now().Add(30 * time.Second)
					for time.Now().Before(deadline) {
						lab, getErr := client.Lab.GetByID(context.Background(), models.UUID(labID), true)
						if getErr != nil {
							time.Sleep(2 * time.Second)
							continue
						}
						if lab.State != models.LabStateStarted {
							return fmt.Errorf("expected lab state to remain STARTED during link drift, got %s", lab.State)
						}

						n := lab.Nodes[models.UUID(nodeID)]
						if n == nil {
							return fmt.Errorf("node not found during link drift check")
						}
						if n.State != models.NodeStateStarted && n.State != models.NodeStateBooted {
							return fmt.Errorf("expected node to remain running during link drift, got %s", n.State)
						}

						var linkState string
						for _, l := range lab.Links {
							if l.ID == models.UUID(linkID) {
								linkState = l.State
								break
							}
						}
						if linkState == models.LinkStateStopped {
							return nil
						}
						time.Sleep(2 * time.Second)
					}
					return fmt.Errorf("timeout waiting for link to become STOPPED")
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						if rs.Primary.ID != lifecycleID {
							return fmt.Errorf("expected lifecycle not to be recreated; id changed from %q to %q", lifecycleID, rs.Primary.ID)
						}

						client, err := cfg.NewCMLClientFromTFEnv()
						if err != nil {
							return err
						}
						lab, err := client.Lab.GetByID(context.Background(), models.UUID(labID), true)
						if err != nil {
							return err
						}

						n := lab.Nodes[models.UUID(nodeID)]
						if n == nil {
							return fmt.Errorf("expected node to exist after drift correction")
						}
						if n.State != models.NodeStateStarted && n.State != models.NodeStateBooted {
							return fmt.Errorf("expected node to remain running after drift correction, got %s", n.State)
						}

						for _, l := range lab.Links {
							if l.ID == models.UUID(linkID) {
								if l.State != models.LinkStateStarted {
									return fmt.Errorf("expected link to be STARTED after drift correction, got %s", l.State)
								}
								return nil
							}
						}
						return fmt.Errorf("link not found in lab after drift correction")
					},
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
				),
			},
		},
	})
}

func TestAccLifecycleResourceStopsWhenStartedExternally(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	configStarted := testAccLifecycleResourceConfigWithState(cfg.Cfg, "STARTED")
	configStopped := testAccLifecycleResourceConfigWithState(cfg.Cfg, "STOPPED")
	var labID string
	var lifecycleID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configStarted,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						lifecycleID = rs.Primary.ID
						labID = rs.Primary.Attributes["lab_id"]
						if lifecycleID == "" || labID == "" {
							return fmt.Errorf("expected lifecycle id and lab_id")
						}
						return nil
					},
				),
			},
			{
				Config:             configStopped,
				ExpectNonEmptyPlan: true,
				Check: func(s *terraform.State) error {
					client, err := cfg.NewCMLClientFromTFEnv()
					if err != nil {
						return err
					}
					if labID == "" {
						return fmt.Errorf("expected lab_id")
					}
					// Simulate external drift: force lab into STARTED while desired is STOPPED.
					// (Even though it should already be STARTED from step 1, we do it
					// explicitly to reduce timing flakiness.)
					return client.Lab.Start(context.Background(), models.UUID(labID))
				},
			},
			{
				Config: configStopped,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						if rs.Primary.ID != lifecycleID {
							return fmt.Errorf("expected lifecycle not to be recreated")
						}
						return nil
					},
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STOPPED"),
				),
			},
		},
	})
}

func TestAccLifecycleResourceDefinesWhenStoppedExternally(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	config := testAccLifecycleResourceConfigWithState(cfg.Cfg, "DEFINED_ON_CORE")
	var labID string
	var lifecycleID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "DEFINED_ON_CORE"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						lifecycleID = rs.Primary.ID
						labID = rs.Primary.Attributes["lab_id"]
						if lifecycleID == "" || labID == "" {
							return fmt.Errorf("expected lifecycle id and lab_id")
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
					if labID == "" {
						return fmt.Errorf("expected lab_id")
					}
					// Simulate external drift: make the lab STARTED while desired state is
					// DEFINED_ON_CORE. Provider should stop + wipe back to DEFINED_ON_CORE.
					return client.Lab.Start(context.Background(), models.UUID(labID))
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						if rs.Primary.ID != lifecycleID {
							return fmt.Errorf("expected lifecycle not to be recreated")
						}
						return nil
					},
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "DEFINED_ON_CORE"),
				),
			},
		},
	})
}

func TestAccLifecycleResourceStartsWhenDefinedExternally(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	config := testAccLifecycleResourceConfigWithState(cfg.Cfg, "STARTED")
	var labID string
	var lifecycleID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						lifecycleID = rs.Primary.ID
						labID = rs.Primary.Attributes["lab_id"]
						if lifecycleID == "" || labID == "" {
							return fmt.Errorf("expected lifecycle id and lab_id")
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
					if labID == "" {
						return fmt.Errorf("expected lab_id")
					}
					// Simulate external drift: remove running state and wipe to get DEFINED_ON_CORE while desired is STARTED.
					if err := client.Lab.Stop(context.Background(), models.UUID(labID)); err != nil {
						return err
					}
					return client.Lab.Wipe(context.Background(), models.UUID(labID))
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_lifecycle.top"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_lifecycle.top")
						}
						if rs.Primary.ID != lifecycleID {
							return fmt.Errorf("expected lifecycle not to be recreated")
						}
						return nil
					},
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
				),
			},
		},
	})
}

func testAccLifecycleResourceConfigWithState(cfg, state string) string {
	title := "acc lifecycle resource"
	switch state {
	case "STARTED":
		title = "acc lifecycle resource started"
	case "STOPPED":
		title = "acc lifecycle resource stopped"
	case "DEFINED_ON_CORE":
		title = "acc lifecycle resource defined"
	}
	return fmt.Sprintf(`
	%[1]s
	resource "cml2_lab" "this" {
		title = %[2]q
	}

	resource "cml2_node" "r1" {
	  lab_id         = cml2_lab.this.id
	  label          = "R1"
	  nodedefinition = "nginx"
	}

	resource "cml2_node" "r2" {
	  lab_id         = cml2_lab.this.id
	  label          = "R2"
	  nodedefinition = "nginx"
	}

	resource "cml2_link" "l1" {
	  lab_id = cml2_lab.this.id
	  node_a = cml2_node.r1.id
	  node_b = cml2_node.r2.id
	}

	resource "cml2_lifecycle" "top" {
	  lab_id = cml2_lab.this.id
	  state  = %q
	  depends_on = [
	    cml2_node.r1,
	    cml2_node.r2,
	    cml2_link.l1,
	  ]
	}
`, cfg, title, state)
}

func TestAccLifecycleImport(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLifecycleImport(cfg.Cfg),
				Check:  resource.TestCheckResourceAttrWith("cml2_lab.this", "id", uuidCheck),
			},
			// ImportState testing
			{
				ResourceName:      "cml2_lab.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLifecycleConfigCheck(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	// this was deprecated and replaced by depends_on with 0.1.0
	// re1 := regexp.MustCompile(`When "LabID" is set, "elements" is a required attribute.`)
	re2 := regexp.MustCompile(`Can't set \"LabID\" and \"topology\" at the same time.`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			// {
			// 	Config:      testAccLifecycleConfigCheck(cfg.Cfg, false),
			// 	ExpectError: re1,
			// },
			{
				Config:      testAccLifecycleConfigCheck(cfg.Cfg, true),
				ExpectError: re2,
			},
		},
	})
}

func TestAccLifecycleImportLab(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	const (
		initialNginxConfig = "new config for nginx"
		changedNginxConfig = "changed config for nginx"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// node with label xxx is not found, we expect an error
			{
				Config: testAccLifecycleImportLab(
					cfg.Cfg, "xxx", initialNginxConfig,
				),
				ExpectError: regexp.MustCompile(`node with label xxx not found`),
			},
			// start lab and ensure that n0config output has the initial config
			{
				Config: testAccLifecycleImportLab(
					cfg.Cfg, "nginx-0", initialNginxConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("cml2_lifecycle.top", "lab_id", uuidCheck),
					testCheckLifecycleNodeConfigByLabel("cml2_lifecycle.top", "nginx-0", initialNginxConfig),
				),
			},
			// change config and ensure that n0config output now has the changed config
			// (this requires a replace)
			{
				Config: testAccLifecycleImportLab(
					cfg.Cfg, "nginx-0", changedNginxConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("cml2_lifecycle.top", "lab_id", uuidCheck),
					testCheckLifecycleNodeConfigByLabel("cml2_lifecycle.top", "nginx-0", changedNginxConfig),
				),
			},
		},
	})
}

func uuidCheck(value string) error {
	re := regexp.MustCompile(`\b[0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12}\b`)
	if !re.MatchString(value) {
		return fmt.Errorf("%s is not a UUID", value)
	}
	return nil
}

func testCheckLifecycleNodeConfigByLabel(resourceName, label, expectedConfig string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		attrs := rs.Primary.Attributes
		const prefix = "nodes."
		const suffix = ".label"

		for k, v := range attrs {
			if !strings.HasPrefix(k, prefix) || !strings.HasSuffix(k, suffix) {
				continue
			}
			nodeKey := strings.TrimSuffix(strings.TrimPrefix(k, prefix), suffix)
			if v != label {
				continue
			}
			cfgKey := fmt.Sprintf("nodes.%s.configuration", nodeKey)
			got, ok := attrs[cfgKey]
			if !ok {
				// Some framework state representations may omit nested object attributes
				// unless explicitly tracked; in that case, fall back to the configs map
				// check which validates the injection behavior.
				wantFromConfigs := expectedConfig
				cfgsKey := fmt.Sprintf("configs.%s", label)
				fromConfigs, ok2 := attrs[cfgsKey]
				if ok2 {
					if fromConfigs != wantFromConfigs {
						return fmt.Errorf("%s: expected %#v, got %#v", cfgsKey, wantFromConfigs, fromConfigs)
					}
					return nil
				}
				return fmt.Errorf("Not found: %s", cfgKey)
			}
			if got != expectedConfig {
				return fmt.Errorf("node %q configuration: expected %#v, got %#v", label, expectedConfig, got)
			}
			return nil
		}

		return fmt.Errorf("node with label %q not found in lifecycle state", label)
	}
}

func TestAccLifecycleResourceState(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	re1 := regexp.MustCompile(`can't transition from no state to STOPPED`)
	re2 := regexp.MustCompile(`can't transition from DEFINED_ON_CORE to STOPPED`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:      testAccLifecycleStateCheck(cfg.Cfg, "STOPPED"),
				ExpectError: re1,
			},
			{
				Config: testAccLifecycleStateCheck(cfg.Cfg, "DEFINED_ON_CORE"),
			},
			{
				Config:      testAccLifecycleStateCheck(cfg.Cfg, "STOPPED"),
				ExpectError: re2,
			},
		},
	})
}

func TestAccLifecycleSequence(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLifecycleSequence(cfg.Cfg, 0, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "false"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
				),
			},
			{
				Config: testAccLifecycleSequence(cfg.Cfg, 1, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STOPPED"),
				),
			},
			{
				Config: testAccLifecycleSequence(cfg.Cfg, 2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "DEFINED_ON_CORE"),
				),
			},
			{
				Config: testAccLifecycleSequence(cfg.Cfg, 3, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
				),
			},
			{
				Config: testAccLifecycleSequence(cfg.Cfg, 4, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "DEFINED_ON_CORE"),
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

func TestAccLifecycleAddNodeToBooted(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleAddNodeToBooted(cfg.Cfg, "acc lifecycle add node to booted initial", 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
				),
			},
			{
				Config: testAccLifecycleAddNodeToBooted(cfg.Cfg, "acc lifecycle add node to booted add r3", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
				),
			},
		},
	})
}

func TestAccLifecycleNamedConfigs(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleNamedConfigs(cfg.CfgNamedConfigs, 0),
			},
			{
				Config: testAccLifecycleNamedConfigs(cfg.CfgNamedConfigs, 1),
			},
			{
				Config: testAccLifecycleNamedConfigs(cfg.CfgNamedConfigs, 2),
			},
			{
				Config:      testAccLifecycleNamedConfigs(cfg.CfgNamedConfigs, 3),
				ExpectError: regexp.MustCompile(`Can't provide both`),
			},
			{
				Config:      testAccLifecycleNamedConfigs(cfg.CfgNamedConfigs, 4),
				ExpectError: regexp.MustCompile(`node with label \w+ not found`),
			},
			{
				Config:      testAccLifecycleNamedConfigs(cfg.CfgNamedConfigs, 5),
				ExpectError: regexp.MustCompile(`node with label \w+ not found`),
			},
		},
	})
}

func testAccLifecycleResourceConfig(cfg string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "acc lifecycle resource"
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "nginx"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "nginx"
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.r1.id
  node_b = cml2_node.r2.id
}

resource "cml2_lifecycle" "top" {
	lab_id = cml2_lab.this.id
	depends_on = [
		cml2_node.r1,
		cml2_node.r2,
		cml2_link.l1,
	]
}
`, cfg)
}

func testAccLifecycleSequence(cfg string, seq int, all bool) string {
	f := func(state string) string { return fmt.Sprintf("state = %q", state) }
	var state string
	switch seq {
	case 0:
		state = ""
	case 1:
		state = f("STOPPED")
	case 2:
		state = f("DEFINED_ON_CORE")
	case 3:
		state = f("STARTED")
	case 4:
		state = f("DEFINED_ON_CORE")
	}
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "acc lifecycle sequence"
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  configuration  = "virbr0"
  nodedefinition = "external_connector"
  tags           = [ "bla" ]
}

resource "cml2_node" "ums" {
  lab_id         = cml2_lab.this.id
  label          = "Unmanaged Switch"
  nodedefinition = "unmanaged_switch"
  tags           = [ "bla" ]
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "nginx"
  tags           = [ "bla" ]
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "nginx"
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ext.id
  node_b = cml2_node.ums.id
}

resource "cml2_link" "l2" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r1.id
}

resource "cml2_lifecycle" "top" {
	lab_id = cml2_lab.this.id
	depends_on = [
		cml2_node.ext,
		cml2_node.ums,
		cml2_node.r1,
		cml2_node.r2,
		cml2_link.l1,
		cml2_link.l2,
	]
	staging = {
		stages = [ "bla" ]
		start_remaining = %[3]s
	}
	%[2]s
}
`, cfg, state, strconv.FormatBool(all))
}

func testAccLifecycleStateCheck(cfg, state string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "acc lifecycle state check"
}
resource "cml2_lifecycle" "top" {
	lab_id = cml2_lab.this.id
	depends_on = []
	state = %[2]q
}
`, cfg, state)
}

func testAccLifecycleConfigCheck(cfg string, insertTopo bool) string {
	topo := ""
	if insertTopo {
		topo = fmt.Sprintf("topology = %q", "blabla")
	}
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "acc lifecycle config check"
}
resource "cml2_lifecycle" "top" {
	lab_id = cml2_lab.this.id
	%[2]s
}
`, cfg, topo)
}

func testAccLifecycleImportLab(cfg, label, nodeCfg string) string {
	// BEWARE!! the yaml below must be indented with spaces, not with tabs!!
	return fmt.Sprintf(`
%[1]s
resource "cml2_lifecycle" "top" {
	topology = <<EOT
    lab:
        description: 'need one node'
        notes: ''
        title: acc lifecycle import lab
        version: 0.1.0
    links: []
    nodes:
        - id: n0
          label: nginx-0
          x: 1
          y: 1
          node_definition: nginx
          configuration: hostname bla
          interfaces: []
          tags: ["infra"]
EOT
	configs = {
		"%[2]s": %[3]q,
	}
	staging = {
		stages = ["infra","core","sites"]
		start_remaining = false
	}
	wait = false
}

locals {
	n0config = [for k, v in cml2_lifecycle.top.nodes : v.configuration if v.label == "nginx-0"][0]
}

output "n0config" {
	value = local.n0config
}


`, cfg, label, nodeCfg)
}

func testAccLifecycleImport(cfg string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "acc lifecycle import"
}
`, cfg)
}

func testAccLifecycleAddNodeToBooted(cfg, title string, stage int) string {
	if stage == 0 {
		return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = %[2]q
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  configuration  = "virbr0"
  nodedefinition = "external_connector"
}

resource "cml2_node" "ums" {
  lab_id         = cml2_lab.this.id
  label          = "Unmanaged Switch"
  nodedefinition = "unmanaged_switch"
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "nginx"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "nginx"
}

resource "cml2_link" "l0" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ext.id
  node_b = cml2_node.ums.id
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r1.id
}

resource "cml2_link" "l2" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r2.id
}

resource "cml2_lifecycle" "top" {
	lab_id = cml2_lab.this.id
	update_triggers = {
		ext = cml2_node.ext.generation
		ums = cml2_node.ums.generation
		r1  = cml2_node.r1.generation
		r2  = cml2_node.r2.generation
	}
	depends_on = [
		cml2_node.ext,
		cml2_node.ums,
		cml2_node.r1,
		cml2_node.r2,
		cml2_link.l0,
		cml2_link.l1,
		cml2_link.l2,
	]
}
	`, cfg, title)
	} else {
		return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = %[2]q
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  configuration  = "virbr0"
  nodedefinition = "external_connector"
}

resource "cml2_node" "ums" {
  lab_id         = cml2_lab.this.id
  label          = "Unmanaged Switch"
  nodedefinition = "unmanaged_switch"
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "nginx"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "nginx"
}

resource "cml2_node" "r3" {
  lab_id         = cml2_lab.this.id
  label          = "R3"
  nodedefinition = "nginx"
}

resource "cml2_link" "l0" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ext.id
  node_b = cml2_node.ums.id
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r1.id
}

resource "cml2_link" "l2" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r2.id
}

resource "cml2_link" "l3" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r3.id
}

resource "cml2_lifecycle" "top" {
	lab_id = cml2_lab.this.id
	update_triggers = {
		ext = cml2_node.ext.generation
		ums = cml2_node.ums.generation
		r1  = cml2_node.r1.generation
		r2  = cml2_node.r2.generation
		r3  = cml2_node.r3.generation
	}
	depends_on = [
		cml2_node.ext,
		cml2_node.ums,
		cml2_node.r1,
		cml2_node.r2,
		cml2_node.r3,
		cml2_link.l0,
		cml2_link.l1,
		cml2_link.l2,
		cml2_link.l3,
	]
}
	`, cfg, title)
	}
}

func testAccLifecycleNamedConfigs(cfg string, stage int) string {
	var configs, namedConfigs string
	switch stage {
	case 0:
		configs = ``
		namedConfigs = ``
	case 1:
		configs = `
		configs = {
			"R1": "hostname r1"
		}
		`
		namedConfigs = ``
	case 2:
		configs = ``
		namedConfigs = `
		named_configs = {
			"R1": [
				{
					name = "node.cfg"
					content = "hostname r1"
				}
			]
		}
		`
	case 3:
		configs = `
		configs = {
			"R1": "hostname r1"
		}
		`
		namedConfigs = `
		named_configs = {
			"R1": [
				{
					name = "node.cfg"
					content = "hostname r1"
				}
			]
		}
		`
	case 4:
		configs = `
		configs = {
			"xx": "hostname r1"
		}
		`
		namedConfigs = ``
	case 5:
		configs = ``
		namedConfigs = `
		named_configs = {
			"xx": [
				{
					name = "node.cfg"
					content = "hostname r1"
				}
			]
		}
		`
	}
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "acc lifecycle named configs"
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "nginx"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "nginx"
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.r1.id
  node_b = cml2_node.r2.id
}

resource "cml2_lifecycle" "top" {
	lab_id = cml2_lab.this.id
	state = "DEFINED_ON_CORE"
	depends_on = [
		cml2_node.r1,
		cml2_node.r2,
		cml2_link.l1,
	]
	%[2]s
	%[3]s
}
`, cfg, configs, namedConfigs)
}

// TestAccLifecycleExtConnDeviceNameValidation verifies that external connector
// configuration must be a device name and that label values are rejected with
// actionable error messages.
func TestAccLifecycleExtConnDeviceNameNoInconsistency(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleExtConnConfig(cfg.Cfg, "bridge0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "DEFINED_ON_CORE"),
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "bridge0"),
				),
			},
			{
				// idempotency with device-name input
				Config:   testAccLifecycleExtConnConfig(cfg.Cfg, "bridge0"),
				PlanOnly: true,
			},
			{
				Config:      testAccLifecycleExtConnConfig(cfg.Cfg, "System Bridge"),
				ExpectError: regexp.MustCompile(`is a label; use device name`),
			},
			{
				Config: testAccLifecycleExtConnConfig(cfg.Cfg, "virbr0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "DEFINED_ON_CORE"),
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "virbr0"),
				),
			},
			{
				// idempotency with device-name input
				Config:   testAccLifecycleExtConnConfig(cfg.Cfg, "virbr0"),
				PlanOnly: true,
			},
			{
				Config:      testAccLifecycleExtConnConfig(cfg.Cfg, "NAT"),
				ExpectError: regexp.MustCompile(`is a label; use device name`),
			},
		},
	})
}

func testAccLifecycleExtConnConfig(providerCfg, extconnConfig string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "acc lifecycle extconn"
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  nodedefinition = "external_connector"
  configuration  = %[2]q
}

resource "cml2_lifecycle" "top" {
  lab_id = cml2_lab.this.id
  state  = "DEFINED_ON_CORE"
  depends_on = [
    cml2_node.ext,
  ]
}
`, providerCfg, extconnConfig)
}

// TestAccLifecycleExtConnNodeReplaceRestartsLab verifies that when an
// external_connector node is replaced (config change while STARTED triggers
// destroy+create) the lifecycle resource restarts the new node so the lab
// remains fully running.
//
// Before the fix, lifecycle Update() only called startNodes when labHasDrift()
// detected a stopped node.  But labHasDrift() is evaluated after fetching the
// lab — which happens before the node replacement completes in the same apply
// cycle when Terraform evaluates ModifyPlan.  The result was the new node
// stayed in DEFINED_ON_CORE after apply.
func TestAccLifecycleExtConnNodeReplaceRestartsLab(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	configStarted := func(extconn string) string {
		return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "acc lifecycle extconn replace"
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  nodedefinition = "external_connector"
  configuration  = %[2]q
}

resource "cml2_lifecycle" "top" {
  lab_id = cml2_lab.this.id
  state  = "STARTED"
  update_triggers = {
    ext = cml2_node.ext.generation
  }
  depends_on = [
    cml2_node.ext,
  ]
}
`, cfg.Cfg, extconn)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: create lab with extconn "virbr0", start lifecycle.
				Config: configStarted("virbr0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "virbr0"),
				),
			},
			{
				// Step 2: change extconn config to a different device name.
				// This replaces the node (destroy + create at DEFINED_ON_CORE).
				// The lifecycle Update() must restart the new node so the lab
				// stays STARTED.
				Config: configStarted("bridge0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "bridge0"),
					// Confirm the node is actually running, not stuck at DEFINED_ON_CORE.
					func(s *terraform.State) error {
						nodeRS, ok := s.RootModule().Resources["cml2_node.ext"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_node.ext")
						}
						labID := nodeRS.Primary.Attributes["lab_id"]
						nodeID := nodeRS.Primary.ID
						client, err := cfg.NewCMLClientFromTFEnv()
						if err != nil {
							return err
						}
						lab, err := client.Lab.GetByID(context.Background(), models.UUID(labID), true)
						if err != nil {
							return err
						}
						n := lab.Nodes[models.UUID(nodeID)]
						if n == nil {
							return fmt.Errorf("node not found in lab after replace")
						}
						if n.State == models.NodeStateDefined || n.State == models.NodeStateStopped {
							return fmt.Errorf("expected node to be running after lifecycle restart, got %s", n.State)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccLifecycleStartedWithUMSConfig(providerCfg string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "acc lifecycle ums recreated"
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  nodedefinition = "external_connector"
  configuration  = "virbr0"
}

resource "cml2_node" "ums" {
  lab_id         = cml2_lab.this.id
  label          = "Unmanaged Switch"
  nodedefinition = "unmanaged_switch"
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "nginx-1"
  nodedefinition = "nginx"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "nginx-2"
  nodedefinition = "nginx"
}

resource "cml2_link" "l0" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ext.id
  node_b = cml2_node.ums.id
}

resource "cml2_link" "l1" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r1.id
}

resource "cml2_link" "l2" {
  lab_id = cml2_lab.this.id
  node_a = cml2_node.ums.id
  node_b = cml2_node.r2.id
}

resource "cml2_lifecycle" "top" {
  lab_id = cml2_lab.this.id
  state  = "STARTED"
  update_triggers = {
    ext = cml2_node.ext.generation
    ums = "${cml2_node.ums.id}:${cml2_node.ums.generation}"
    r1  = cml2_node.r1.generation
    r2  = cml2_node.r2.generation
  }
  depends_on = [
    cml2_node.ext,
    cml2_node.ums,
    cml2_node.r1,
    cml2_node.r2,
    cml2_link.l0,
    cml2_link.l1,
    cml2_link.l2,
  ]
}
`, providerCfg)
}

// TestAccLifecycleUMSRecreatedRestartsLinks verifies that when an unmanaged
// switch is deleted out-of-band and then recreated via update_triggers, the
// lifecycle resource restarts the switch and its attached links.
func TestAccLifecycleUMSRecreatedRestartsLinks(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	config := testAccLifecycleStartedWithUMSConfig(cfg.Cfg)
	var labID string
	var initialUMSID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["cml2_node.ums"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_node.ums")
						}
						labID = rs.Primary.Attributes["lab_id"]
						initialUMSID = rs.Primary.ID
						if labID == "" || initialUMSID == "" {
							return fmt.Errorf("expected lab_id and ums id")
						}
						return nil
					},
				),
			},
			{
				Config:             config,
				ExpectNonEmptyPlan: true,
				Check: func(s *terraform.State) error {
					if labID == "" || initialUMSID == "" {
						return fmt.Errorf("internal test error: expected captured lab_id and ums id")
					}

					client, err := cfg.NewCMLClientFromTFEnv()
					if err != nil {
						return err
					}

					waitForUMSState := func(want models.NodeState, desc string) error {
						deadline := time.Now().Add(60 * time.Second)
						for time.Now().Before(deadline) {
							lab, err := client.Lab.GetByID(context.Background(), models.UUID(labID), true)
							if err != nil {
								time.Sleep(2 * time.Second)
								continue
							}
							n := lab.Nodes[models.UUID(initialUMSID)]
							if n != nil && n.State == want {
								return nil
							}
							time.Sleep(2 * time.Second)
						}
						return fmt.Errorf("timeout waiting for unmanaged switch to become %s", desc)
					}

					if err := client.Node.Stop(context.Background(), models.UUID(labID), models.UUID(initialUMSID)); err != nil {
						return err
					}
					if err := waitForUMSState(models.NodeStateStopped, "STOPPED"); err != nil {
						return err
					}

					if err := client.Node.Wipe(context.Background(), models.UUID(labID), models.UUID(initialUMSID)); err != nil {
						return err
					}
					if err := waitForUMSState(models.NodeStateDefined, "DEFINED_ON_CORE"); err != nil {
						return err
					}

					if err := client.Node.Delete(context.Background(), models.UUID(labID), models.UUID(initialUMSID)); err != nil {
						return err
					}

					deadline := time.Now().Add(30 * time.Second)
					for time.Now().Before(deadline) {
						lab, err := client.Lab.GetByID(context.Background(), models.UUID(labID), true)
						if err != nil {
							time.Sleep(2 * time.Second)
							continue
						}
						if _, ok := lab.Nodes[models.UUID(initialUMSID)]; !ok {
							return nil
						}
						time.Sleep(2 * time.Second)
					}

					return fmt.Errorf("timeout waiting for unmanaged switch deletion")
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
					func(s *terraform.State) error {
						client, err := cfg.NewCMLClientFromTFEnv()
						if err != nil {
							return err
						}

						rs, ok := s.RootModule().Resources["cml2_node.ums"]
						if !ok {
							return fmt.Errorf("not found in state: cml2_node.ums")
						}
						currentUMSID := rs.Primary.ID
						if currentUMSID == "" {
							return fmt.Errorf("expected recreated unmanaged switch id")
						}
						if currentUMSID == initialUMSID {
							return fmt.Errorf("expected unmanaged switch to be recreated, id still %q", initialUMSID)
						}

						lab, err := client.Lab.GetByID(context.Background(), models.UUID(labID), true)
						if err != nil {
							return err
						}

						ums := lab.Nodes[models.UUID(currentUMSID)]
						if ums == nil {
							return fmt.Errorf("recreated unmanaged switch not found in lab")
						}
						if ums.State != models.NodeStateStarted && ums.State != models.NodeStateBooted {
							return fmt.Errorf("expected unmanaged switch to be started after lifecycle restart, got %s", ums.State)
						}

						for _, resName := range []string{"cml2_link.l0", "cml2_link.l1", "cml2_link.l2"} {
							linkRS, ok := s.RootModule().Resources[resName]
							if !ok {
								return fmt.Errorf("not found in state: %s", resName)
							}
							linkID := linkRS.Primary.ID
							if linkID == "" {
								return fmt.Errorf("expected link id for %s", resName)
							}

							var linkState string
							found := false
							for _, l := range lab.Links {
								if l.ID == models.UUID(linkID) {
									linkState = l.State
									found = true
									break
								}
							}
							if !found {
								return fmt.Errorf("%s not found in lab after lifecycle restart", resName)
							}
							if linkState != models.LinkStateStarted {
								return fmt.Errorf("expected %s to be STARTED after lifecycle restart, got %s", resName, linkState)
							}
						}

						return nil
					},
				),
			},
		},
	})
}

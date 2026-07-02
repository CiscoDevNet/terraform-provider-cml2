package lifecycle_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/rschmied/gocmlclient/pkg/models"

	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
)

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

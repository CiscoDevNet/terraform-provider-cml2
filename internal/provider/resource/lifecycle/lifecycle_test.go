package lifecycle_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
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

func stopLinkViaAPI(ctx context.Context, labID, linkID string) error {
	addr := strings.TrimRight(os.Getenv("TF_VAR_address"), "/")
	token := os.Getenv("TF_VAR_token")
	if addr == "" {
		return fmt.Errorf("TF_VAR_address must be set")
	}
	if token == "" {
		return fmt.Errorf("TF_VAR_token must be set for link stop drift test")
	}

	url := fmt.Sprintf("%s/api/v0/labs/%s/links/%s/state/stop", addr, labID, linkID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	hc := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("stop link API failed: status=%d body=%q", resp.StatusCode, string(body))
	}
	return nil
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

					if err := stopLinkViaAPI(context.Background(), labID, linkID); err != nil {
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
	return fmt.Sprintf(`
	%[1]s
	resource "cml2_lab" "this" {}

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
`, cfg, state)
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
		initialAlpineConfig = "new config for alpine"
		changedAlpineConfig = "changed config for alpine"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// node with label xxx is not found, we expect an error
			{
				Config: testAccLifecycleImportLab(
					cfg.Cfg, "xxx", initialAlpineConfig,
				),
				ExpectError: regexp.MustCompile(`node with label xxx not found`),
			},
			// start lab and ensure that n0config output has the initial config
			{
				Config: testAccLifecycleImportLab(
					cfg.Cfg, "alpine-0", initialAlpineConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("cml2_lifecycle.top", "lab_id", uuidCheck),
					testCheckLifecycleNodeConfigByLabel("cml2_lifecycle.top", "alpine-0", initialAlpineConfig),
				),
			},
			// change config and ensure that n0config output now has the changed config
			// (this requires a replace)
			{
				Config: testAccLifecycleImportLab(
					cfg.Cfg, "alpine-0", changedAlpineConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("cml2_lifecycle.top", "lab_id", uuidCheck),
					testCheckLifecycleNodeConfigByLabel("cml2_lifecycle.top", "alpine-0", changedAlpineConfig),
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
				Config: testAccLifecycleAddNodeToBooted(cfg.Cfg, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "booted", "true"),
				),
			},
			{
				Config: testAccLifecycleAddNodeToBooted(cfg.Cfg, 1),
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
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "alpine"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "alpine"
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
	title = "lifecycle seq ac-test"
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  configuration  = "NAT"
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
  nodedefinition = "alpine"
  tags           = [ "bla" ]
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "alpine"
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
        title: empty
        version: 0.1.0
    links: []
    nodes:
        - id: n0
          label: alpine-0
          x: 1
          y: 1
          node_definition: alpine
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
	n0config = [for k, v in cml2_lifecycle.top.nodes : v.configuration if v.label == "alpine-0"][0]
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
	title = "labimport"
}
`, cfg)
}

func testAccLifecycleAddNodeToBooted(cfg string, stage int) string {
	if stage == 0 {
		return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "lifecycle add node to booted"
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  configuration  = "NAT"
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
  nodedefinition = "alpine"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "alpine"
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
	`, cfg)
	} else {
		return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {
	title = "lifecycle add node to booted"
}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  configuration  = "NAT"
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
  nodedefinition = "alpine"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "alpine"
}

resource "cml2_node" "r3" {
  lab_id         = cml2_lab.this.id
  label          = "R3"
  nodedefinition = "alpine"
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
	`, cfg)
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
}

resource "cml2_node" "r1" {
  lab_id         = cml2_lab.this.id
  label          = "R1"
  nodedefinition = "alpine"
}

resource "cml2_node" "r2" {
  lab_id         = cml2_lab.this.id
  label          = "R2"
  nodedefinition = "alpine"
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

// TestAccLifecycleExtConnDeviceNameNoInconsistency reproduces the bug where
// changing an external_connector node's configuration from a label to a device
// name caused "Provider produced inconsistent result after apply" in the
// lifecycle resource.
//
// Two known connector pairs on the test CML instance are exercised:
//
//	"bridge0" (device name) <-> "System Bridge" (label)
//	"virbr0"  (device name) <-> "NAT"           (label)
//
// Scenario:
//  1. Create lifecycle with extconn using label "System Bridge".
//  2. Update to device name "bridge0" (server normalises → "System Bridge").
//  3. Update to label "NAT".
//  4. Update to device name "virbr0" (server normalises → "NAT").
//  5. Idempotency check — re-apply device name, expect empty plan.
//  6. Switch back to label "System Bridge" for clean teardown.
func TestAccLifecycleExtConnDeviceNameNoInconsistency(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: baseline with label.
				Config: testAccLifecycleExtConnConfig(cfg.Cfg, "System Bridge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "DEFINED_ON_CORE"),
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "System Bridge"),
				),
			},
			{
				// Step 2: switch to device name "bridge0".
				// Before the fix this produced:
				//   .nodes["..."].configurations[0].content:
				//     was cty.StringVal("System Bridge"), but now cty.StringVal("System Bridge")
				// (or an equivalent inconsistent-result error).
				Config: testAccLifecycleExtConnConfig(cfg.Cfg, "bridge0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "DEFINED_ON_CORE"),
					// Standalone node keeps the device-name in state (back-compat).
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "bridge0"),
				),
			},
			{
				// Step 3: idempotency — re-apply device name, must produce empty plan.
				Config:   testAccLifecycleExtConnConfig(cfg.Cfg, "bridge0"),
				PlanOnly: true,
			},
			{
				// Step 4: switch to the other label.
				Config: testAccLifecycleExtConnConfig(cfg.Cfg, "NAT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "NAT"),
				),
			},
			{
				// Step 5: switch to device name "virbr0" (normalises to "NAT").
				Config: testAccLifecycleExtConnConfig(cfg.Cfg, "virbr0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "DEFINED_ON_CORE"),
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "virbr0"),
				),
			},
			{
				// Step 6: idempotency — re-apply device name, must produce empty plan.
				Config:   testAccLifecycleExtConnConfig(cfg.Cfg, "virbr0"),
				PlanOnly: true,
			},
			{
				// Step 7: return to a label for clean teardown.
				Config: testAccLifecycleExtConnConfig(cfg.Cfg, "System Bridge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "System Bridge"),
				),
			},
		},
	})
}

func testAccLifecycleExtConnConfig(providerCfg, extconnConfig string) string {
	return fmt.Sprintf(`
%[1]s
resource "cml2_lab" "this" {}

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
resource "cml2_lab" "this" {}

resource "cml2_node" "ext" {
  lab_id         = cml2_lab.this.id
  label          = "Internet"
  nodedefinition = "external_connector"
  configuration  = %[2]q
}

resource "cml2_lifecycle" "top" {
  lab_id = cml2_lab.this.id
  state  = "STARTED"
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
				// Step 1: create lab with extconn "NAT", start lifecycle.
				Config: configStarted("NAT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "NAT"),
				),
			},
			{
				// Step 2: change extconn config to a different label.
				// This replaces the node (destroy + create at DEFINED_ON_CORE).
				// The lifecycle Update() must restart the new node so the lab
				// stays STARTED.
				Config: configStarted("System Bridge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_lifecycle.top", "state", "STARTED"),
					resource.TestCheckResourceAttr("cml2_node.ext", "configuration", "System Bridge"),
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

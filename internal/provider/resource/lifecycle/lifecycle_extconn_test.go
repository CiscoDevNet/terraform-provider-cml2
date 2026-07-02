package lifecycle_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/rschmied/gocmlclient/pkg/models"

	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
)

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

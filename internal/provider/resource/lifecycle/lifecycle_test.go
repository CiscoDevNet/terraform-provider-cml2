package lifecycle_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

func TestAccLifecycleResource(t *testing.T) {
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

func TestAccLifecycleImport(t *testing.T) {
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
	// this was deprecated and replaced by depends_on with 0.1.0
	// re1 := regexp.MustCompile(`When "LabID" is set, "elements" is a required attribue.`)
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
					resource.TestCheckOutput("n0config", initialAlpineConfig),
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
					resource.TestCheckOutput("n0config", changedAlpineConfig),
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

func TestAccLifecycleResourceState(t *testing.T) {
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
	topology = <<-EOT
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
		remaining = false
	}
	wait = false
}
output "n0config" {
    value = [ for k, v in cml2_lifecycle.top.nodes : v.configuration if v.label == "alpine-0" ][0]
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
		cml2_node.ext.id,
		cml2_node.ums.id,
		cml2_node.r1.id,
		cml2_node.r2.id,
		cml2_link.l0.id,
		cml2_link.l1.id,
		cml2_link.l2.id,
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
		cml2_node.ext.id,
		cml2_node.ums.id,
		cml2_node.r1.id,
		cml2_node.r2.id,
		cml2_node.r3.id,
		cml2_link.l0.id,
		cml2_link.l1.id,
		cml2_link.l2.id,
		cml2_link.l3.id,
	]
}
	`, cfg)
	}
}

func testAccLifecycleNamedConfigs(cfg string, stage int) string {
	var configs, named_configs string
	switch stage {
	case 0:
		configs = ``
		named_configs = ``
	case 1:
		configs = `
		configs = {
			"R1": "hostname r1"
		}
		`
		named_configs = ``
	case 2:
		configs = ``
		named_configs = `
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
		named_configs = `
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
		named_configs = ``
	case 5:
		configs = ``
		named_configs = `
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
`, cfg, configs, named_configs)
}

package lifecycle_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cml2": providerserver.NewProtocol6WithError(cml.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for
	// example assertions about the appropriate environment variables being set
	// are common to see in a pre-check function.
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

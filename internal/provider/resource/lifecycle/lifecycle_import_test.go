package lifecycle_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
)

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

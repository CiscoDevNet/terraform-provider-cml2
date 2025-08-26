package lab_test

import (
	"fmt"
	"regexp"
	"testing"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cml2": providerserver.NewProtocol6WithError(cml.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for
	// example assertions about the appropriate environment variables being set
	// are common to see in a pre-check function.
}

func TestLabDataSource(t *testing.T) {
	title := "thislab"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testLabDataSourceConfig2(cfg.Cfg, title),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("theoutput", title),
				),
			},
			{
				Config: testLabDataSourceConfig1(cfg.Cfg, title),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("theoutput", title),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testLabDataSourceConfig3(cfg.Cfg, title),
				ExpectError: regexp.MustCompile("need to provide either title"),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testLabDataSourceConfig4(cfg.Cfg, title),
				ExpectError: regexp.MustCompile("Unable to get lab, got error"),
			},
		},
	})
}

func testLabDataSourceConfig1(cfg, title string) string {
	return fmt.Sprintf(`
	%[1]s
	resource "cml2_lab" "test" {
			title = %[2]q
	}
	data "cml2_lab" "acc_test" {
		title = %[2]q
	}
	output "theoutput" {
		value = data.cml2_lab.acc_test.lab.title
	}
	`, cfg, title)
}

func testLabDataSourceConfig2(cfg, title string) string {
	return fmt.Sprintf(`
	%[1]s
	resource "cml2_lab" "test" {
			title = %[2]q
	}
	data "cml2_lab" "acc_test" {
		id = cml2_lab.test.id
	}
	output "theoutput" {
		value = data.cml2_lab.acc_test.lab.title
	}
	`, cfg, title)
}

func testLabDataSourceConfig3(cfg, title string) string {
	return fmt.Sprintf(`
	%[1]s
		# resource "cml2_lab" "test" {
		# 		title = %[2]q
		# }
	data "cml2_lab" "acc_test" {
	}
	`, cfg, title)
}

func testLabDataSourceConfig4(cfg, title string) string {
	return fmt.Sprintf(`
	%[1]s
	data "cml2_lab" "acc_test" {
		id = "3a86a9b9-7885-490b-b59d-4061519aad8c"
	}
	`, cfg, title)
}

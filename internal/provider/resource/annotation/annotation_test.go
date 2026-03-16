package annotation_test

import (
	"fmt"
	"testing"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var testAccAnnotationProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cml2": providerserver.NewProtocol6WithError(cml.New("test")()),
}

func TestAccAnnotationResourceText(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: testAccAnnotationProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnnotationText(cfg.Cfg, "hello"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "text"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "text.text_content", "hello"),
				),
			},
			{
				Config: testAccAnnotationText(cfg.Cfg, "hello2"),
				Check:  resource.TestCheckResourceAttr("cml2_annotation.a", "text.text_content", "hello2"),
			},
		},
	})
}

func testAccAnnotationText(cfg string, content string) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_lab" "l" {
	title = "acc annotation"
}

resource "cml2_annotation" "a" {
	lab_id = cml2_lab.l.id
	type   = "text"

	text = {
		text_content = %[2]q
		x1          = 10
		y1          = 20
	}
}
`, cfg, content)
}

func testAccAnnotationImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["cml2_annotation.a"]
	if !ok {
		return "", fmt.Errorf("resource not found")
	}
	labID := rs.Primary.Attributes["lab_id"]
	annID := rs.Primary.ID
	if labID == "" || annID == "" {
		return "", fmt.Errorf("missing lab_id or id")
	}
	return fmt.Sprintf("%s/%s", labID, annID), nil
}

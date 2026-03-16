package annotation_test

import (
	"fmt"
	"testing"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

func TestAccAnnotationResourceRectangle(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: testAccAnnotationProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnnotationRectangle(cfg.Cfg, 10, 20, 30, 40),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "rectangle"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "rectangle.x2", "30"),
				),
			},
			{
				Config: testAccAnnotationRectangle(cfg.Cfg, 10, 20, 35, 45),
				Check:  resource.TestCheckResourceAttr("cml2_annotation.a", "rectangle.x2", "35"),
			},
		},
	})
}

func TestAccAnnotationResourceEllipse(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: testAccAnnotationProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnnotationEllipse(cfg.Cfg, 10, 20, 30, 40),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "ellipse"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "ellipse.x2", "30"),
				),
			},
			{
				Config: testAccAnnotationEllipse(cfg.Cfg, 10, 20, 35, 45),
				Check:  resource.TestCheckResourceAttr("cml2_annotation.a", "ellipse.x2", "35"),
			},
		},
	})
}

func TestAccAnnotationResourceLine(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: testAccAnnotationProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnnotationLine(cfg.Cfg, 10, 20, 30, 40),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "line"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "line.x2", "30"),
				),
			},
			{
				Config: testAccAnnotationLine(cfg.Cfg, 10, 20, 35, 45),
				Check:  resource.TestCheckResourceAttr("cml2_annotation.a", "line.x2", "35"),
			},
		},
	})
}

func testAccAnnotationText(cfgStr string, content string) string {
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
`, cfgStr, content)
}

func testAccAnnotationRectangle(cfgStr string, x1, y1, x2, y2 int) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_lab" "l" {
	title = "acc annotation rectangle"
}

resource "cml2_annotation" "a" {
	lab_id = cml2_lab.l.id
	type   = "rectangle"

	rectangle = {
		x1 = %[2]d
		y1 = %[3]d
		x2 = %[4]d
		y2 = %[5]d
	}
}
`, cfgStr, x1, y1, x2, y2)
}

func testAccAnnotationEllipse(cfgStr string, x1, y1, x2, y2 int) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_lab" "l" {
	title = "acc annotation ellipse"
}

resource "cml2_annotation" "a" {
	lab_id = cml2_lab.l.id
	type   = "ellipse"

	ellipse = {
		x1 = %[2]d
		y1 = %[3]d
		x2 = %[4]d
		y2 = %[5]d
	}
}
`, cfgStr, x1, y1, x2, y2)
}

func testAccAnnotationLine(cfgStr string, x1, y1, x2, y2 int) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_lab" "l" {
	title = "acc annotation line"
}

resource "cml2_annotation" "a" {
	lab_id = cml2_lab.l.id
	type   = "line"

	line = {
		x1 = %[2]d
		y1 = %[3]d
		x2 = %[4]d
		y2 = %[5]d
		line_start = "arrow"
		line_end   = "arrow"
	}
}
`, cfgStr, x1, y1, x2, y2)
}

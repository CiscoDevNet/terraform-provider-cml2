package annotation_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	cml "github.com/ciscodevnet/terraform-provider-cml2/internal/provider"
	cfg "github.com/ciscodevnet/terraform-provider-cml2/internal/testing"
)

func testCheckAttrNullOrEmpty(resourceName, attrPath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		v, ok := rs.Primary.Attributes[attrPath]
		if !ok {
			// attribute not present => treat as null
			return nil
		}
		if v == "" {
			return nil
		}
		return fmt.Errorf("%s: expected null/empty, got %#v", attrPath, v)
	}
}

func testCheckAttrNotEqual(resourceName, attrPath, notValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		v, ok := rs.Primary.Attributes[attrPath]
		if !ok {
			return nil
		}
		if v == notValue {
			return fmt.Errorf("%s: expected not %#v, got %#v", attrPath, notValue, v)
		}
		return nil
	}
}

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
				Config: testAccAnnotationText(cfg.Cfg, "hello", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "text"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "text.text_content", "hello"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "text.text_font", "monospace"),
				),
			},
			{
				Config: testAccAnnotationText(cfg.Cfg, "hello2", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "text.text_content", "hello2"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "text.text_bold", "true"),
				),
			},
			{
				ResourceName:      "cml2_annotation.a",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["cml2_annotation.a"]
					if !ok {
						return "", fmt.Errorf("Not found: cml2_annotation.a")
					}
					labID, ok := rs.Primary.Attributes["lab_id"]
					if !ok {
						return "", fmt.Errorf("Not found: lab_id")
					}
					return fmt.Sprintf("%s/%s", labID, rs.Primary.ID), nil
				},
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
				Config: testAccAnnotationRectangle(cfg.Cfg, 10, 20, 30, 40, 0, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "rectangle"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "rectangle.x2", "30"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "rectangle.rotation", "0"),
				),
			},
			{
				Config: testAccAnnotationRectangle(cfg.Cfg, 10, 20, 35, 45, 45, "4,2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "rectangle.x2", "35"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "rectangle.rotation", "45"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "rectangle.border_style", "4,2"),
				),
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
				Config: testAccAnnotationEllipse(cfg.Cfg, 10, 20, 30, 40, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "ellipse"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "ellipse.x2", "30"),
				),
			},
			{
				Config: testAccAnnotationEllipse(cfg.Cfg, 10, 20, 35, 45, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "ellipse.x2", "35"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "ellipse.rotation", "30"),
				),
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
				Config: testAccAnnotationLine(cfg.Cfg, 10, 20, 30, 40, "arrow", "arrow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "line"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "line.x2", "30"),
				),
			},
			{
				Config: testAccAnnotationLine(cfg.Cfg, 10, 20, 35, 45, "square", "circle"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "line.x2", "35"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "line.line_start", "square"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "line.line_end", "circle"),
				),
			},
			{
				Config: testAccAnnotationLineNullMarkers(cfg.Cfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "line"),
					testCheckAttrNullOrEmpty("cml2_annotation.a", "line.line_start"),
					testCheckAttrNullOrEmpty("cml2_annotation.a", "line.line_end"),
					testCheckAttrNotEqual("cml2_annotation.a", "line.line_start", "arrow"),
					testCheckAttrNotEqual("cml2_annotation.a", "line.line_end", "arrow"),
				),
			},
			{
				Config: testAccAnnotationLineNullCombo(cfg.Cfg, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "line"),
					testCheckAttrNullOrEmpty("cml2_annotation.a", "line.line_start"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "line.line_end", "arrow"),
				),
			},
			{
				Config: testAccAnnotationLineNullCombo(cfg.Cfg, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cml2_annotation.a", "type", "line"),
					resource.TestCheckResourceAttr("cml2_annotation.a", "line.line_start", "arrow"),
					testCheckAttrNullOrEmpty("cml2_annotation.a", "line.line_end"),
				),
			},
			{
				Config:      testAccAnnotationLineInvalidMarker(cfg.Cfg),
				ExpectError: resourceConfigErrorRegex(),
			},
			// Ensure the final config is valid so post-test destroy can run.
			{
				Config: cfg.Cfg,
			},
		},
	})
}

func TestAccAnnotationResourceInvalidBorderStyle(t *testing.T) {
	cfg.SkipUnlessAcc(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: testAccAnnotationProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAnnotationTextInvalidBorderStyle(cfg.Cfg),
				ExpectError: resourceConfigErrorRegex(),
			},
			// Ensure the final config is valid so post-test destroy can run.
			{
				Config: cfg.Cfg,
			},
		},
	})
}

func resourceConfigErrorRegex() *regexp.Regexp {
	return regexp.MustCompile(`(?i)(invalid value|expected one of|one of)`)
}

func testAccAnnotationText(cfgStr, content string, bold bool) string {
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
		text_font   = "monospace"
		text_bold   = %[3]t
		rotation    = 10
		border_style = "2,2"
	}
}
`, cfgStr, content, bold)
}

func testAccAnnotationRectangle(cfgStr string, x1, y1, x2, y2 int, rotation int, borderStyle string) string {
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
		rotation = %[6]d
		border_style = %[7]q
		border_radius = 12
	}
}
`, cfgStr, x1, y1, x2, y2, rotation, borderStyle)
}

func testAccAnnotationEllipse(cfgStr string, x1, y1, x2, y2 int, rotation int) string {
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
		rotation = %[6]d
		border_style = ""
	}
}
`, cfgStr, x1, y1, x2, y2, rotation)
}

func testAccAnnotationLine(cfgStr string, x1, y1, x2, y2 int, lineStart, lineEnd string) string {
	marker := ""
	if len(lineStart) > 0 {
		marker = fmt.Sprintf("\n\t\tline_start = %q\n\t\tline_end   = %q", lineStart, lineEnd)
	}
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
		border_color = "#808080FF"%[6]s
	}
}
	`, cfgStr, x1, y1, x2, y2, marker)
}

func testAccAnnotationLineNullMarkers(cfgStr string) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_lab" "l" {
	title = "acc annotation line null markers"
}

resource "cml2_annotation" "a" {
	lab_id = cml2_lab.l.id
	type   = "line"

	line = {
		x1 = 10
		y1 = 20
		x2 = 35
		y2 = 45
		border_color = "#808080FF"
		line_start = null
		line_end   = null
	}
}
`, cfgStr)
}

func testAccAnnotationLineNullCombo(cfgStr string, startNull, endNull bool) string {
	start := "\n\t\tline_start = \"arrow\""
	if startNull {
		start = "\n\t\tline_start = null"
	}
	end := "\n\t\tline_end   = \"arrow\""
	if endNull {
		end = "\n\t\tline_end   = null"
	}
	return fmt.Sprintf(`
%[1]s

resource "cml2_lab" "l" {
	title = "acc annotation line null combo"
}

resource "cml2_annotation" "a" {
	lab_id = cml2_lab.l.id
	type   = "line"

	line = {
		x1 = 10
		y1 = 20
		x2 = 35
		y2 = 45
		border_color = "#808080FF"%[2]s%[3]s
	}
}
`, cfgStr, start, end)
}

func testAccAnnotationLineInvalidMarker(cfgStr string) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_lab" "l" {
	title = "acc annotation line invalid marker"
}

resource "cml2_annotation" "a" {
	lab_id = cml2_lab.l.id
	type   = "line"

	line = {
		x1 = 10
		y1 = 20
		x2 = 30
		y2 = 40
		line_start = "triangle"
	}
}
`, cfgStr)
}

func testAccAnnotationTextInvalidBorderStyle(cfgStr string) string {
	return fmt.Sprintf(`
%[1]s

resource "cml2_lab" "l" {
	title = "acc annotation invalid border style"
}

resource "cml2_annotation" "a" {
	lab_id = cml2_lab.l.id
	type   = "text"

	text = {
		text_content = "hello"
		x1          = 10
		y1          = 20
		border_style = "1,2,3"
	}
}
`, cfgStr)
}

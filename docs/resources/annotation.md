---
page_title: "cml2_annotation Resource"
subcategory: "Resources"
description: "Manages a classic annotation in a CML lab (text/rectangle/ellipse/line)."
---

# cml2_annotation (Resource)

Manages a classic annotation in a CML lab.

Supported types: `text`, `rectangle`, `ellipse`, `line`.

## Example Usage

```hcl
resource "cml2_annotation" "note" {
  lab_id = cml2_lab.this.id
  type   = "text"

  text = {
    text_content = "Hello from Terraform"
    x1          = 10
    y1          = 20
  }
}
```

Line example:

```hcl
resource "cml2_annotation" "edge" {
  lab_id = cml2_lab.this.id
  type   = "line"

  line = {
    x1         = 10
    y1         = 20
    x2         = 30
    y2         = 40
    line_start = "arrow"
    line_end   = "arrow"
  }
}
```

## Schema

### Required

- `lab_id` (String) Lab ID.
- `type` (String) Annotation type. Supported: `text`, `rectangle`, `ellipse`, `line`.

### Optional

- `text` (Block) Text annotation attributes (required when `type = "text"`).
- `rectangle` (Block) Rectangle annotation attributes (required when `type = "rectangle"`).
- `ellipse` (Block) Ellipse annotation attributes (required when `type = "ellipse"`).
- `line` (Block) Line annotation attributes (required when `type = "line"`).

### Read-only

- `id` (String) Annotation ID.

## Import

Import format: `<lab_id>/<annotation_id>`

```bash
terraform import cml2_annotation.note "<lab_id>/<annotation_id>"
```

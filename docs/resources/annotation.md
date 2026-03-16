---
page_title: "cml2_annotation Resource"
subcategory: "Resources"
description: "Manages a classic annotation in a CML lab (text annotations)."
---

# cml2_annotation (Resource)

Manages a classic annotation in a CML lab.

Currently supported: `type = "text"`.

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

## Schema

### Required

- `lab_id` (String) Lab ID.
- `type` (String) Annotation type. Currently supported: `text`.

### Optional

- `text` (Block) Text annotation attributes (required when `type = "text"`).

### Read-only

- `id` (String) Annotation ID.

## Import

Import format: `<lab_id>/<annotation_id>`

```bash
terraform import cml2_annotation.note "<lab_id>/<annotation_id>"
```

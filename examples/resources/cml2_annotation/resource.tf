resource "cml2_lab" "lab" {
  title = "annotation-example"
}

resource "cml2_annotation" "text" {
  lab_id = cml2_lab.lab.id
  type   = "text"

  text = {
    text_content = "Hello"
    x1           = 0
    y1           = 0

    # Optional styling
    color        = "#808080FF"
    border_color = "#00000000"
    text_font    = "monospace"
    text_size    = 12
    text_unit    = "pt"
  }
}

resource "cml2_annotation" "rect" {
  lab_id = cml2_lab.lab.id
  type   = "rectangle"

  rectangle = {
    # x1/y1: origin, x2/y2: width/height
    x1 = -440
    y1 = -80
    x2 = 120
    y2 = 120

    rotation      = 45
    border_style  = "4,2"
    border_radius = 12
    thickness     = 1
    color         = "#FFFFFFFF"
    border_color  = "#808080FF"
  }
}

resource "cml2_annotation" "line_plain" {
  lab_id = cml2_lab.lab.id
  type   = "line"

  line = {
    x1 = -40
    y1 = -80
    x2 = -120
    y2 = 40

    # Explicit null clears line endings (plain line)
    line_start = null
    line_end   = null

    border_style = "2,2"
    thickness    = 1
    color        = "#FFFFFFFF"
    border_color = "#808080FF"
  }
}

package annotation

import "github.com/rschmied/gocmlclient/pkg/models"

func borderStyleForCML(style string, modern bool) string {
	if !modern {
		return style
	}
	switch style {
	case "":
		return "solid"
	case "2,2":
		return "dotted"
	case "4,2":
		return "dashed"
	default:
		return style
	}
}

func borderStyleForTerraform(style string, modern bool) string {
	if !modern {
		return style
	}
	switch style {
	case "solid":
		return ""
	case "dotted":
		return "2,2"
	case "dashed":
		return "4,2"
	default:
		return style
	}
}

func translateAnnotationCreateBorderStyle(create *models.AnnotationCreate, modern bool) {
	if create.Text != nil {
		create.Text.BorderStyle = models.BorderStyle(borderStyleForCML(string(create.Text.BorderStyle), modern))
	}
	if create.Rectangle != nil {
		create.Rectangle.BorderStyle = models.BorderStyle(borderStyleForCML(string(create.Rectangle.BorderStyle), modern))
	}
	if create.Ellipse != nil {
		create.Ellipse.BorderStyle = models.BorderStyle(borderStyleForCML(string(create.Ellipse.BorderStyle), modern))
	}
	if create.Line != nil {
		create.Line.BorderStyle = models.BorderStyle(borderStyleForCML(string(create.Line.BorderStyle), modern))
	}
}

func translateAnnotationUpdateBorderStyle(update *models.AnnotationUpdate, modern bool) {
	if update.Text != nil && update.Text.BorderStyle != nil {
		v := models.BorderStyle(borderStyleForCML(string(*update.Text.BorderStyle), modern))
		update.Text.BorderStyle = &v
	}
	if update.Rectangle != nil && update.Rectangle.BorderStyle != nil {
		v := models.BorderStyle(borderStyleForCML(string(*update.Rectangle.BorderStyle), modern))
		update.Rectangle.BorderStyle = &v
	}
	if update.Ellipse != nil && update.Ellipse.BorderStyle != nil {
		v := models.BorderStyle(borderStyleForCML(string(*update.Ellipse.BorderStyle), modern))
		update.Ellipse.BorderStyle = &v
	}
	if update.Line != nil && update.Line.BorderStyle != nil {
		v := models.BorderStyle(borderStyleForCML(string(*update.Line.BorderStyle), modern))
		update.Line.BorderStyle = &v
	}
}

func translateAnnotationResponseBorderStyle(annotation *models.Annotation, modern bool) {
	if annotation.Text != nil {
		annotation.Text.BorderStyle = models.BorderStyle(borderStyleForTerraform(string(annotation.Text.BorderStyle), modern))
	}
	if annotation.Rectangle != nil {
		annotation.Rectangle.BorderStyle = models.BorderStyle(borderStyleForTerraform(string(annotation.Rectangle.BorderStyle), modern))
	}
	if annotation.Ellipse != nil {
		annotation.Ellipse.BorderStyle = models.BorderStyle(borderStyleForTerraform(string(annotation.Ellipse.BorderStyle), modern))
	}
	if annotation.Line != nil {
		annotation.Line.BorderStyle = models.BorderStyle(borderStyleForTerraform(string(annotation.Line.BorderStyle), modern))
	}
}

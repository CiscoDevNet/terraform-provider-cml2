package cmlschema_test

import (
	"context"
	"testing"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

var user1 = &models.User{
	UserBase: models.UserBase{
		Username:    "root",
		Fullname:    "root with first name",
		Email:       "root string",
		Description: "root string",
		IsAdmin:     true,
		Groups: []models.UUID{
			"90f84e38-a71c-4d57-8d90-00fa8a197385",
			"60f84e39-ffff-4d99-8a78-00fa8aaf5666",
		},
		OptIn: func() *models.OptInState {
			v := models.OptInAccepted
			return &v
		}(),
	},
	ID:          "ce5ed922-9aff-4d27-ac3e-f62b4440d2e0",
	DirectoryDN: "DN=none",
	Labs: []models.UUID{
		"e0e18ef5-9d1f-4cbb-99e8-a6da60c20113",
		"712c0b01-e2d7-445f-88cc-31b274aece82",
	},
}

var user2 = func() *models.User {
	u := *user1
	rp := models.UUID("6e3f384c-713d-471f-9059-6a81cd00632f")
	u.ResourcePool = &rp
	return &u
}()

func TestUser(t *testing.T) {
	diag := &diag.Diagnostics{}
	ctx := context.Background()

	for _, user := range []*models.User{user1, user2} {
		value := cmlschema.NewUser(ctx, user, diag)
		t.Logf("value: %+v", value)
		t.Logf("errors: %+v", diag.Errors())
		assert.False(t, diag.HasError())
		var newUser cmlschema.UserModel
		diag.Append(tfsdk.ValueAs(ctx, value, &newUser)...)
		assert.False(t, diag.HasError())
	}
}

func TestUserSchema(t *testing.T) {
	userSchema := schema.Schema{
		Attributes: cmlschema.Converter(cmlschema.User()),
	}
	got, diag := userSchema.TypeAtPath(context.TODO(), path.Root("id"))
	assert.False(t, diag.HasError())
	assert.Equal(t, types.StringType, got)
}

package cmlclient

import (
	"context"
	"testing"

	mr "github.com/rschmied/terraform-provider-cml2/m/v2/pkg/mockresponder"
	"github.com/stretchr/testify/assert"
)

func TestClient_methoderror(t *testing.T) {
	c := NewClient("", true)
	err := c.jsonReq(context.Background(), "ü", "###", nil, nil)
	assert.Error(t, err)
}

func TestClient_putpatch(t *testing.T) {

	putResponse := mr.MockRespList{
		mr.MockResp{Code: 204},
	}

	patchResponse := mr.MockRespList{
		mr.MockResp{Data: []byte("\"OK\"")},
	}

	c := NewClient("", true)
	mresp, ctx := mr.NewMockResponder()
	c.httpClient = mresp
	mresp.SetData(putResponse)
	c.authChecked = true
	c.versionChecked = true

	err := c.jsonPut(ctx, "###")
	assert.NoError(t, err)

	mresp.SetData(patchResponse)
	var result string
	err = c.jsonPatch(ctx, "###", nil, &result)
	assert.NoError(t, err)
	assert.Equal(t, result, "OK")
}

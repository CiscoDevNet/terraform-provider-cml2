package cmlschema

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestConfig_StringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentConfig Config
		givenConfig   basetypes.StringValuable
		expectedMatch bool
		expectedDiags diag.Diagnostics
	}{
		"semantically equal - CRLF is identical ": {
			currentConfig: NewConfigValue("hostname bla\nip add add 1.2.3.4/24 dev eth0\nend\n"),
			givenConfig:   NewConfigValue("hostname bla\nip add add 1.2.3.4/24 dev eth0\nend\n"),
			expectedMatch: true,
		},
		"not equal - CRLF mismatch DOS/Unix": {
			givenConfig:   NewConfigValue("hostname bla\r\nip add add 1.2.3.4/24 dev eth0\r\nend\r\n"),
			currentConfig: NewConfigValue("hostname bla\nip add add 1.2.3.4/24 dev eth0\nend\n"),
			expectedMatch: true,
		},
		"incoorect type": {
			currentConfig: NewConfigValue("hostname bla\r\nip add add 1.2.3.4/24 dev eth0\r\nend\r\n"),
			givenConfig:   basetypes.NewStringValue("hostname bla\r\nip add add 1.2.3.4/24 dev eth0\r\nend\r\n"),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected value type was received while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Expected Value Type: cmlschema.Config\n"+
						"Got Value Type: basetypes.StringValue",
				),
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			match, diags := testCase.currentConfig.StringSemanticEquals(context.Background(), testCase.givenConfig)

			if testCase.expectedMatch != match {
				t.Errorf("Expected StringSemanticEquals to return: %t, but got: %t", testCase.expectedMatch, match)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

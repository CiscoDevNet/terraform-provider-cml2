// Package testing provides base configuration for (integration) testing
package testing

import (
	"os"
	"testing"
)

// SkipUnlessAcc skips acceptance-style tests unless explicitly enabled.
//
// Convention: terraform providers use TF_ACC=1 to enable tests that require
// external systems and/or a real Terraform CLI run.
func SkipUnlessAcc(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests skipped (set TF_ACC=1 to enable)")
	}
}

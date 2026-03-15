package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/rschmied/gocmlclient/pkg/client"
)

func Converge(ctx context.Context, client *client.Client, diags *diag.Diagnostics, id, timeout string) {
	converged := false
	snoozeFor := 5 // seconds
	var err error

	tflog.Info(ctx, "waiting for convergence")

	tov, err := time.ParseDuration(timeout)
	if err != nil {
		diags.AddError(ErrorLabel, fmt.Sprintf("can't parse timeout %q: %s", timeout, err))
		return
	}
	endTime := time.Now().Add(tov)

	ticker := time.NewTicker(time.Second * time.Duration(snoozeFor))
	defer ticker.Stop()

	attempts := 0

	for !converged {

		converged, err = client.LabHasConverged(ctx, id)
		if err != nil {
			diags.AddError(
				ErrorLabel,
				fmt.Sprintf("Wait for convergence of lab, got error: %s", err),
			)
			return
		}
		if converged {
			return
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
		if time.Now().After(endTime) {
			diags.AddError(ErrorLabel, fmt.Sprintf("ran into timeout (max %s)", timeout))
			return
		}
		attempts++
		tflog.Info(
			ctx, "converging",
			map[string]any{"seconds": attempts * snoozeFor},
		)
	}
}

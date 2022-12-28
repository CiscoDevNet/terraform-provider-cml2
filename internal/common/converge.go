package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"
)

func Converge(ctx context.Context, client *cmlclient.Client, diags *diag.Diagnostics, id string, timeout string) {
	converged := false
	waited := 0
	snoozeFor := 5 // seconds
	var err error

	tflog.Info(ctx, "waiting for convergence")

	tov, err := time.ParseDuration(timeout)
	if err != nil {
		panic("can't parse timeout -- should be validated")
	}
	endTime := time.Now().Add(tov)

	for !converged {

		converged, err = client.HasLabConverged(ctx, id)
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
		case <-time.After(time.Second * time.Duration(snoozeFor)):
		case <-ctx.Done():
			return
		}
		if time.Now().After(endTime) {
			diags.AddError(ErrorLabel, fmt.Sprintf("ran into timeout (max %s)", timeout))
			return
		}
		waited++
		tflog.Info(
			ctx, "converging",
			map[string]any{"seconds": waited * snoozeFor},
		)
	}
}

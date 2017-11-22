package main

import (
	"fmt"
	"io"
	"regexp"
	"text/tabwriter"

	hapi_release "k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/timeconv"
)

type hapiRelease struct {
	*hapi_release.Release
}

func (r *hapiRelease) PrintStatus(out io.Writer) error {
	if r.Info.LastDeployed != nil {
		fmt.Fprintf(out, "LAST DEPLOYED: %s\n", timeconv.String(r.Info.LastDeployed))
	}
	fmt.Fprintf(out, "NAMESPACE: %s\n", r.Namespace)
	fmt.Fprintf(out, "STATUS: %s\n", r.Info.Status.Code)
	fmt.Fprintf(out, "\n")
	if len(r.Info.Status.Resources) > 0 {
		re := regexp.MustCompile("  +")

		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "RESOURCES:\n%s\n", re.ReplaceAllString(r.Info.Status.Resources, "\t"))
		w.Flush()
	}
	if r.Info.Status.LastTestSuiteRun != nil {
		lastRun := r.Info.Status.LastTestSuiteRun
		fmt.Fprintf(out, "TEST SUITE:\n%s\n%s\n\n",
			fmt.Sprintf("Last Started: %s", timeconv.String(lastRun.StartedAt)),
			fmt.Sprintf("Last Completed: %s", timeconv.String(lastRun.CompletedAt)),
		)
	}

	if len(r.Info.Status.Notes) > 0 {
		fmt.Fprintf(out, "NOTES:\n%s\n", r.Info.Status.Notes)
	}
	return nil
}

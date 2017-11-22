package main

import (
	"bytes"
	"strings"
	"testing"

	hapi_release "k8s.io/helm/pkg/proto/hapi/release"
)

func TestPrintStatus(t *testing.T) {
	r := hapiRelease{
		Release: &hapi_release.Release{
			Name:      "chaoskube",
			Namespace: "kube-system",
			Info: &hapi_release.Info{
				Status: &hapi_release.Status{
					Code: hapi_release.Status_DEPLOYED,
				},
			},
		},
	}
	out := &bytes.Buffer{}

	if err := r.PrintStatus(out); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "STATUS: DEPLOYED") {
		t.Errorf("Unexpected PrintStatus(): %s", out.String())
	}
}

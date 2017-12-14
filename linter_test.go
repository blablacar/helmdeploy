package main

import (
	"strings"
	"testing"
)

func TestLintResource(t *testing.T) {
	resourceYaml :=
		`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: dashboard
spec:
  template:
    metadata:
      name: dashboard
    spec:
      containers:
      - name: dashboard`

	err := LintResource(resourceYaml)
	if !strings.Contains(err.Error(), "deploy/dashboard container have no liveness nor readiness probe") {
		t.Errorf("Unexpected StateProbes() error : %s", err)
	}
}

package main

import (
	"fmt"
	"testing"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	hapi_release "k8s.io/helm/pkg/proto/hapi/release"
	hapi_services "k8s.io/helm/pkg/proto/hapi/services"
)

func TestNewDeployerFromManifest(t *testing.T) {

}

type FakeClient struct {
	helm.FakeClient
}

func (c *FakeClient) UpdateReleaseFromChart(rlsName string, chart *chart.Chart, opts ...helm.UpdateOption) (*hapi_services.UpdateReleaseResponse, error) {
	return &hapi_services.UpdateReleaseResponse{c.Rels[0]}, c.Err
}

func TestIsIstalled(t *testing.T) {
	deploy := &Deploy{
		ReleaseName: "chaoskube",
		Namespace:   "default",
		TillerClient: &helm.FakeClient{
			Rels: []*hapi_release.Release{
				&hapi_release.Release{Name: "chaoskube"},
			},
			Err: nil,
		},
	}
	installed := deploy.IsInstalled()
	if installed != true {
		t.Errorf("Unexpected IsInstalled() : %v", installed)
	}

	deploy.TillerClient = &helm.FakeClient{
		Err: fmt.Errorf("release: %q not found", deploy.ReleaseName),
	}
	installed = deploy.IsInstalled()
	if installed != false {
		t.Errorf("Unexpected IsInstalled() : %v", installed)
	}
}

func TestDeploy(t *testing.T) {
	deploy := &Deploy{
		ReleaseName: "chaoskube",
		Namespace:   "default",
		TillerClient: &FakeClient{
			FakeClient: helm.FakeClient{
				Rels: []*hapi_release.Release{
					&hapi_release.Release{Name: "chaoskube"},
				},
				Err: nil,
			},
		},
	}

	release, err := deploy.Deploy(true)
	if err != nil {
		t.Fatal(err)
	}
	if release.Name != "chaoskube" {
		t.Errorf("Unexpected release : %v", release)
	}
}

func TestContent(t *testing.T) {
	deploy := &Deploy{
		ReleaseName: "chaoskube",
		Namespace:   "default",
		TillerClient: &FakeClient{
			FakeClient: helm.FakeClient{
				Rels: []*hapi_release.Release{
					&hapi_release.Release{Name: "chaoskube"},
				},
				Err: nil,
			},
		},
	}

	release, err := deploy.Content()
	if err != nil {
		t.Fatal(err)
	}
	if release.Name != "chaoskube" {
		t.Errorf("Unexpected release : %v", release)
	}
}

func TestRender(t *testing.T) {
	//deploy, err := NewDeployerFromManifest("./examples/kubernetes-dashboard-release.yaml", "kube-system", "tiller-deploy")
	manifest, err := NewManifestFromFile("./examples/kubernetes-dashboard-release.yaml")
	if err != nil {
		t.Fatal(err)
	}
	chart, err := chartutil.Load(manifest.Chart)
	if err != nil {
		t.Fatal(err)
	}

	deploy := &Deploy{
		ReleaseName: manifest.Name,
		Namespace:   manifest.Namespace,
		Chart:       chart,
	}
	deploy.Render()
}

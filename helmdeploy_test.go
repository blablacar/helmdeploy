package main

import (
	"testing"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	hapi_release "k8s.io/helm/pkg/proto/hapi/release"
)

func TestNewDeployerFromManifest(t *testing.T) {

}

type FakeClient struct {
	helm.FakeClient
}

func TestIsIstalled(t *testing.T) {
	deploy := &Deploy{
		ReleaseName: "chaoskube",
		Namespace:   "default",
		TillerClient: &helm.FakeClient{
			Rels: []*hapi_release.Release{
				&hapi_release.Release{Name: "chaoskube"},
			},
		},
	}
	installed := deploy.IsInstalled()
	if installed != true {
		t.Errorf("Unexpected IsInstalled() : %v", installed)
	}

	deploy.TillerClient = &helm.FakeClient{
		Rels: nil,
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
	res, err := deploy.Render()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := res["kubernetes-dashboard/templates/deployment.yaml"]; !ok {
		t.Errorf("Unexpected templating result : %v", res)
	}
}

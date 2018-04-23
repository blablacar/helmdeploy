package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/apex/log"
	"github.com/imdario/mergo"
	yaml "gopkg.in/yaml.v2"

	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	hapi_release "k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/strvals"
)

type Deployer interface {
	Deploy() error
}

type Deploy struct {
	ReleaseName  string
	Namespace    string
	Chart        *chart.Chart
	ValueFiles   []string
	Values       []string
	KubeClient   kubernetes.Interface
	TillerClient helm.Interface
}

func NewDeployerFromManifest(manifestPath string, tillerNamespace, tillerServiceName string) (*Deploy, error) {
	manifest, err := NewManifestFromFile(manifestPath)
	if err != nil {
		return &Deploy{}, err
	}

	chart, err := chartutil.Load(manifest.Chart)
	if err != nil {
		return &Deploy{}, fmt.Errorf("Failed to load chart %s: %v", manifest.Chart, err)
	}

	configPath := path.Join(os.Getenv("HOME"), ".kube", "config")
	kubeClient, err := NewKubeClient(configPath, manifest.Context, manifest.Cluster)
	if err != nil {
		return &Deploy{}, err
	}

	tillerEndpoints, err := kubeClient.GetEndpoints(tillerNamespace, tillerService)
	if err != nil {
		return &Deploy{}, err
	}
	log.Debugf("Using Tiller endpoint %s", tillerEndpoints[0])

	return &Deploy{
		ReleaseName:  manifest.Name,
		Namespace:    manifest.Namespace,
		Chart:        chart,
		Values:       manifest.Values,
		ValueFiles:   manifest.ValueFiles,
		KubeClient:   kubeClient,
		TillerClient: helm.NewClient(helm.Host(tillerEndpoints[0]), helm.ConnectTimeout(5)),
	}, nil
}

func (d *Deploy) IsInstalled() bool {
	h, err := d.TillerClient.ReleaseHistory(d.ReleaseName, helm.WithMaxHistory(1))
	if err != nil && strings.Contains(err.Error(), fmt.Sprintf("release: %q not found", d.ReleaseName)) {
		return false
	}

	if err != nil {
		log.Fatal(err.Error())
	}

	if len(h.Releases) == 0 {
		return false
	}

	return true
}

func (d *Deploy) Deploy(dryRun bool) (*hapiRelease, error) {
	release := &hapiRelease{}
	overrides, err := d.MergeOverrides()
	if err != nil {
		return &hapiRelease{}, err
	}

	logger := log.WithFields(log.Fields{
		"Name":      d.ReleaseName,
		"Namespace": d.Namespace,
	})

	if !d.IsInstalled() {
		logger.Debug("Installing release")
		response, err := d.TillerClient.InstallReleaseFromChart(
			d.Chart,
			d.Namespace,
			helm.ReleaseName(d.ReleaseName),
			helm.ValueOverrides(overrides),
			helm.InstallDryRun(dryRun),
		)
		if err != nil {
			return &hapiRelease{}, err
		}
		release.Release = response.GetRelease()
	} else {
		logger.Debug("Upgrading release")
		response, err := d.TillerClient.UpdateReleaseFromChart(
			d.ReleaseName,
			d.Chart,
			helm.UpdateValueOverrides(overrides),
			helm.UpgradeDryRun(dryRun),
		)
		if err != nil {
			return &hapiRelease{}, err
		}
		release.Release = response.GetRelease()
	}

	return release, nil
}

func (d *Deploy) AddOverrides(valueFiles []string, values []string) {
	for _, vf := range valueFiles {
		d.ValueFiles = append(d.ValueFiles, vf)
	}
	for _, v := range values {
		d.Values = append(d.Values, v)
	}
}

func (d *Deploy) MergeOverrides() ([]byte, error) {
	base := map[string]interface{}{}

	for _, filePath := range d.ValueFiles {
		currentMap := map[string]interface{}{}

		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse ValueFile %s : %s", filePath, err)
		}

		mergo.Merge(&base, currentMap)
	}

	for _, Value := range d.Values {
		if err := strvals.ParseInto(Value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing Value: %s", err)
		}
	}

	return yaml.Marshal(base)
}

func (d *Deploy) Status() (*hapiRelease, error) {
	response, err := d.TillerClient.ReleaseStatus(d.ReleaseName)
	if err != nil {
		return &hapiRelease{}, err
	}
	return &hapiRelease{
		Release: &hapi_release.Release{
			Name:      response.Name,
			Namespace: response.Namespace,
			Info:      response.Info,
		},
	}, nil

}

func (d *Deploy) Content() (*hapiRelease, error) {
	response, err := d.TillerClient.ReleaseContent(d.ReleaseName)
	if err != nil {
		return &hapiRelease{}, err
	}
	return &hapiRelease{Release: response.GetRelease()}, nil
}

func (d *Deploy) Render() (map[string]string, error) {
	options := chartutil.ReleaseOptions{
		Name:      d.ReleaseName,
		Namespace: d.Namespace,
	}
	overrides, err := d.MergeOverrides()
	if err != nil {
		return nil, err
	}

	config := &chart.Config{
		Raw: string(overrides),
	}
	capabilities := &chartutil.Capabilities{}

	values, err := chartutil.ToRenderValuesCaps(d.Chart, config, options, capabilities)

	if err != nil {
		return nil, err
	}

	renderer := engine.New()
	out, err := renderer.Render(d.Chart, values)
	return out, nil
}

package main

import (
	"io/ioutil"
	"path"

	yaml "gopkg.in/yaml.v2"
)

type Manifest struct {
	Name       string   `yaml:"name"`
	Context    string   `yaml:"context"`
	Cluster    string   `yaml:"cluster"`
	Namespace  string   `yaml:"namespace"`
	Chart      string   `yaml:"chart"`
	ValueFiles []string `yaml:"values"`
	Values     []string `yaml:"set"`
}

func NewManifestFromFile(filePath string) (*Manifest, error) {
	var manifest Manifest

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, err
	}

	// normalize paths
	basePath := path.Dir(filePath)
	if !path.IsAbs(manifest.Chart) {
		manifest.Chart = path.Join(basePath, manifest.Chart)
	}
	for i, vf := range manifest.ValueFiles {
		if path.IsAbs(vf) {
			continue
		}
		manifest.ValueFiles[i] = path.Join(basePath, vf)

	}
	return &manifest, nil
}

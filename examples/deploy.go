package main

import "github.com/blablacar/helmdeploy"

func main() {
	helmdeployer, _ := helmdeploy.NewDeployerFromManifest("")
	helmdeployer.deploy()
}

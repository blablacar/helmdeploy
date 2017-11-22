package main

import "github.com/remyLemeunier/helmdeploy"

func main() {
	helmdeployer, _ := helmdeploy.NewDeployerFromManifest("")
	helmdeployer.deploy()
}

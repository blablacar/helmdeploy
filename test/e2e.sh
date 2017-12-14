#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

kubernetes_version=${KUBERNETES_VERSION:-"1.8.0"}
helm_version=${HELM_VERSION:-"2.7.2"}
use_sudo=${E2E_USE_SUDO:-"no"}


function kubectl::install() {
	curl -Lo kubectl "https://storage.googleapis.com/kubernetes-release/release/v${kubernetes_version}/bin/linux/amd64/kubectl"
	chmod +x kubectl 
	maybesudo mv kubectl /usr/local/bin/
}


function minikube::install() {
	curl -Lo minikube "https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64"
	chmod +x minikube
	maybesudo mv minikube /usr/local/bin/
}

function helm::install() {
  curl -Lo helm.tar.gz "https://storage.googleapis.com/kubernetes-helm/helm-v${helm_version}-linux-amd64.tar.gz"
  tar --strip 1 -xzvf helm.tar.gz linux-amd64/helm
  chmod +x helm
  maybesudo mv helm /usr/local/bin/
}

function minikube::setup() {
	if [ x"$(minikube::status)" != x"Running" ]; then
		export CHANGE_MINIKUBE_NONE_USER=true
		maybesudo minikube start --vm-driver=none --kubernetes-version="v${kubernetes_version}"
		minikube update-context
		minikube::wait_ready
	fi
}

function minikube::stop() {
	maybesudo minikube stop
}

function minikube::status() {
	minikube status --format '{{.MinikubeStatus}}'
}

function minikube::wait_ready() {
	JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}';
	until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1  | grep -q "Ready=True"; do
		sleep 1;
	done
}

function maybesudo() {
	if [ x"$use_sudo" == x"yes" ] || [ x"$use_sudo" == x"y" ]; then
		sudo "$@"
	else
		"$@"
	fi
}

function helm::setup(){
	helm init --upgrade
	kubectl rollout status --watch --namespace=kube-system deployment/tiller-deploy 
	#sudo route -n add -net 172.17.0.1/24 192.168.99.100 # it's an ugly hack to have direct acces to the pods subnet :(
}

function travis::setup() {
	use_sudo="yes"
	kubectl::install
	minikube::install
	helm::install
}

function e2e::test_deploy() {
	./helmdeploy deploy "$(dirname $0)/../examples/kubernetes-dashboard-release.yaml"
	kubectl rollout status --watch --namespace=kube-system --cluster=minikube deployment/kubernetes-dashboard-kubernetes-dashboard
}


function main() {
	TRAVIS=${TRAVIS:-"false"}
	if [ x"${TRAVIS}" == x"true" ]; then
		travis::setup
	fi
	
	minikube::setup
	trap minikube::stop EXIT
	
	minikube version
	
	helm::setup

	e2e::test_deploy
}

main

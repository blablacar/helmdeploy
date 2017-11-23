#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

kubernetes_version=${KUBERNETES_VERSION:-"1.8.0"}
helm_version=${HELM_VERSION:-"2.7.2"}

if [ ! -x "$(command -v kubectl)" ]; then
	curl -Lo kubectl "https://storage.googleapis.com/kubernetes-release/release/v${kubernetes_version}/bin/linux/amd64/kubectl"
	chmod +x kubectl 
	sudo mv kubectl /usr/local/bin/
fi

if [ ! -x "$(command -v minikube)" ]; then
	curl -Lo minikube "https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64"
	chmod +x minikube
	sudo mv minikube /usr/local/bin/
fi

if [ ! -x "$(command -v helm)" ]; then
  curl -Lo helm.tar.gz "https://storage.googleapis.com/kubernetes-helm/helm-v${helm_version}-linux-amd64.tar.gz"
  tar --strip 1 -xzvf helm.tar.gz linux-amd64/helm
  chmod +x helm
  sudo mv helm /usr/local/bin/
fi

if [ x"$(minikube status --format '{{.MinikubeStatus}}')" != x"Running" ]; then
	export CHANGE_MINIKUBE_NONE_USER=true
	sudo minikube start --vm-driver=none --kubernetes-version="v${kubernetes_version}"
	minikube update-context
	#trap "sudo minikube stop" EXIT
fi

minikube version

 JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1; done

helm init --upgrade
kubectl rollout status --watch --namespace=kube-system deployment/tiller-deploy 
#sudo route -n add -net 172.17.0.1/24 192.168.99.100 # it's an ugly hack to have direct acces to the pods subnet :(

./helmdeploy deploy "$(dirname $0)/../examples/kubernetes-dashboard-release.yaml"
kubectl rollout status --watch --namespace=kube-system --cluster=minikube deployment/kubernetes-dashboard-kubernetes-dashboard

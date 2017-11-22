package main

import (
	"testing"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	clientset *Clientset
)

func init() {
	kubeclientset := fake.NewSimpleClientset(
		&v1.EndpointsList{
			Items: []v1.Endpoints{
				v1.Endpoints{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "tiller",
						Namespace: "kube-system",
					},
					Subsets: []v1.EndpointSubset{
						v1.EndpointSubset{
							Addresses: []v1.EndpointAddress{
								v1.EndpointAddress{
									IP: "127.0.0.1",
								},
							},
							Ports: []v1.EndpointPort{
								v1.EndpointPort{
									Port: 8080,
								},
							},
						},
					},
				},
			},
		},
	)
	clientset = &Clientset{
		Interface: kubeclientset,
	}
}

func TestNewKubeClient(t *testing.T) {
	clientset, err := NewKubeClient("./testdata/kubeconfig", "", "local")
	if err != nil {
		t.Fatal(err)
	}
	if clientset == (&Clientset{}) {
		t.Errorf("Unexpected return from NewKubeClient() : %q", clientset)
	}

}

func TestGetEndpoint(t *testing.T) {
	endpoints, err := clientset.GetEndpoints("kube-system", "tiller")
	if err != nil {
		t.Fatal(err)
	}
	if len(endpoints) != 1 {
		t.Errorf("Unexpected return from GetEndpoints() : %q", endpoints)
	}
	if endpoints[0] != "127.0.0.1:8080" {
		t.Errorf("Unexpected return from GetEndpoints() : %q", endpoints)
	}
}

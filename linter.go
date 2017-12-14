package main

import (
	"fmt"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
)

func LintResource(resourceYaml string) error {
	unmarshaler := scheme.Codecs.UniversalDeserializer()
	resource, _, _ := unmarshaler.Decode([]byte(resourceYaml), nil, nil)

	switch resource.(type) {
	case *v1beta1.Deployment:
		err := StateProbes(resource.(*v1beta1.Deployment))
		if err != nil {
			return err
		}
	}
	return nil
}

func StateProbes(deploy *v1beta1.Deployment) error {
	for _, container := range deploy.Spec.Template.Spec.Containers {
		if container.LivenessProbe == nil && container.ReadinessProbe == nil {
			return fmt.Errorf("deploy/%s container have no liveness nor readiness probe", deploy.Name)
		}
	}
	return nil
}

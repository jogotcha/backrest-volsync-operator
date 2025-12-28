package volsync

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	Group   = "volsync.backube"
	Version = "v1alpha1"
)

func RepositorySecretName(obj *unstructured.Unstructured) (string, error) {
	secretName, found, err := unstructured.NestedString(obj.Object, "spec", "restic", "repository")
	if err != nil {
		return "", fmt.Errorf("read spec.restic.repository: %w", err)
	}
	if !found || secretName == "" {
		return "", fmt.Errorf("spec.restic.repository not set")
	}
	return secretName, nil
}

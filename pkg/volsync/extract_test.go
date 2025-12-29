package volsync

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestRepositorySecretName(t *testing.T) {
	mk := func(obj map[string]any) *unstructured.Unstructured {
		return &unstructured.Unstructured{Object: obj}
	}

	t.Run("ok", func(t *testing.T) {
		obj := mk(map[string]any{
			"spec": map[string]any{
				"restic": map[string]any{
					"repository": "repo-secret",
				},
			},
		})
		name, err := RepositorySecretName(obj)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if name != "repo-secret" {
			t.Fatalf("expected repo-secret, got %q", name)
		}
	})

	t.Run("missing", func(t *testing.T) {
		obj := mk(map[string]any{
			"spec": map[string]any{
				"restic": map[string]any{},
			},
		})
		_, err := RepositorySecretName(obj)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		obj := mk(map[string]any{
			"spec": map[string]any{
				"restic": map[string]any{
					"repository": 123,
				},
			},
		})
		_, err := RepositorySecretName(obj)
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

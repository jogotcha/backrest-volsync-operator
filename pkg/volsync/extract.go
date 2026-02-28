package volsync

import (
	"fmt"
	"strings"

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

func ReplicationSourceCompletionMarker(obj *unstructured.Unstructured) (string, bool, error) {
	if obj == nil {
		return "", false, fmt.Errorf("volsync object is nil")
	}
	if obj.GetKind() != "ReplicationSource" {
		return "", false, nil
	}

	for _, p := range []struct {
		key  string
		path []string
	}{
		{key: "lastSnapshotID", path: []string{"status", "lastSnapshotID"}},
		{key: "lastSnapshot", path: []string{"status", "lastSnapshot"}},
		{key: "latestImage", path: []string{"status", "latestImage", "name"}},
		{key: "lastManualSync", path: []string{"status", "lastManualSync"}},
	} {
		v, found, err := unstructured.NestedString(obj.Object, p.path...)
		if err != nil {
			return "", false, fmt.Errorf("read %s: %w", strings.Join(p.path, "."), err)
		}
		v = strings.TrimSpace(v)
		if found && v != "" {
			return p.key + "=" + v, true, nil
		}
	}

	lastSyncTime, found, err := unstructured.NestedString(obj.Object, "status", "lastSyncTime")
	if err != nil {
		return "", false, fmt.Errorf("read status.lastSyncTime: %w", err)
	}
	lastSyncTime = strings.TrimSpace(lastSyncTime)
	if found && lastSyncTime != "" {
		return "lastSyncTime=" + lastSyncTime, true, nil
	}

	return "", false, nil
}

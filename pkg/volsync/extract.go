package volsync

import (
	"fmt"
	"sort"
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

	parts := make([]string, 0, 4)

	lastSyncTime, found, err := unstructured.NestedString(obj.Object, "status", "lastSyncTime")
	if err != nil {
		return "", false, fmt.Errorf("read status.lastSyncTime: %w", err)
	}
	if found && strings.TrimSpace(lastSyncTime) != "" {
		parts = append(parts, "lastSyncTime="+strings.TrimSpace(lastSyncTime))
	}

	for _, p := range []struct {
		key  string
		path []string
	}{
		{key: "latestImage", path: []string{"status", "latestImage", "name"}},
		{key: "lastSnapshot", path: []string{"status", "lastSnapshot"}},
		{key: "lastSnapshotID", path: []string{"status", "lastSnapshotID"}},
		{key: "lastManualSync", path: []string{"status", "lastManualSync"}},
	} {
		v, ok, nestedErr := unstructured.NestedString(obj.Object, p.path...)
		if nestedErr != nil {
			return "", false, fmt.Errorf("read %s: %w", strings.Join(p.path, "."), nestedErr)
		}
		v = strings.TrimSpace(v)
		if ok && v != "" {
			parts = append(parts, p.key+"="+v)
		}
	}

	if len(parts) == 0 {
		return "", false, nil
	}

	sort.Strings(parts)
	return strings.Join(parts, "|"), true, nil
}

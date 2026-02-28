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

func TestReplicationSourceCompletionMarker(t *testing.T) {
	mk := func(kind string, obj map[string]any) *unstructured.Unstructured {
		u := &unstructured.Unstructured{Object: obj}
		u.SetKind(kind)
		return u
	}

	t.Run("non source ignored", func(t *testing.T) {
		obj := mk("ReplicationDestination", map[string]any{})
		marker, ready, err := ReplicationSourceCompletionMarker(obj)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if ready {
			t.Fatalf("expected ready=false")
		}
		if marker != "" {
			t.Fatalf("expected empty marker, got %q", marker)
		}
	})

	t.Run("no status fields", func(t *testing.T) {
		obj := mk("ReplicationSource", map[string]any{"status": map[string]any{}})
		marker, ready, err := ReplicationSourceCompletionMarker(obj)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if ready {
			t.Fatalf("expected ready=false")
		}
		if marker != "" {
			t.Fatalf("expected empty marker, got %q", marker)
		}
	})

	t.Run("lastSyncTime only", func(t *testing.T) {
		obj := mk("ReplicationSource", map[string]any{
			"status": map[string]any{
				"lastSyncTime": "2026-02-24T12:00:00Z",
			},
		})
		marker, ready, err := ReplicationSourceCompletionMarker(obj)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !ready {
			t.Fatalf("expected ready=true")
		}
		expected := "lastSyncTime=2026-02-24T12:00:00Z"
		if marker != expected {
			t.Fatalf("expected %q, got %q", expected, marker)
		}
	})

	t.Run("prefers snapshot identity fields over lastSyncTime", func(t *testing.T) {
		obj := mk("ReplicationSource", map[string]any{
			"status": map[string]any{
				"lastSyncTime":   "2026-02-24T12:00:00Z",
				"lastSnapshot":   "abcd",
				"lastSnapshotID": "xyz",
				"latestImage": map[string]any{
					"name": "source-snap-1",
				},
			},
		})
		marker, ready, err := ReplicationSourceCompletionMarker(obj)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !ready {
			t.Fatalf("expected ready=true")
		}
		expected := "lastSnapshotID=xyz"
		if marker != expected {
			t.Fatalf("expected %q, got %q", expected, marker)
		}
	})

	t.Run("marker remains stable when only lastSyncTime changes", func(t *testing.T) {
		obj := mk("ReplicationSource", map[string]any{
			"status": map[string]any{
				"lastSyncTime": "2026-02-24T12:00:00Z",
				"lastSnapshot": "snap-a",
			},
		})
		marker1, ready, err := ReplicationSourceCompletionMarker(obj)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !ready {
			t.Fatalf("expected ready=true")
		}

		obj.Object["status"].(map[string]any)["lastSyncTime"] = "2026-02-24T12:05:00Z"
		marker2, ready, err := ReplicationSourceCompletionMarker(obj)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !ready {
			t.Fatalf("expected ready=true")
		}

		if marker1 != "lastSnapshot=snap-a" || marker2 != "lastSnapshot=snap-a" {
			t.Fatalf("expected stable snapshot marker, got marker1=%q marker2=%q", marker1, marker2)
		}
	})

	t.Run("wrong field type returns error", func(t *testing.T) {
		obj := mk("ReplicationSource", map[string]any{
			"status": map[string]any{
				"lastSyncTime": 123,
			},
		})
		_, _, err := ReplicationSourceCompletionMarker(obj)
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

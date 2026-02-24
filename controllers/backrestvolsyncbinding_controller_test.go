package controllers

import (
	"context"
	"testing"

	v1 "github.com/garethgeorge/backrest/gen/go/v1"
	"github.com/jogotcha/backrest-volsync-operator/api/v1alpha1"
	"github.com/jogotcha/backrest-volsync-operator/pkg/backrest"
	"github.com/jogotcha/backrest-volsync-operator/pkg/volsync"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type fakeBackrestRepoClient struct {
	addRepoCalls int
	taskCalls    []v1.DoRepoTaskRequest_Task
}

func (f *fakeBackrestRepoClient) AddRepo(_ context.Context, _ *v1.Repo) (*v1.Config, error) {
	f.addRepoCalls++
	return &v1.Config{}, nil
}

func (f *fakeBackrestRepoClient) DoRepoTask(_ context.Context, _ string, task v1.DoRepoTaskRequest_Task) error {
	f.taskCalls = append(f.taskCalls, task)
	return nil
}

func bindingTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme corev1: %v", err)
	}
	for _, kind := range []string{"ReplicationSource", "ReplicationDestination"} {
		gvk := schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: kind}
		scheme.AddKnownTypeWithName(gvk, &unstructured.Unstructured{})
	}
	return scheme
}

func getReadyReason(b *v1alpha1.BackrestVolSyncBinding) string {
	for _, c := range b.Status.Conditions {
		if c.Type == conditionReady {
			return c.Reason
		}
	}
	return ""
}

func TestBackrestVolSyncBindingReconcile_InvalidSpec(t *testing.T) {
	ctx := context.Background()
	scheme := bindingTestScheme(t)

	b := &v1alpha1.BackrestVolSyncBinding{}
	b.Namespace = "workload"
	b.Name = "b"
	// Missing spec.backrest.url etc.

	c := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.BackrestVolSyncBinding{}).
		WithObjects(b).
		Build()

	r := &BackrestVolSyncBindingReconciler{Client: c, Scheme: scheme}
	_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: b.Namespace, Name: b.Name}})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var got v1alpha1.BackrestVolSyncBinding
	if err := c.Get(ctx, types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, &got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if reason := getReadyReason(&got); reason != "InvalidSpec" {
		t.Fatalf("expected InvalidSpec, got %q", reason)
	}
}

func TestBackrestVolSyncBindingReconcile_PausedOverridesInvalid(t *testing.T) {
	ctx := context.Background()
	scheme := bindingTestScheme(t)

	cfg := &v1alpha1.BackrestVolSyncOperatorConfig{}
	cfg.Namespace = "backrest-volsync-operator"
	cfg.Name = "backrest-volsync-operator"
	cfg.Spec.Paused = true

	b := &v1alpha1.BackrestVolSyncBinding{}
	b.Namespace = "workload"
	b.Name = "b"

	c := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.BackrestVolSyncBinding{}).
		WithObjects(cfg, b).
		Build()

	r := &BackrestVolSyncBindingReconciler{Client: c, Scheme: scheme, OperatorConfig: types.NamespacedName{Namespace: cfg.Namespace, Name: cfg.Name}}
	_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: b.Namespace, Name: b.Name}})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var got v1alpha1.BackrestVolSyncBinding
	if err := c.Get(ctx, types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, &got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if reason := getReadyReason(&got); reason != "Paused" {
		t.Fatalf("expected Paused, got %q", reason)
	}
}

func TestBackrestVolSyncBindingReconcile_VolSyncMissingRepository(t *testing.T) {
	ctx := context.Background()
	scheme := bindingTestScheme(t)

	b := &v1alpha1.BackrestVolSyncBinding{}
	b.Namespace = "workload"
	b.Name = "b"
	b.Spec.Backrest.URL = "http://example.invalid"
	b.Spec.Source = v1alpha1.VolSyncSourceRef{Kind: "ReplicationSource", Name: "demo"}

	vs := &unstructured.Unstructured{}
	vs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	vs.SetNamespace("workload")
	vs.SetName("demo")
	vs.SetUID(types.UID("1111"))

	c := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.BackrestVolSyncBinding{}).
		WithObjects(b, vs).
		Build()

	r := &BackrestVolSyncBindingReconciler{Client: c, Scheme: scheme}
	_, _ = r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: b.Namespace, Name: b.Name}})

	var got v1alpha1.BackrestVolSyncBinding
	if err := c.Get(ctx, types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, &got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if reason := getReadyReason(&got); reason != "VolSyncMissingRepository" {
		t.Fatalf("expected VolSyncMissingRepository, got %q", reason)
	}
	if got.Status.LastErrorHash == "" {
		t.Fatalf("expected LastErrorHash set")
	}
}

func TestBackrestVolSyncBindingReconcile_RepositorySecretInvalid(t *testing.T) {
	ctx := context.Background()
	scheme := bindingTestScheme(t)

	b := &v1alpha1.BackrestVolSyncBinding{}
	b.Namespace = "workload"
	b.Name = "b"
	b.Spec.Backrest.URL = "http://example.invalid"
	b.Spec.Source = v1alpha1.VolSyncSourceRef{Kind: "ReplicationSource", Name: "demo"}

	vs := &unstructured.Unstructured{}
	vs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	vs.SetNamespace("workload")
	vs.SetName("demo")
	vs.Object = map[string]any{
		"apiVersion": volsync.Group + "/" + volsync.Version,
		"kind":       "ReplicationSource",
		"metadata": map[string]any{
			"name":      "demo",
			"namespace": "workload",
		},
		"spec": map[string]any{
			"restic": map[string]any{
				"repository": "repo-secret",
			},
		},
	}

	sec := &corev1.Secret{}
	sec.Namespace = "workload"
	sec.Name = "repo-secret"
	sec.Data = map[string][]byte{
		"RESTIC_REPOSITORY": []byte("s3://bucket/repo"),
		// Missing RESTIC_PASSWORD
	}
	sec.SetUID(types.UID("2222"))
	sec.SetResourceVersion("1")

	c := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.BackrestVolSyncBinding{}).
		WithObjects(b, vs, sec).
		Build()

	r := &BackrestVolSyncBindingReconciler{Client: c, Scheme: scheme}
	_, _ = r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: b.Namespace, Name: b.Name}})

	var got v1alpha1.BackrestVolSyncBinding
	if err := c.Get(ctx, types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, &got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if reason := getReadyReason(&got); reason != "RepositorySecretInvalid" {
		t.Fatalf("expected RepositorySecretInvalid, got %q", reason)
	}
}

func TestBackrestVolSyncBindingReconcile_BackrestAuthInvalidWhenSecretMissing(t *testing.T) {
	ctx := context.Background()
	scheme := bindingTestScheme(t)

	b := &v1alpha1.BackrestVolSyncBinding{}
	b.Namespace = "workload"
	b.Name = "b"
	b.Spec.Backrest.URL = "http://example.invalid"
	b.Spec.Backrest.AuthRef = &v1alpha1.SecretRef{Name: "auth"}
	b.Spec.Source = v1alpha1.VolSyncSourceRef{Kind: "ReplicationSource", Name: "demo"}

	vs := &unstructured.Unstructured{}
	vs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	vs.SetNamespace("workload")
	vs.SetName("demo")
	vs.Object = map[string]any{
		"apiVersion": volsync.Group + "/" + volsync.Version,
		"kind":       "ReplicationSource",
		"metadata": map[string]any{
			"name":      "demo",
			"namespace": "workload",
		},
		"spec": map[string]any{
			"restic": map[string]any{
				"repository": "repo-secret",
			},
		},
	}

	sec := &corev1.Secret{}
	sec.Namespace = "workload"
	sec.Name = "repo-secret"
	sec.Data = map[string][]byte{
		"RESTIC_REPOSITORY": []byte("s3://bucket/repo"),
		"RESTIC_PASSWORD":   []byte("pass"),
	}
	sec.SetUID(types.UID("2222"))
	sec.SetResourceVersion("1")

	c := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.BackrestVolSyncBinding{}).
		WithObjects(b, vs, sec).
		Build()

	r := &BackrestVolSyncBindingReconciler{Client: c, Scheme: scheme}
	_, _ = r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: b.Namespace, Name: b.Name}})

	var got v1alpha1.BackrestVolSyncBinding
	if err := c.Get(ctx, types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, &got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if reason := getReadyReason(&got); reason != "BackrestAuthInvalid" {
		t.Fatalf("expected BackrestAuthInvalid, got %q", reason)
	}
}

func TestBackrestVolSyncBindingReconcile_TriggersSnapshotTasksWithDedupe(t *testing.T) {
	ctx := context.Background()
	scheme := bindingTestScheme(t)

	b := &v1alpha1.BackrestVolSyncBinding{}
	b.Namespace = "workload"
	b.Name = "b"
	b.Spec.Backrest.URL = "http://backrest.invalid"
	b.Spec.Source = v1alpha1.VolSyncSourceRef{Kind: "ReplicationSource", Name: "demo"}
	enabled := true
	b.Spec.Repo.TriggerTasksOnSnapshot = &enabled

	vs := &unstructured.Unstructured{}
	vs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	vs.SetNamespace("workload")
	vs.SetName("demo")
	vs.SetUID(types.UID("1111"))
	vs.Object = map[string]any{
		"apiVersion": volsync.Group + "/" + volsync.Version,
		"kind":       "ReplicationSource",
		"metadata": map[string]any{
			"name":      "demo",
			"namespace": "workload",
		},
		"spec": map[string]any{
			"restic": map[string]any{
				"repository": "repo-secret",
			},
		},
		"status": map[string]any{
			"lastSyncTime": "2026-02-24T12:00:00Z",
			"lastSnapshot": "snap-1",
		},
	}

	sec := &corev1.Secret{}
	sec.Namespace = "workload"
	sec.Name = "repo-secret"
	sec.Data = map[string][]byte{
		"RESTIC_REPOSITORY": []byte("s3://bucket/repo"),
		"RESTIC_PASSWORD":   []byte("pass"),
	}
	sec.SetUID(types.UID("2222"))
	sec.SetResourceVersion("1")

	c := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.BackrestVolSyncBinding{}).
		WithObjects(b, vs, sec).
		Build()

	br := &fakeBackrestRepoClient{}
	r := &BackrestVolSyncBindingReconciler{
		Client: c,
		Scheme: scheme,
		BackrestClientFactory: func(_ string, _ backrest.Auth) backrestRepoClient {
			return br
		},
	}

	if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: b.Namespace, Name: b.Name}}); err != nil {
		t.Fatalf("reconcile #1: %v", err)
	}
	if br.addRepoCalls != 1 {
		t.Fatalf("expected addRepoCalls=1, got %d", br.addRepoCalls)
	}
	if len(br.taskCalls) != 2 {
		t.Fatalf("expected 2 task calls, got %d", len(br.taskCalls))
	}
	if br.taskCalls[0] != v1.DoRepoTaskRequest_TASK_INDEX_SNAPSHOTS || br.taskCalls[1] != v1.DoRepoTaskRequest_TASK_STATS {
		t.Fatalf("unexpected task order: %#v", br.taskCalls)
	}

	var got v1alpha1.BackrestVolSyncBinding
	if err := c.Get(ctx, types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, &got); err != nil {
		t.Fatalf("get binding after reconcile #1: %v", err)
	}
	if got.Status.LastSnapshotMarker == "" {
		t.Fatalf("expected LastSnapshotMarker set")
	}

	if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: b.Namespace, Name: b.Name}}); err != nil {
		t.Fatalf("reconcile #2: %v", err)
	}
	if br.addRepoCalls != 1 {
		t.Fatalf("expected addRepoCalls still 1, got %d", br.addRepoCalls)
	}
	if len(br.taskCalls) != 2 {
		t.Fatalf("expected no additional task calls, got %d", len(br.taskCalls))
	}

	var vsUpdated unstructured.Unstructured
	vsUpdated.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	if err := c.Get(ctx, types.NamespacedName{Namespace: "workload", Name: "demo"}, &vsUpdated); err != nil {
		t.Fatalf("get volsync object: %v", err)
	}
	if err := unstructured.SetNestedField(vsUpdated.Object, "2026-02-24T12:30:00Z", "status", "lastSyncTime"); err != nil {
		t.Fatalf("set status.lastSyncTime: %v", err)
	}
	if err := unstructured.SetNestedField(vsUpdated.Object, "snap-2", "status", "lastSnapshot"); err != nil {
		t.Fatalf("set status.lastSnapshot: %v", err)
	}
	if err := c.Update(ctx, &vsUpdated); err != nil {
		t.Fatalf("update volsync object: %v", err)
	}

	if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: b.Namespace, Name: b.Name}}); err != nil {
		t.Fatalf("reconcile #3: %v", err)
	}
	if br.addRepoCalls != 1 {
		t.Fatalf("expected addRepoCalls still 1, got %d", br.addRepoCalls)
	}
	if len(br.taskCalls) != 4 {
		t.Fatalf("expected 4 task calls after new snapshot, got %d", len(br.taskCalls))
	}
}

func TestBackrestVolSyncBindingReconcile_DoesNotTriggerSnapshotTasksByDefault(t *testing.T) {
	ctx := context.Background()
	scheme := bindingTestScheme(t)

	b := &v1alpha1.BackrestVolSyncBinding{}
	b.Namespace = "workload"
	b.Name = "b"
	b.Spec.Backrest.URL = "http://backrest.invalid"
	b.Spec.Source = v1alpha1.VolSyncSourceRef{Kind: "ReplicationSource", Name: "demo"}

	vs := &unstructured.Unstructured{}
	vs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	vs.SetNamespace("workload")
	vs.SetName("demo")
	vs.SetUID(types.UID("1111"))
	vs.Object = map[string]any{
		"apiVersion": volsync.Group + "/" + volsync.Version,
		"kind":       "ReplicationSource",
		"metadata": map[string]any{
			"name":      "demo",
			"namespace": "workload",
		},
		"spec": map[string]any{
			"restic": map[string]any{
				"repository": "repo-secret",
			},
		},
		"status": map[string]any{
			"lastSyncTime": "2026-02-24T12:00:00Z",
			"lastSnapshot": "snap-1",
		},
	}

	sec := &corev1.Secret{}
	sec.Namespace = "workload"
	sec.Name = "repo-secret"
	sec.Data = map[string][]byte{
		"RESTIC_REPOSITORY": []byte("s3://bucket/repo"),
		"RESTIC_PASSWORD":   []byte("pass"),
	}
	sec.SetUID(types.UID("2222"))
	sec.SetResourceVersion("1")

	c := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.BackrestVolSyncBinding{}).
		WithObjects(b, vs, sec).
		Build()

	br := &fakeBackrestRepoClient{}
	r := &BackrestVolSyncBindingReconciler{
		Client: c,
		Scheme: scheme,
		BackrestClientFactory: func(_ string, _ backrest.Auth) backrestRepoClient {
			return br
		},
	}

	if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: b.Namespace, Name: b.Name}}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if br.addRepoCalls != 1 {
		t.Fatalf("expected addRepoCalls=1, got %d", br.addRepoCalls)
	}
	if len(br.taskCalls) != 0 {
		t.Fatalf("expected 0 task calls, got %d", len(br.taskCalls))
	}
}

var _ client.Object = (*v1alpha1.BackrestVolSyncBinding)(nil)
var _ metav1.Object = (*v1alpha1.BackrestVolSyncBinding)(nil)

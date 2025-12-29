package controllers

import (
	"context"
	"strings"
	"testing"

	"github.com/jogotcha/backrest-volsync-operator/api/v1alpha1"
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

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme corev1: %v", err)
	}

	// Register unstructured VolSync types for fake client.
	for _, kind := range []string{"ReplicationSource", "ReplicationDestination"} {
		gvk := schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: kind}
		scheme.AddKnownTypeWithName(gvk, &unstructured.Unstructured{})
		listGVK := schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: kind + "List"}
		scheme.AddKnownTypeWithName(listGVK, &unstructured.UnstructuredList{})
	}
	return scheme
}

func TestIsAutoBindingAllowed(t *testing.T) {
	mk := func(val string) *unstructured.Unstructured {
		u := &unstructured.Unstructured{}
		u.SetAnnotations(map[string]string{annotationAutoBinding: val})
		return u
	}

	cases := []struct {
		name   string
		policy BindingGenerationPolicy
		ann    string
		want   bool
	}{
		{"all+unset", BindingPolicyAll, "", true},
		{"all+true", BindingPolicyAll, "true", true},
		{"all+false", BindingPolicyAll, "false", false},
		{"annotated+unset", BindingPolicyAnnotated, "", false},
		{"annotated+true", BindingPolicyAnnotated, "true", true},
		{"annotated+yes", BindingPolicyAnnotated, "yes", true},
		{"annotated+1", BindingPolicyAnnotated, "1", true},
		{"annotated+false", BindingPolicyAnnotated, "false", false},
		{"annotated+0", BindingPolicyAnnotated, "0", false},
		{"annotated+no", BindingPolicyAnnotated, "no", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isAutoBindingAllowed(tc.policy, mk(tc.ann))
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestVolSyncAutoBindingReconcile_CreatesBinding(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)

	cfg := &v1alpha1.BackrestVolSyncOperatorConfig{}
	cfg.Namespace = "backrest-volsync-operator"
	cfg.Name = "backrest-volsync-operator"
	cfg.Spec.BindingGeneration.Policy = string(BindingPolicyAll)
	cfg.Spec.DefaultBackrest.URL = "http://example.invalid"

	vs := &unstructured.Unstructured{}
	vs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	vs.SetNamespace("workload")
	vs.SetName("demo")
	vs.SetUID(types.UID("1111"))

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg, vs).Build()

	r := &VolSyncAutoBindingReconciler{
		Client:         c,
		Scheme:         scheme,
		OperatorConfig: types.NamespacedName{Namespace: cfg.Namespace, Name: cfg.Name},
	}

	_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "workload", Name: "demo"}})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	var binding v1alpha1.BackrestVolSyncBinding
	if err := c.Get(ctx, types.NamespacedName{Namespace: "workload", Name: "bvsb-rs-demo"}, &binding); err != nil {
		t.Fatalf("expected binding created: %v", err)
	}
	if binding.Labels[labelManaged] != "true" {
		t.Fatalf("expected managed label")
	}
	if binding.Annotations[annotationManagedBy] != "backrest-volsync-operator" {
		t.Fatalf("expected managed-by annotation")
	}
	if binding.Annotations[annotationVolSyncRef] != "replicationsource/demo" {
		t.Fatalf("expected volsync-ref annotation")
	}
	if binding.Spec.Backrest.URL != "http://example.invalid" {
		t.Fatalf("expected backrest url propagated")
	}
	if binding.Spec.Source.Kind != "ReplicationSource" || binding.Spec.Source.Name != "demo" {
		t.Fatalf("expected source propagated")
	}
	if len(binding.OwnerReferences) != 1 {
		t.Fatalf("expected ownerref")
	}
	or := binding.OwnerReferences[0]
	if or.APIVersion != volsync.Group+"/"+volsync.Version || or.Kind != "ReplicationSource" || or.Name != "demo" {
		t.Fatalf("unexpected ownerref: %#v", or)
	}
}

func TestVolSyncAutoBindingReconcile_DoesNotTouchUserBinding(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)

	cfg := &v1alpha1.BackrestVolSyncOperatorConfig{}
	cfg.Namespace = "backrest-volsync-operator"
	cfg.Name = "backrest-volsync-operator"
	cfg.Spec.BindingGeneration.Policy = string(BindingPolicyAll)
	cfg.Spec.DefaultBackrest.URL = "http://example.invalid"

	vs := &unstructured.Unstructured{}
	vs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	vs.SetNamespace("workload")
	vs.SetName("demo")
	vs.SetUID(types.UID("1111"))

	userBinding := &v1alpha1.BackrestVolSyncBinding{}
	userBinding.Namespace = "workload"
	userBinding.Name = "bvsb-rs-demo"
	userBinding.Spec.Backrest.URL = "http://user"
	userBinding.Spec.Source = v1alpha1.VolSyncSourceRef{Kind: "ReplicationSource", Name: "demo"}
	userBinding.ObjectMeta.CreationTimestamp = metav1.Now()

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg, vs, userBinding).Build()

	r := &VolSyncAutoBindingReconciler{
		Client:         c,
		Scheme:         scheme,
		OperatorConfig: types.NamespacedName{Namespace: cfg.Namespace, Name: cfg.Name},
	}

	_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "workload", Name: "demo"}})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	var got v1alpha1.BackrestVolSyncBinding
	if err := c.Get(ctx, types.NamespacedName{Namespace: "workload", Name: "bvsb-rs-demo"}, &got); err != nil {
		t.Fatalf("get binding: %v", err)
	}
	if got.Spec.Backrest.URL != "http://user" {
		t.Fatalf("expected user binding unchanged")
	}
	if got.Labels != nil {
		if _, ok := got.Labels[labelManaged]; ok {
			t.Fatalf("expected managed label not added")
		}
	}
}

func TestDesiredBindingName_Truncates(t *testing.T) {
	long := "A" + strings.Repeat("b", 100)
	name := desiredBindingName("ReplicationSource", long)
	if len(name) > 63 {
		t.Fatalf("expected <=63")
	}
	if name == "" {
		t.Fatalf("expected non-empty")
	}
}

var _ client.Object = (*v1alpha1.BackrestVolSyncBinding)(nil)

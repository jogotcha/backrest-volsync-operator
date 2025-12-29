package controllers

import (
	"context"
	"testing"

	"github.com/jogotcha/backrest-volsync-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLoadOperatorConfig_DefaultsAndValidation(t *testing.T) {
	ctx := context.Background()

	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}

	nn := types.NamespacedName{Namespace: "ns", Name: "cfg"}

	t.Run("missing config treated as disabled", func(t *testing.T) {
		c := fake.NewClientBuilder().WithScheme(scheme).Build()
		snap, err := LoadOperatorConfig(ctx, c, nn)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if snap.BindingPolicy != BindingPolicyDisabled {
			t.Fatalf("expected policy Disabled, got %q", snap.BindingPolicy)
		}
	})

	t.Run("invalid policy errors", func(t *testing.T) {
		cfg := &v1alpha1.BackrestVolSyncOperatorConfig{}
		cfg.Namespace = nn.Namespace
		cfg.Name = nn.Name
		cfg.Spec.BindingGeneration.Policy = "Nope"
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg).Build()
		_, err := LoadOperatorConfig(ctx, c, nn)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("empty kinds allows both", func(t *testing.T) {
		cfg := &v1alpha1.BackrestVolSyncOperatorConfig{}
		cfg.Namespace = nn.Namespace
		cfg.Name = nn.Name
		cfg.Spec.BindingGeneration.Policy = string(BindingPolicyAnnotated)
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg).Build()
		snap, err := LoadOperatorConfig(ctx, c, nn)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !snap.IsVolSyncKindAllowed("ReplicationSource") || !snap.IsVolSyncKindAllowed("ReplicationDestination") {
			t.Fatalf("expected both kinds allowed")
		}
	})

	t.Run("kinds filter", func(t *testing.T) {
		cfg := &v1alpha1.BackrestVolSyncOperatorConfig{}
		cfg.Namespace = nn.Namespace
		cfg.Name = nn.Name
		cfg.Spec.BindingGeneration.Policy = string(BindingPolicyAll)
		cfg.Spec.BindingGeneration.Kinds = []string{"ReplicationSource"}
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg).Build()
		snap, err := LoadOperatorConfig(ctx, c, nn)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !snap.IsVolSyncKindAllowed("ReplicationSource") {
			t.Fatalf("expected RS allowed")
		}
		if snap.IsVolSyncKindAllowed("ReplicationDestination") {
			t.Fatalf("expected RD not allowed")
		}
	})

	t.Run("invalid kinds entry errors", func(t *testing.T) {
		cfg := &v1alpha1.BackrestVolSyncOperatorConfig{}
		cfg.Namespace = nn.Namespace
		cfg.Name = nn.Name
		cfg.Spec.BindingGeneration.Policy = string(BindingPolicyAll)
		cfg.Spec.BindingGeneration.Kinds = []string{"ReplicationSource", "BadKind"}
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg).Build()
		_, err := LoadOperatorConfig(ctx, c, nn)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("authRef only set when name present", func(t *testing.T) {
		cfg := &v1alpha1.BackrestVolSyncOperatorConfig{}
		cfg.Namespace = nn.Namespace
		cfg.Name = nn.Name
		cfg.Spec.BindingGeneration.Policy = string(BindingPolicyAll)
		cfg.Spec.DefaultBackrest.URL = "http://example"
		cfg.Spec.DefaultBackrest.AuthRef = &v1alpha1.SecretRef{Name: ""}
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg).Build()
		snap, err := LoadOperatorConfig(ctx, c, nn)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if snap.DefaultBackrestAuthRef != nil {
			t.Fatalf("expected nil authRef")
		}

		cfg2 := cfg.DeepCopy()
		cfg2.Spec.DefaultBackrest.AuthRef = &v1alpha1.SecretRef{Name: "secret"}
		c2 := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cfg2).Build()
		snap2, err := LoadOperatorConfig(ctx, c2, nn)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if snap2.DefaultBackrestAuthRef == nil || snap2.DefaultBackrestAuthRef.Name != "secret" {
			t.Fatalf("expected authRef secret")
		}
	})
}

// Ensure our tests don't accidentally rely on controller-runtime global scheme.
var _ ctrlclient.Object = (*v1alpha1.BackrestVolSyncOperatorConfig)(nil)

package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"

	"github.com/jogotcha/backrest-volsync-operator/api/v1alpha1"
	"github.com/jogotcha/backrest-volsync-operator/pkg/volsync"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	annotationAutoBinding = "backrest.garethgeorge.com/binding"
	labelManaged          = "backrest.garethgeorge.com/managed"
	annotationManagedBy   = "backrest.garethgeorge.com/managed-by"
	annotationVolSyncRef  = "backrest.garethgeorge.com/volsync-ref"
)

type VolSyncAutoBindingReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	OperatorConfig types.NamespacedName
}

func (r *VolSyncAutoBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("VolSyncAutoBindingReconciler triggered", "namespace", req.Namespace, "name", req.Name)

	cfg, err := LoadOperatorConfig(ctx, r.Client, r.OperatorConfig)
	if err != nil {
		logger.Error(err, "Failed to load operator config")
		return ctrl.Result{}, err
	}
	if cfg.Paused {
		logger.Info("Operator is paused, skipping")
		return ctrl.Result{}, nil
	}
	if cfg.BindingPolicy == BindingPolicyDisabled {
		logger.Info("Binding policy is disabled, skipping")
		return ctrl.Result{}, nil
	}

	vsObj, kind, err := r.getVolSyncObjectEither(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "Failed to get VolSync object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Found VolSync object", "kind", kind, "name", vsObj.GetName())

	if !cfg.IsVolSyncKindAllowed(kind) {
		logger.Info("Auto-binding not enabled for this kind; skipping", "kind", kind, "name", vsObj.GetName())
		return ctrl.Result{}, nil
	}

	allowed := isAutoBindingAllowed(cfg.BindingPolicy, vsObj)
	if !allowed {
		logger.Info("Auto-binding not allowed for this object", "kind", kind, "name", vsObj.GetName())
		return ctrl.Result{}, nil
	}

	if strings.TrimSpace(cfg.DefaultBackrestURL) == "" {
		logger.Info("Auto-binding enabled but defaultBackrest.url is empty; skipping", "volsyncKind", kind, "volsyncName", vsObj.GetName())
		return ctrl.Result{}, nil
	}

	bindingName := desiredBindingName(kind, vsObj.GetName())

	desired := v1alpha1.BackrestVolSyncBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
			Name:      bindingName,
			Labels: map[string]string{
				labelManaged: "true",
			},
			Annotations: map[string]string{
				annotationManagedBy:  "backrest-volsync-operator",
				annotationVolSyncRef: strings.ToLower(kind) + "/" + vsObj.GetName(),
			},
		},
		Spec: v1alpha1.BackrestVolSyncBindingSpec{
			Backrest: v1alpha1.BackrestConnection{
				URL:     cfg.DefaultBackrestURL,
				AuthRef: cfg.DefaultBackrestAuthRef,
			},
			Source: v1alpha1.VolSyncSourceRef{Kind: kind, Name: vsObj.GetName()},
			Repo:   cfg.DefaultRepo,
		},
	}

	setOwnerReference(&desired, vsObj, kind)

	var existing v1alpha1.BackrestVolSyncBinding
	err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: bindingName}, &existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.Create(ctx, &desired); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Created BackrestVolSyncBinding", "binding", desired.Name, "volsyncKind", kind, "volsyncName", vsObj.GetName())
			if r.Recorder != nil {
				r.Recorder.Eventf(vsObj, corev1.EventTypeNormal, "BindingCreated", "Created BackrestVolSyncBinding %s", desired.Name)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Do not modify user-managed bindings.
	if existing.Labels[labelManaged] != "true" {
		logger.Info("Binding exists but is not managed; skipping", "binding", existing.Name)
		return ctrl.Result{}, nil
	}

	// Patch only if spec/metadata drifted.
	original := existing.DeepCopy()
	mutated := existing.DeepCopy()
	changed := false

	if !reflect.DeepEqual(mutated.Spec, desired.Spec) {
		mutated.Spec = desired.Spec
		changed = true
	}
	if mutated.Labels == nil {
		mutated.Labels = map[string]string{}
		changed = true
	}
	if mutated.Labels[labelManaged] != "true" {
		mutated.Labels[labelManaged] = "true"
		changed = true
	}
	if mutated.Annotations == nil {
		mutated.Annotations = map[string]string{}
		changed = true
	}
	if mutated.Annotations[annotationManagedBy] != desired.Annotations[annotationManagedBy] {
		mutated.Annotations[annotationManagedBy] = desired.Annotations[annotationManagedBy]
		changed = true
	}
	if mutated.Annotations[annotationVolSyncRef] != desired.Annotations[annotationVolSyncRef] {
		mutated.Annotations[annotationVolSyncRef] = desired.Annotations[annotationVolSyncRef]
		changed = true
	}
	if !reflect.DeepEqual(mutated.OwnerReferences, desired.OwnerReferences) {
		mutated.OwnerReferences = desired.OwnerReferences
		changed = true
	}

	if !changed {
		return ctrl.Result{}, nil
	}

	patch := client.MergeFrom(original)
	if err := r.Patch(ctx, mutated, patch); err != nil {
		return ctrl.Result{}, err
	}
	if r.Recorder != nil {
		r.Recorder.Eventf(vsObj, corev1.EventTypeNormal, "BindingUpdated", "Updated BackrestVolSyncBinding %s", mutated.Name)
	}

	return ctrl.Result{}, nil
}

func (r *VolSyncAutoBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	rs := &unstructured.Unstructured{}
	rs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	rd := &unstructured.Unstructured{}
	rd.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationDestination"})

	return ctrl.NewControllerManagedBy(mgr).
		Named("volsync-autobinding").
		Watches(rs, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			return []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}}}
		})).
		Watches(rd, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			return []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}}}
		})).
		Watches(&v1alpha1.BackrestVolSyncOperatorConfig{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			if r.OperatorConfig.Name == "" || r.OperatorConfig.Namespace == "" {
				return nil
			}
			if obj.GetNamespace() != r.OperatorConfig.Namespace || obj.GetName() != r.OperatorConfig.Name {
				return nil
			}
			cfg, ok := obj.(*v1alpha1.BackrestVolSyncOperatorConfig)
			if !ok {
				return nil
			}
			if cfg.Spec.Paused {
				return nil
			}
			policy := strings.TrimSpace(cfg.Spec.BindingGeneration.Policy)
			if policy == "" || BindingGenerationPolicy(policy) == BindingPolicyDisabled {
				return nil
			}
			allowRS := true
			allowRD := true
			if len(cfg.Spec.BindingGeneration.Kinds) > 0 {
				allowRS = false
				allowRD = false
				for _, k := range cfg.Spec.BindingGeneration.Kinds {
					switch strings.TrimSpace(k) {
					case "ReplicationSource":
						allowRS = true
					case "ReplicationDestination":
						allowRD = true
					}
				}
			}
			if !allowRS && !allowRD {
				return nil
			}
			reqs := make([]reconcile.Request, 0)
			if allowRS {
				// Enqueue all ReplicationSources cluster-wide.
				var rsList unstructured.UnstructuredList
				rsList.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSourceList"})
				if err := r.List(ctx, &rsList); err == nil {
					for i := range rsList.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: rsList.Items[i].GetNamespace(), Name: rsList.Items[i].GetName()}})
					}
				}
			}
			if allowRD {
				// Enqueue all ReplicationDestinations cluster-wide.
				var rdList unstructured.UnstructuredList
				rdList.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationDestinationList"})
				if err := r.List(ctx, &rdList); err == nil {
					for i := range rdList.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: rdList.Items[i].GetNamespace(), Name: rdList.Items[i].GetName()}})
					}
				}
			}
			return reqs
		})).
		Complete(r)
}

func (r *VolSyncAutoBindingReconciler) getVolSyncObjectEither(ctx context.Context, nn types.NamespacedName) (*unstructured.Unstructured, string, error) {
	rs := &unstructured.Unstructured{}
	rs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	if err := r.Get(ctx, nn, rs); err == nil {
		return rs, "ReplicationSource", nil
	} else if !apierrors.IsNotFound(err) {
		return nil, "", err
	}

	rd := &unstructured.Unstructured{}
	rd.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationDestination"})
	if err := r.Get(ctx, nn, rd); err != nil {
		return nil, "", err
	}
	return rd, "ReplicationDestination", nil
}

func isAutoBindingAllowed(policy BindingGenerationPolicy, vsObj *unstructured.Unstructured) bool {
	v := strings.TrimSpace(vsObj.GetAnnotations()[annotationAutoBinding])
	v = strings.ToLower(v)
	if v == "false" || v == "0" || v == "no" {
		return false
	}
	if policy == BindingPolicyAll {
		return true
	}
	// Annotated policy
	return v == "true" || v == "1" || v == "yes"
}

func desiredBindingName(kind, volsyncName string) string {
	prefix := "bvsb-rs-"
	if kind == "ReplicationDestination" {
		prefix = "bvsb-rd-"
	}
	name := dns1123Safe(prefix + volsyncName)
	if len(name) <= 63 {
		return name
	}
	sum := sha256.Sum256([]byte(prefix + volsyncName))
	suffix := hex.EncodeToString(sum[:])[:8]
	trim := 63 - 1 - len(suffix)
	if trim < 1 {
		return name[:63]
	}
	return name[:trim] + "-" + suffix
}

func dns1123Safe(s string) string {
	s = strings.ToLower(s)
	sb := strings.Builder{}
	sb.Grow(len(s))
	lastDash := false
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			sb.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			sb.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(sb.String(), "-")
	if out == "" {
		return "x"
	}
	return out
}

func setOwnerReference(binding *v1alpha1.BackrestVolSyncBinding, vsObj *unstructured.Unstructured, kind string) {
	controller := true
	block := true
	binding.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         volsync.Group + "/" + volsync.Version,
			Kind:               kind,
			Name:               vsObj.GetName(),
			UID:                vsObj.GetUID(),
			Controller:         &controller,
			BlockOwnerDeletion: &block,
		},
	}
}

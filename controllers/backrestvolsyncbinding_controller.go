package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	v1 "github.com/garethgeorge/backrest/gen/go/v1"
	"github.com/jogotcha/backrest-volsync-operator/api/v1alpha1"
	"github.com/jogotcha/backrest-volsync-operator/pkg/backrest"
	"github.com/jogotcha/backrest-volsync-operator/pkg/volsync"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	conditionReady = "Ready"

	indexRepositorySecret = "status.resolvedRepositorySecret"
	indexVolSyncKey       = "spec.volsyncKey"
)

type BackrestVolSyncBindingReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	OperatorConfig types.NamespacedName
}

func (r *BackrestVolSyncBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var binding v1alpha1.BackrestVolSyncBinding
	if err := r.Get(ctx, req.NamespacedName, &binding); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if cfg, err := LoadOperatorConfig(ctx, r.Client, r.OperatorConfig); err != nil {
		return ctrl.Result{}, err
	} else if cfg.Paused {
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:               conditionReady,
			Status:             metav1.ConditionFalse,
			Reason:             "Paused",
			Message:            "Operator is paused by BackrestVolSyncOperatorConfig",
			ObservedGeneration: binding.Generation,
			LastTransitionTime: metav1.Now(),
		})
		binding.Status.ObservedGeneration = binding.Generation
		_ = r.Status().Update(ctx, &binding)
		return ctrl.Result{}, nil
	}

	// Default reconcile requeue on transient errors.
	defer func() {
		// Ensure we never accidentally log secret material via %+v.
		_ = logger
	}()

	if errs := validateBinding(&binding); len(errs) > 0 {
		err := errs.ToAggregate()
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:               conditionReady,
			Status:             metav1.ConditionFalse,
			Reason:             "InvalidSpec",
			Message:            err.Error(),
			ObservedGeneration: binding.Generation,
			LastTransitionTime: metav1.Now(),
		})
		binding.Status.ObservedGeneration = binding.Generation
		_ = r.Status().Update(ctx, &binding)
		return ctrl.Result{}, nil
	}

	vsObj, err := r.getVolSyncObject(ctx, &binding)
	if err != nil {
		return r.fail(ctx, &binding, "VolSyncNotFound", err)
	}

	repoSecretName, err := volsync.RepositorySecretName(vsObj)
	if err != nil {
		return r.fail(ctx, &binding, "VolSyncMissingRepository", err)
	}

	// Fetch repository secret (expected to be namespaced with the VolSync object).
	var repoSecret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Namespace: binding.Namespace, Name: repoSecretName}, &repoSecret); err != nil {
		return r.fail(ctx, &binding, "RepositorySecretNotFound", err)
	}

	resticRepo, resticPass, env, err := extractResticSecret(&repoSecret, binding.Spec.Repo.EnvAllowlist)
	if err != nil {
		return r.fail(ctx, &binding, "RepositorySecretInvalid", err)
	}

	inputHash := computeInputHash(&binding, vsObj, &repoSecret)
	if binding.Status.LastAppliedInputHash == inputHash && isReady(&binding) {
		// Keep status in sync for resolved secret name.
		if binding.Status.ResolvedRepositorySecret != repoSecretName {
			binding.Status.ResolvedRepositorySecret = repoSecretName
			binding.Status.ObservedGeneration = binding.Generation
			_ = r.Status().Update(ctx, &binding)
		}
		return ctrl.Result{}, nil
	}

	auth, err := r.loadBackrestAuth(ctx, &binding)
	if err != nil {
		return r.fail(ctx, &binding, "BackrestAuthInvalid", err)
	}

	repo := &v1.Repo{
		Id:             desiredRepoID(&binding),
		Uri:            resticRepo,
		Password:       resticPass,
		Env:            env,
		Flags:          append([]string(nil), binding.Spec.Repo.ExtraFlags...),
		AutoUnlock:     ptr.Deref(binding.Spec.Repo.AutoUnlock, false),
		AutoInitialize: ptr.Deref(binding.Spec.Repo.AutoInitialize, false),
	}

	// Ensure stable ordering so identical inputs do not churn.
	sort.Strings(repo.Env)
	sort.Strings(repo.Flags)

	client := backrest.New(binding.Spec.Backrest.URL, auth)
	_, err = client.AddRepo(ctx, repo)
	if err != nil {
		if isAlreadyInitializedError(err) {
			logger.Info(
				"Backrest repo already initialized; treating as applied",
				"repoID", repo.Id,
				"volsyncKind", binding.Spec.Source.Kind,
				"volsyncName", binding.Spec.Source.Name,
			)
		} else {
			return r.fail(ctx, &binding, "BackrestAddRepoFailed", err)
		}
	}

	logger.Info(
		"Backrest repo applied",
		"repoID", repo.Id,
		"volsyncKind", binding.Spec.Source.Kind,
		"volsyncName", binding.Spec.Source.Name,
	)

	now := metav1.Now()
	binding.Status.ResolvedRepositorySecret = repoSecretName
	binding.Status.LastAppliedInputHash = inputHash
	binding.Status.LastApplyTime = &now
	binding.Status.ObservedGeneration = binding.Generation
	meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
		Type:               conditionReady,
		Status:             metav1.ConditionTrue,
		Reason:             "Applied",
		Message:            "Repository registered/updated in Backrest",
		ObservedGeneration: binding.Generation,
		LastTransitionTime: now,
	})

	if err := r.Status().Update(ctx, &binding); err != nil {
		if apierrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *BackrestVolSyncBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()

	if err := mgr.GetFieldIndexer().IndexField(ctx, &v1alpha1.BackrestVolSyncBinding{}, indexRepositorySecret, func(obj client.Object) []string {
		b, ok := obj.(*v1alpha1.BackrestVolSyncBinding)
		if !ok {
			return nil
		}
		if b.Status.ResolvedRepositorySecret == "" {
			return nil
		}
		return []string{b.Status.ResolvedRepositorySecret}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &v1alpha1.BackrestVolSyncBinding{}, indexVolSyncKey, func(obj client.Object) []string {
		b, ok := obj.(*v1alpha1.BackrestVolSyncBinding)
		if !ok {
			return nil
		}
		if b.Spec.Source.Kind == "" || b.Spec.Source.Name == "" {
			return nil
		}
		return []string{strings.ToLower(b.Spec.Source.Kind) + "/" + b.Spec.Source.Name}
	}); err != nil {
		return err
	}

	rs := &unstructured.Unstructured{}
	rs.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationSource"})
	rd := &unstructured.Unstructured{}
	rd.SetGroupVersionKind(schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: "ReplicationDestination"})

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.BackrestVolSyncBinding{}).
		Watches(&v1alpha1.BackrestVolSyncOperatorConfig{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			if r.OperatorConfig.Name == "" || r.OperatorConfig.Namespace == "" {
				return nil
			}
			if obj.GetNamespace() != r.OperatorConfig.Namespace || obj.GetName() != r.OperatorConfig.Name {
				return nil
			}
			var list v1alpha1.BackrestVolSyncBindingList
			if err := r.List(ctx, &list, client.InNamespace(obj.GetNamespace())); err != nil {
				return nil
			}
			reqs := make([]reconcile.Request, 0, len(list.Items))
			for i := range list.Items {
				reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: list.Items[i].Namespace, Name: list.Items[i].Name}})
			}
			return reqs
		})).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			secret, ok := obj.(*corev1.Secret)
			if !ok {
				return nil
			}
			var list v1alpha1.BackrestVolSyncBindingList
			if err := r.List(ctx, &list, client.InNamespace(secret.Namespace), client.MatchingFields{indexRepositorySecret: secret.Name}); err != nil {
				return nil
			}
			reqs := make([]reconcile.Request, 0, len(list.Items))
			for i := range list.Items {
				reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: list.Items[i].Namespace, Name: list.Items[i].Name}})
			}
			return reqs
		})).
		Watches(
			rs,
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				u, ok := obj.(*unstructured.Unstructured)
				if !ok {
					return nil
				}
				key := strings.ToLower(u.GetKind()) + "/" + u.GetName()
				var list v1alpha1.BackrestVolSyncBindingList
				if err := r.List(ctx, &list, client.InNamespace(u.GetNamespace()), client.MatchingFields{indexVolSyncKey: key}); err != nil {
					return nil
				}
				reqs := make([]reconcile.Request, 0, len(list.Items))
				for i := range list.Items {
					reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: list.Items[i].Namespace, Name: list.Items[i].Name}})
				}
				return reqs
			}),
		).
		Watches(
			rd,
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				u, ok := obj.(*unstructured.Unstructured)
				if !ok {
					return nil
				}
				key := strings.ToLower(u.GetKind()) + "/" + u.GetName()
				var list v1alpha1.BackrestVolSyncBindingList
				if err := r.List(ctx, &list, client.InNamespace(u.GetNamespace()), client.MatchingFields{indexVolSyncKey: key}); err != nil {
					return nil
				}
				reqs := make([]reconcile.Request, 0, len(list.Items))
				for i := range list.Items {
					reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: list.Items[i].Namespace, Name: list.Items[i].Name}})
				}
				return reqs
			}),
		).
		Complete(r)
}

func (r *BackrestVolSyncBindingReconciler) getVolSyncObject(ctx context.Context, binding *v1alpha1.BackrestVolSyncBinding) (*unstructured.Unstructured, error) {
	gvk := schema.GroupVersionKind{Group: volsync.Group, Version: volsync.Version, Kind: binding.Spec.Source.Kind}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	if err := r.Get(ctx, types.NamespacedName{Namespace: binding.Namespace, Name: binding.Spec.Source.Name}, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (r *BackrestVolSyncBindingReconciler) loadBackrestAuth(ctx context.Context, binding *v1alpha1.BackrestVolSyncBinding) (backrest.Auth, error) {
	if binding.Spec.Backrest.AuthRef == nil || binding.Spec.Backrest.AuthRef.Name == "" {
		return backrest.Auth{}, nil
	}
	var sec corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Namespace: binding.Namespace, Name: binding.Spec.Backrest.AuthRef.Name}, &sec); err != nil {
		return backrest.Auth{}, err
	}
	// Supported keys:
	// - token (bearer)
	// - username/password (basic)
	if b, ok := sec.Data["token"]; ok && len(b) > 0 {
		return backrest.Auth{BearerToken: string(b)}, nil
	}
	user := string(sec.Data["username"])
	pass := string(sec.Data["password"])
	if user == "" && pass == "" {
		return backrest.Auth{}, fmt.Errorf("auth secret must contain either 'token' or 'username'/'password'")
	}
	return backrest.Auth{BasicUsername: user, BasicPassword: pass}, nil
}

func (r *BackrestVolSyncBindingReconciler) fail(ctx context.Context, binding *v1alpha1.BackrestVolSyncBinding, reason string, err error) (ctrl.Result, error) {
	errHash := hashString(err.Error())
	binding.Status.LastErrorHash = errHash
	log.FromContext(ctx).Info(
		"Reconcile failed",
		"reason", reason,
		"namespace", binding.Namespace,
		"name", binding.Name,
		"errorHash", errHash,
	)
	meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
		Type:               conditionReady,
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            fmt.Sprintf("%s (details omitted; errorHash=%s)", reason, errHash),
		ObservedGeneration: binding.Generation,
		LastTransitionTime: metav1.Now(),
	})
	binding.Status.ObservedGeneration = binding.Generation
	_ = r.Status().Update(ctx, binding)
	// Requeue with backoff handled by controller-runtime.
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func isAlreadyInitializedError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already initialized")
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func validateBinding(b *v1alpha1.BackrestVolSyncBinding) field.ErrorList {
	var errs field.ErrorList
	if b.Spec.Backrest.URL == "" {
		errs = append(errs, field.Required(field.NewPath("spec", "backrest", "url"), "required"))
	}
	if b.Spec.Source.Kind != "ReplicationSource" && b.Spec.Source.Kind != "ReplicationDestination" {
		errs = append(errs, field.Invalid(field.NewPath("spec", "source", "kind"), b.Spec.Source.Kind, "must be ReplicationSource or ReplicationDestination"))
	}
	if b.Spec.Source.Name == "" {
		errs = append(errs, field.Required(field.NewPath("spec", "source", "name"), "required"))
	}
	return errs
}

func desiredRepoID(b *v1alpha1.BackrestVolSyncBinding) string {
	if b.Spec.Repo.IDOverride != "" {
		return b.Spec.Repo.IDOverride
	}
	return fmt.Sprintf("volsync-%s-%s-%s", b.Namespace, strings.ToLower(b.Spec.Source.Kind), b.Spec.Source.Name)
}

func extractResticSecret(sec *corev1.Secret, allowlist []string) (repo string, pass string, env []string, err error) {
	repo = strings.TrimSpace(string(sec.Data["RESTIC_REPOSITORY"]))
	pass = string(sec.Data["RESTIC_PASSWORD"])
	if repo == "" {
		return "", "", nil, fmt.Errorf("missing RESTIC_REPOSITORY")
	}
	if pass == "" {
		return "", "", nil, fmt.Errorf("missing RESTIC_PASSWORD")
	}

	allowed := map[string]struct{}{}
	if len(allowlist) > 0 {
		for _, k := range allowlist {
			allowed[k] = struct{}{}
		}
	}

	for k, v := range sec.Data {
		if k == "RESTIC_REPOSITORY" || k == "RESTIC_PASSWORD" {
			continue
		}
		if strings.HasPrefix(k, "RESTIC_") {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[k]; !ok {
				continue
			}
		}
		env = append(env, k+"="+string(v))
	}
	return repo, pass, env, nil
}

func computeInputHash(binding *v1alpha1.BackrestVolSyncBinding, vsObj *unstructured.Unstructured, sec *corev1.Secret) string {
	h := sha256.New()
	write := func(s string) {
		_, _ = h.Write([]byte(s))
		_, _ = h.Write([]byte{0})
	}
	write(binding.Spec.Backrest.URL)
	write(binding.Spec.Source.Kind)
	write(binding.Spec.Source.Name)
	write(desiredRepoID(binding))
	write(fmt.Sprintf("autoUnlock=%v", ptr.Deref(binding.Spec.Repo.AutoUnlock, false)))
	write(fmt.Sprintf("autoInitialize=%v", ptr.Deref(binding.Spec.Repo.AutoInitialize, false)))
	flags := append([]string(nil), binding.Spec.Repo.ExtraFlags...)
	sort.Strings(flags)
	write("flags=" + strings.Join(flags, ","))
	allow := append([]string(nil), binding.Spec.Repo.EnvAllowlist...)
	sort.Strings(allow)
	write("envAllowlist=" + strings.Join(allow, ","))
	write("volsync.uid=" + string(vsObj.GetUID()))
	write("volsync.rv=" + vsObj.GetResourceVersion())
	write("secret.uid=" + string(sec.GetUID()))
	write("secret.rv=" + sec.GetResourceVersion())
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func isReady(b *v1alpha1.BackrestVolSyncBinding) bool {
	cond := meta.FindStatusCondition(b.Status.Conditions, conditionReady)
	return cond != nil && cond.Status == metav1.ConditionTrue
}

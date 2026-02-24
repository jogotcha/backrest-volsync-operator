package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type BackrestVolSyncOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackrestVolSyncOperatorConfigSpec   `json:"spec,omitempty"`
	Status BackrestVolSyncOperatorConfigStatus `json:"status,omitempty"`
}

type BackrestVolSyncOperatorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackrestVolSyncOperatorConfig `json:"items"`
}

type BackrestVolSyncOperatorConfigSpec struct {
	// Paused disables all controller side-effects (including Backrest API calls and auto-binding creation).
	Paused bool `json:"paused,omitempty"`

	// DefaultBackrest provides defaults used by auto-binding generation.
	DefaultBackrest BackrestConnection `json:"defaultBackrest,omitempty"`

	BindingGeneration BindingGenerationSpec `json:"bindingGeneration,omitempty"`
}

type BindingGenerationSpec struct {
	// Policy controls auto-creation of BackrestVolSyncBindings from VolSync objects.
	//
	// Allowed values:
	// - Disabled: do not auto-create bindings
	// - Annotated: only auto-create when the VolSync object has annotation backrest.garethgeorge.com/binding="true"
	// - All: auto-create for all VolSync objects (unless explicitly opted out)
	Policy string `json:"policy,omitempty"`

	// Kinds restricts which VolSync object kinds are eligible for auto-binding.
	//
	// Allowed values:
	// - ReplicationSource
	// - ReplicationDestination
	//
	// If omitted or empty, both kinds are allowed.
	Kinds []string `json:"kinds,omitempty"`

	// DefaultRepo provides defaults for generated bindings. Fields are optional.
	DefaultRepo BackrestRepoSpec `json:"defaultRepo,omitempty"`
}

type BackrestVolSyncOperatorConfigStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

func init() {
	SchemeBuilder.Register(&BackrestVolSyncOperatorConfig{}, &BackrestVolSyncOperatorConfigList{})
}

// DeepCopyInto, DeepCopy, and DeepCopyObject are implemented manually to avoid requiring codegen.
func (in *BackrestVolSyncOperatorConfig) DeepCopyInto(out *BackrestVolSyncOperatorConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = BackrestVolSyncOperatorConfigStatus{ObservedGeneration: in.Status.ObservedGeneration}
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
	if in.Spec.DefaultBackrest.AuthRef != nil {
		out.Spec.DefaultBackrest.AuthRef = &SecretRef{Name: in.Spec.DefaultBackrest.AuthRef.Name}
	}
	if in.Spec.BindingGeneration.DefaultRepo.ExtraFlags != nil {
		out.Spec.BindingGeneration.DefaultRepo.ExtraFlags = append([]string(nil), in.Spec.BindingGeneration.DefaultRepo.ExtraFlags...)
	}
	if in.Spec.BindingGeneration.Kinds != nil {
		out.Spec.BindingGeneration.Kinds = append([]string(nil), in.Spec.BindingGeneration.Kinds...)
	}
	if in.Spec.BindingGeneration.DefaultRepo.EnvAllowlist != nil {
		out.Spec.BindingGeneration.DefaultRepo.EnvAllowlist = append([]string(nil), in.Spec.BindingGeneration.DefaultRepo.EnvAllowlist...)
	}
	if in.Spec.BindingGeneration.DefaultRepo.AutoUnlock != nil {
		v := *in.Spec.BindingGeneration.DefaultRepo.AutoUnlock
		out.Spec.BindingGeneration.DefaultRepo.AutoUnlock = &v
	}
	if in.Spec.BindingGeneration.DefaultRepo.AutoInitialize != nil {
		v := *in.Spec.BindingGeneration.DefaultRepo.AutoInitialize
		out.Spec.BindingGeneration.DefaultRepo.AutoInitialize = &v
	}
	if in.Spec.BindingGeneration.DefaultRepo.TriggerTasksOnSnapshot != nil {
		v := *in.Spec.BindingGeneration.DefaultRepo.TriggerTasksOnSnapshot
		out.Spec.BindingGeneration.DefaultRepo.TriggerTasksOnSnapshot = &v
	}
}

func (in *BackrestVolSyncOperatorConfig) DeepCopy() *BackrestVolSyncOperatorConfig {
	if in == nil {
		return nil
	}
	out := new(BackrestVolSyncOperatorConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *BackrestVolSyncOperatorConfig) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

func (in *BackrestVolSyncOperatorConfigList) DeepCopyInto(out *BackrestVolSyncOperatorConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]BackrestVolSyncOperatorConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *BackrestVolSyncOperatorConfigList) DeepCopy() *BackrestVolSyncOperatorConfigList {
	if in == nil {
		return nil
	}
	out := new(BackrestVolSyncOperatorConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *BackrestVolSyncOperatorConfigList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

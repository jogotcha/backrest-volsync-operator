package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type BackrestVolSyncBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackrestVolSyncBindingSpec   `json:"spec,omitempty"`
	Status BackrestVolSyncBindingStatus `json:"status,omitempty"`
}

type BackrestVolSyncBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackrestVolSyncBinding `json:"items"`
}

type BackrestVolSyncBindingSpec struct {
	Backrest BackrestConnection `json:"backrest"`
	Source   VolSyncSourceRef   `json:"source"`
	Repo     BackrestRepoSpec   `json:"repo,omitempty"`
}

type BackrestConnection struct {
	URL     string     `json:"url"`
	AuthRef *SecretRef `json:"authRef,omitempty"`
}

type SecretRef struct {
	Name string `json:"name"`
}

type VolSyncSourceRef struct {
	// Kind must be either "ReplicationSource" or "ReplicationDestination".
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type BackrestRepoSpec struct {
	IDOverride string `json:"idOverride,omitempty"`
	// AutoUnlock controls Backrest's repo auto-unlock behavior.
	//
	// When enabled, Backrest will remove restic lockfiles at the start of forget and prune operations.
	// This can be unsafe if the repository is shared by multiple client devices.
	// Disabled by default.
	AutoUnlock     *bool `json:"autoUnlock,omitempty"`
	AutoInitialize *bool `json:"autoInitialize,omitempty"`
	// TriggerTasksOnSnapshot enables enqueueing Backrest INDEX_SNAPSHOTS and STATS
	// tasks when a bound ReplicationSource reports a new completed snapshot/sync marker.
	// Disabled by default.
	TriggerTasksOnSnapshot *bool    `json:"triggerTasksOnSnapshot,omitempty"`
	ExtraFlags             []string `json:"extraFlags,omitempty"`
	EnvAllowlist           []string `json:"envAllowlist,omitempty"`
}

type BackrestVolSyncBindingStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`

	ResolvedRepositorySecret string       `json:"resolvedRepositorySecret,omitempty"`
	LastAppliedInputHash     string       `json:"lastAppliedInputHash,omitempty"`
	LastApplyTime            *metav1.Time `json:"lastApplyTime,omitempty"`
	LastErrorHash            string       `json:"lastErrorHash,omitempty"`
	LastSnapshotMarker       string       `json:"lastSnapshotMarker,omitempty"`
	LastRepoTaskTriggerTime  *metav1.Time `json:"lastRepoTaskTriggerTime,omitempty"`
	LastRepoTaskErrorHash    string       `json:"lastRepoTaskErrorHash,omitempty"`
}

func init() {
	SchemeBuilder.Register(&BackrestVolSyncBinding{}, &BackrestVolSyncBindingList{})
}

// DeepCopyInto, DeepCopy, and DeepCopyObject are implemented manually to avoid requiring codegen.

func (in *BackrestVolSyncBinding) DeepCopyInto(out *BackrestVolSyncBinding) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = BackrestVolSyncBindingStatus{
		ObservedGeneration:       in.Status.ObservedGeneration,
		ResolvedRepositorySecret: in.Status.ResolvedRepositorySecret,
		LastAppliedInputHash:     in.Status.LastAppliedInputHash,
		LastErrorHash:            in.Status.LastErrorHash,
		LastSnapshotMarker:       in.Status.LastSnapshotMarker,
		LastRepoTaskErrorHash:    in.Status.LastRepoTaskErrorHash,
	}
	if in.Status.LastApplyTime != nil {
		out.Status.LastApplyTime = in.Status.LastApplyTime.DeepCopy()
	}
	if in.Status.LastRepoTaskTriggerTime != nil {
		out.Status.LastRepoTaskTriggerTime = in.Status.LastRepoTaskTriggerTime.DeepCopy()
	}
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
	if in.Spec.Repo.ExtraFlags != nil {
		out.Spec.Repo.ExtraFlags = append([]string(nil), in.Spec.Repo.ExtraFlags...)
	}
	if in.Spec.Repo.EnvAllowlist != nil {
		out.Spec.Repo.EnvAllowlist = append([]string(nil), in.Spec.Repo.EnvAllowlist...)
	}
	if in.Spec.Backrest.AuthRef != nil {
		out.Spec.Backrest.AuthRef = &SecretRef{Name: in.Spec.Backrest.AuthRef.Name}
	}
	if in.Spec.Repo.AutoUnlock != nil {
		v := *in.Spec.Repo.AutoUnlock
		out.Spec.Repo.AutoUnlock = &v
	}
	if in.Spec.Repo.AutoInitialize != nil {
		v := *in.Spec.Repo.AutoInitialize
		out.Spec.Repo.AutoInitialize = &v
	}
	if in.Spec.Repo.TriggerTasksOnSnapshot != nil {
		v := *in.Spec.Repo.TriggerTasksOnSnapshot
		out.Spec.Repo.TriggerTasksOnSnapshot = &v
	}
}

func (in *BackrestVolSyncBinding) DeepCopy() *BackrestVolSyncBinding {
	if in == nil {
		return nil
	}
	out := new(BackrestVolSyncBinding)
	in.DeepCopyInto(out)
	return out
}

func (in *BackrestVolSyncBinding) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

func (in *BackrestVolSyncBindingList) DeepCopyInto(out *BackrestVolSyncBindingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]BackrestVolSyncBinding, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *BackrestVolSyncBindingList) DeepCopy() *BackrestVolSyncBindingList {
	if in == nil {
		return nil
	}
	out := new(BackrestVolSyncBindingList)
	in.DeepCopyInto(out)
	return out
}

func (in *BackrestVolSyncBindingList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

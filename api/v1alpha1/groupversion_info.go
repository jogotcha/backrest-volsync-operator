package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Group   = "backrest.garethgeorge.com"
	Version = "v1alpha1"
)

var (
	GroupVersion  = schema.GroupVersion{Group: Group, Version: Version}
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(GroupVersion,
		&BackrestVolSyncBinding{},
		&BackrestVolSyncBindingList{},
		&BackrestVolSyncOperatorConfig{},
		&BackrestVolSyncOperatorConfigList{},
	)
	metav1.AddToGroupVersion(s, GroupVersion)
	return nil
}

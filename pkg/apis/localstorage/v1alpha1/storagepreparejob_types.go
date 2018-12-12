package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StoragePrepareJobSpec defines the desired state of StoragePrepareJob
type StoragePrepareJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// StoragePrepareJobStatus defines the observed state of StoragePrepareJob
type StoragePrepareJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StoragePrepareJob is the Schema for the storagepreparejobs API
// +k8s:openapi-gen=true
type StoragePrepareJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoragePrepareJobSpec   `json:"spec,omitempty"`
	Status StoragePrepareJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StoragePrepareJobList contains a list of StoragePrepareJob
type StoragePrepareJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StoragePrepareJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StoragePrepareJob{}, &StoragePrepareJobList{})
}

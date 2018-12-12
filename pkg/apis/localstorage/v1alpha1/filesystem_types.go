package v1alpha1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FilesystemSpec defines the desired state of Filesystem
type FilesystemSpec struct {
	NodeName     string            `json:"nodeName"`
	DevicePath   string            `json:"device"`
	Type         string            `json:"type"`
	MountPath    string            `json:"mountPath"`
	MountOptions string            `json:"mountOptions"`
	MountEnabled bool              `json:"mountEnabled"`
	Capacity     resource.Quantity `json:"capacity"`
}

// FilesystemStatus defines the observed state of Filesystem
type FilesystemStatus struct {
	PreparePhase StoragePreparePhase `json:"preparePhase"`
	Mounted      bool                `json:"mounted"`
}

func (f *Filesystem) AsPv(storageClass string) *corev1.PersistentVolume {
	pv := &corev1.PersistentVolume{}
	volumeMode := corev1.PersistentVolumeFilesystem
	reclaimPolicy := corev1.PersistentVolumeReclaimRetain

	pv.APIVersion = "v1"
	pv.Kind = "PersistentVolume"

	pv.Name = f.GetName()
	pv.Labels = f.GetLabels()
	pv.Annotations = f.GetAnnotations()
	pv.Spec.Capacity = corev1.ResourceList{
		corev1.ResourceStorage: f.Spec.Capacity,
	}
	pv.Spec.VolumeMode = &volumeMode
	pv.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	pv.Spec.PersistentVolumeReclaimPolicy = reclaimPolicy
	pv.Spec.StorageClassName = storageClass
	pv.Spec.Local = &corev1.LocalVolumeSource{
		Path: f.Spec.MountPath,
	}
	pv.Spec.NodeAffinity = &corev1.VolumeNodeAffinity{
		Required: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						corev1.NodeSelectorRequirement{
							Key:      "kubernetes.io/hostname",
							Operator: "In",
							Values: []string{
								f.Spec.NodeName,
							},
						},
					},
				},
			},
		}}

	return pv
}

func (f *Filesystem) GetEnv(prefix string) []corev1.EnvVar {
	prefix = strings.ToUpper(prefix)
	result := []corev1.EnvVar{}
	return result
}

func (f *Filesystem) GetPreparePhase() StoragePreparePhase {
	return f.Status.PreparePhase
}

func (f *Filesystem) SetPreparePhase(phase StoragePreparePhase) {
	f.Status.PreparePhase = phase
}

func (f *Filesystem) IsMountable() bool {
	return f.Spec.MountEnabled
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Filesystem is the Schema for the filesystems API
// +k8s:openapi-gen=true
type Filesystem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FilesystemSpec   `json:"spec,omitempty"`
	Status FilesystemStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FilesystemList contains a list of Filesystem
type FilesystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Filesystem `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Filesystem{}, &FilesystemList{})
}

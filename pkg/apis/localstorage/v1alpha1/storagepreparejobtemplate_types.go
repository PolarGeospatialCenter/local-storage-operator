package v1alpha1

import (
	"fmt"

	"github.com/PolarGeospatialCenter/local-storage-operator/pkg/pvctemplate"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StoragePrepareJobTemplateSpec defines the desired state of StoragePrepareJobTemplate
type StoragePrepareJobTemplateSpec struct {
	Selector   metav1.LabelSelector `json:"selector"`
	Type       metav1.TypeMeta      `json:"type"`
	Template   batchv1.JobSpec      `json:"template"`
	VolumeName string               `json:"volumeName"`
}

// StoragePrepareJobTemplateStatus defines the observed state of StoragePrepareJobTemplate
type StoragePrepareJobTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StoragePrepareJobTemplate is the Schema for the storagepreparejobtemplates API
// +k8s:openapi-gen=true
type StoragePrepareJobTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoragePrepareJobTemplateSpec   `json:"spec,omitempty"`
	Status StoragePrepareJobTemplateStatus `json:"status,omitempty"`
}

type prepareObject interface {
	GetObjectKind() schema.ObjectKind
	GetLabels() map[string]string
}

var volumeModeMap = map[string]map[string]corev1.PersistentVolumeMode{
	"localstorage.k8s.pgc.umn.edu/v1alpha1": map[string]corev1.PersistentVolumeMode{
		"Disk":       corev1.PersistentVolumeBlock,
		"Filesystem": corev1.PersistentVolumeFilesystem,
	},
}

// CreatePrepareJob returns a Job from a PersistentVolumeClaim
func (s *StoragePrepareJobTemplate) CreatePrepareJob(pvc corev1.PersistentVolumeClaim, jobNameSuffix string) *batchv1.Job {

	job := &batchv1.Job{}
	job.APIVersion = "batch/v1"
	job.Kind = "Job"

	job.Name = fmt.Sprintf("%s-%s", s.Name, jobNameSuffix)
	job.Spec = *s.Spec.Template.DeepCopy()
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, pvctemplate.CreateVolume(s, pvc))

	return job
}

func (s *StoragePrepareJobTemplate) GetVolumeClaimTemplate() *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{}
	pvc.APIVersion = "v1"
	pvc.Kind = "PersistentVolumeClaim"
	pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
		corev1.ReadWriteOnce,
	}
	pvc.Spec.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse("1"),
		},
	}
	pvc.Name = s.Spec.VolumeName

	if apiVersionVmMap, ok := volumeModeMap[s.Spec.Type.APIVersion]; ok {
		if vm, ok := apiVersionVmMap[s.Spec.Type.Kind]; ok {
			pvc.Spec.VolumeMode = &vm
		}
	}

	return pvc
}

func (_ *StoragePrepareJobTemplate) SetPvcMetadata(_ *corev1.PersistentVolumeClaim) {
	return
}

func (s *StoragePrepareJobTemplate) GetVolumeName() string {
	return s.Spec.VolumeName
}

func (l *StoragePrepareJobTemplateList) FindMatchingJob(o prepareObject) (*StoragePrepareJobTemplate, error) {
	var prepareJob *StoragePrepareJobTemplate

	for _, job := range l.Items {
		if job.Matches(o) {
			if prepareJob == nil {
				prepareJob = job.DeepCopy()
			} else {
				return nil, fmt.Errorf("more than matching prepare job found for object")
			}
		}
	}

	return prepareJob, nil
}

func (s *StoragePrepareJobTemplate) Matches(o prepareObject) bool {
	typeMeta := o.GetObjectKind()
	groupVersionKind := typeMeta.GroupVersionKind()
	oAPIVersion, oKind := groupVersionKind.ToAPIVersionAndKind()

	if oAPIVersion != s.Spec.Type.APIVersion || oKind != s.Spec.Type.Kind {
		return false
	}

	sel, err := metav1.LabelSelectorAsSelector(&s.Spec.Selector)

	if err != nil {
		return false
	}

	return sel.Matches(labels.Set(o.GetLabels()))
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StoragePrepareJobTemplateList contains a list of StoragePrepareJobTemplate
type StoragePrepareJobTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StoragePrepareJobTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StoragePrepareJobTemplate{}, &StoragePrepareJobTemplateList{})
}

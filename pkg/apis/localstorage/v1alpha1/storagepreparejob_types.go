package v1alpha1

import (
	"fmt"

	"github.com/PolarGeospatialCenter/local-storage-operator/pkg/pvctemplate"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StoragePrepareJobSpec defines the desired state of StoragePrepareJob
type StoragePrepareJobSpec struct {
	Pv  corev1.PersistentVolume      `json:"pv"`
	Pvc corev1.PersistentVolumeClaim `json:"pvc"`
	Job batchv1.Job                  `json:"job"`
}

type StoragePrepareJobPhase string

const (
	StoragePrepareJobPhasePending   StoragePrepareJobPhase = "pending"
	StoragePrepareJobPhaseRunning   StoragePrepareJobPhase = "running"
	StoragePrepareJobPhaseFailed    StoragePrepareJobPhase = "failed"
	StoragePrepareJobPhaseSucceeded StoragePrepareJobPhase = "succeeded"
)

// StoragePrepareJobStatus defines the observed state of StoragePrepareJob
type StoragePrepareJobStatus struct {
	Phase StoragePrepareJobPhase `json:"phase"`
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

type PreparableStorageObject interface {
	AsPv() *corev1.PersistentVolume
	SetStorageClass(string)
	GetName() string
	GetEnv(string) []corev1.EnvVar
	GetPvLabelSelector() *metav1.LabelSelector
}

func NewStoragePrepareJob(template *StoragePrepareJobTemplate, d PreparableStorageObject, namespace string) *StoragePrepareJob {
	spj := &StoragePrepareJob{}
	spj.APIVersion = "localstorage.k8s.pgc.umn.edu/v1alpha1"
	spj.Kind = "StoragePrepareJob"
	spj.Name = d.GetName()
	spj.Namespace = namespace
	spj.Status.Phase = "pending"

	d.SetStorageClass("prepare-local-storage")
	prepPv := d.AsPv()
	prepPv.Name = fmt.Sprintf("prepare-%s", d.GetName())
	prepPv.Labels = nil
	prepPv.Annotations = nil
	spj.Spec.Pv = *prepPv

	prepPvc := pvctemplate.CreatePVC(template, *prepPv)
	prepPvc.Namespace = namespace
	prepPvc.Annotations = nil
	prepPvc.Spec.StorageClassName = &prepPv.Spec.StorageClassName
	spj.Spec.Pvc = *prepPvc

	job := template.CreatePrepareJob(*prepPvc, d.GetName())
	job.Namespace = namespace

	for i, _ := range job.Spec.Template.Spec.Containers {
		job.Spec.Template.Spec.Containers[i].Env = append(job.Spec.Template.Spec.Containers[i].Env, d.GetEnv("STORAGE")...)
		ls := d.GetPvLabelSelector()
		job.Spec.Template.Spec.Containers[i].Env = append(job.Spec.Template.Spec.Containers[i].Env,
			corev1.EnvVar{
				Name:  "PV_LABEL_SELECTOR",
				Value: metav1.FormatLabelSelector(ls),
			},
		)

	}

	spj.Spec.Job = *job

	return spj
}

func (s *StoragePrepareJob) UpdateOwnerReferences() error {

	if s.UID == "" {
		return fmt.Errorf("uid must not be set")
	}

	or := &metav1.OwnerReference{}
	or.APIVersion = s.APIVersion
	or.Kind = s.Kind
	or.Name = s.Name
	or.UID = s.UID

	s.Spec.Pv.OwnerReferences = append(s.Spec.Pv.OwnerReferences, *or)
	s.Spec.Pvc.OwnerReferences = append(s.Spec.Pvc.OwnerReferences, *or)
	s.Spec.Job.OwnerReferences = append(s.Spec.Job.OwnerReferences, *or)

	return nil
}

func (s *StoragePrepareJob) GetObjects() []runtime.Object {
	return []runtime.Object{
		&s.Spec.Pv,
		&s.Spec.Pvc,
		&s.Spec.Job,
	}
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

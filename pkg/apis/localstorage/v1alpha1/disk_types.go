package v1alpha1

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DiskSpec defines the desired state of Disk
type DiskSpec struct {
	Info         DiskInfo     `json:"diskInfo"`
	Location     DiskLocation `json:"location"`
	StorageClass string       `json:"storageClass"`
	Enabled      bool         `json:"enabled"`
}

type DiskInfo struct {
	Wwn            string            `json:"wwn"`
	Model          string            `json:"model"`
	SerialNumber   string            `json:"serialNumber"`
	Capacity       resource.Quantity `json:"capacity"`
	UdevProperties map[string]string `json:"udevProperties"`
	UdevAttributes map[string]string `json:"udevAttributes"`
}

type DiskLocation struct {
	NodeName      string `json:"node"`
	Backplane     string `json:"backplane"`
	Slot          string `json:"slot"`
	AdapterDriver string `json:"adapterDriver,omitempty"`
}

// DiskStatus defines the observed state of Disk
type DiskStatus struct {
	PreparePhase StoragePreparePhase `json:"preparePhase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Disk is the Schema for the disks API
// +k8s:openapi-gen=true
type Disk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiskSpec   `json:"spec,omitempty"`
	Status DiskStatus `json:"status,omitempty"`
}

//Init initializes a disk.
func (d *Disk) Init(name string) {
	typeMeta := metav1.TypeMeta{
		APIVersion: "localstorage.k8s.pgc.umn.edu/v1alpha1",
		Kind:       "Disk",
	}

	objectMeta := metav1.ObjectMeta{
		Name: name,
	}

	d.TypeMeta = typeMeta
	d.ObjectMeta = objectMeta
}

func (d *Disk) GetStorageClass() string {
	if d.Spec.StorageClass == "" {
		return "local-storage"
	}
	return d.Spec.StorageClass
}

func (d *Disk) SetStorageClass(storageClass string) {
	d.Spec.StorageClass = storageClass
}

//Equals checks if a disk is equal to another disk.
func (d *Disk) Equals(d2 *Disk) bool {

	return reflect.DeepEqual(d.Spec, d2.Spec)
}

//UpdateInfo updates the Info of a disk.
func (d *DiskInfo) UpdateInfo(wwn, model, sn, capacity string) error {
	qty, err := resource.ParseQuantity(capacity)
	if err != nil {
		return err
	}
	d.Wwn = wwn
	d.Model = model
	d.SerialNumber = sn
	d.Capacity = qty
	return nil
}

//UpdateLocation updates the location of a disk.
func (d *DiskLocation) UpdateLocation(nodeName, backPlane, slot, adapterDriver string) error {
	d.NodeName = nodeName
	d.Backplane = backPlane
	d.Slot = slot
	d.AdapterDriver = adapterDriver
	return nil
}

// UpdateLabels sets convenience labels from disk data.
func (d *Disk) UpdateLabels() {
	if d.Labels == nil {
		d.Labels = make(map[string]string)
	}

	d.Labels[fmt.Sprintf("%s/DiskModel", SchemeGroupVersion.Group)] = d.Spec.Info.Model
	d.Labels[fmt.Sprintf("%s/DiskNode", SchemeGroupVersion.Group)] = d.Spec.Location.NodeName
	d.Labels[fmt.Sprintf("%s/DiskBackplane", SchemeGroupVersion.Group)] = d.Spec.Location.Backplane
	d.Labels[fmt.Sprintf("%s/DiskSlot", SchemeGroupVersion.Group)] = d.Spec.Location.Slot
	d.Labels[fmt.Sprintf("%s/DiskDriver", SchemeGroupVersion.Group)] = d.Spec.Location.AdapterDriver
	d.Labels[fmt.Sprintf("%s/DiskCapacityTB", SchemeGroupVersion.Group)] = fmt.Sprintf("%d", d.CapacityTB())
}

// CapacityTB returns capacity rounded to the nearest TB
func (d Disk) CapacityTB() int {
	bytes, _ := d.Spec.Info.Capacity.AsInt64()
	return int(math.Round(float64(bytes) / (1000000000000)))
}

//Valid checks if disk struct is valid.
func (d Disk) Valid() bool {

	if d.Spec.Info.Wwn != "" &&
		d.Spec.Info.SerialNumber != "" &&
		d.Spec.Info.Model != "" &&
		d.Spec.Location.NodeName != "" &&
		d.Spec.Location.Backplane != "" &&
		d.Spec.Location.Slot != "" &&
		d.Spec.Location.AdapterDriver != "" {
		return true
	}

	return false
}

//Populated checks if the disk is populated.
func (d Disk) Populated() bool {
	return d.Spec.Info.Wwn != "" && d.Spec.Location.Backplane != "" && d.Spec.Location.Slot != ""
}

//Name computes the name of a disk.
func (d Disk) Name() string {
	return fmt.Sprintf("wwn-%s", d.Spec.Info.Wwn)
}

//String returns a human readable string of the disk struct.
func (d Disk) String() string {
	return fmt.Sprintf("Name: %v - WWN: %v - Model: %v - S/N: %v - Adapter: %v - Backplane: %v - Slot: %v - NodeName: %v - Capacity: %v", d.Name(), d.Spec.Info.Wwn, d.Spec.Info.Model, d.Spec.Info.SerialNumber, d.Spec.Location.AdapterDriver, d.Spec.Location.Backplane, d.Spec.Location.Slot, d.Spec.Location.NodeName, d.Spec.Info.Capacity)
}

func (d Disk) devicePath() string {
	return fmt.Sprintf("/dev/disk/by-id/wwn-%s", d.Spec.Info.Wwn)
}

func (d *Disk) AsPv() *corev1.PersistentVolume {
	pv := &corev1.PersistentVolume{}
	volumeMode := corev1.PersistentVolumeBlock
	reclaimPolicy := corev1.PersistentVolumeReclaimRetain

	pv.APIVersion = "v1"
	pv.Kind = "PersistentVolume"

	pv.Name = d.Name()
	pv.Labels = d.GetLabels()
	pv.Annotations = d.GetAnnotations()
	pv.Spec.Capacity = corev1.ResourceList{
		corev1.ResourceStorage: d.Spec.Info.Capacity,
	}
	pv.Spec.VolumeMode = &volumeMode
	pv.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	pv.Spec.PersistentVolumeReclaimPolicy = reclaimPolicy
	pv.Spec.StorageClassName = d.GetStorageClass()
	pv.Spec.Local = &corev1.LocalVolumeSource{
		Path: d.devicePath(),
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
								d.Spec.Location.NodeName,
							},
						},
					},
				},
			},
		}}

	return pv
}

func (d *Disk) GetEnv(prefix string) []corev1.EnvVar {
	prefix = strings.ToUpper(prefix)
	result := make([]corev1.EnvVar, 0, len(d.Spec.Info.UdevProperties)+len(d.Spec.Info.UdevAttributes))

	for k, v := range d.Spec.Info.UdevAttributes {
		result = append(result, corev1.EnvVar{
			Name:  fmt.Sprintf("%s_%s", prefix, strings.ToUpper(k)),
			Value: v,
		})
	}

	for k, v := range d.Spec.Info.UdevProperties {
		result = append(result, corev1.EnvVar{
			Name:  fmt.Sprintf("%s_%s", prefix, strings.ToUpper(k)),
			Value: v,
		})
	}

	result = append(result, corev1.EnvVar{
		Name:  fmt.Sprintf("%s_%s", prefix, "BACKPLANE"),
		Value: d.Spec.Location.Backplane,
	})

	result = append(result, corev1.EnvVar{
		Name:  fmt.Sprintf("%s_%s", prefix, "SLOT"),
		Value: d.Spec.Location.Slot,
	})

	return result
}

func (d *Disk) GetPreparePhase() StoragePreparePhase {
	return d.Status.PreparePhase
}

func (d *Disk) SetPreparePhase(phase StoragePreparePhase) {
	d.Status.PreparePhase = phase
}

func (d *Disk) GetEnabled() bool {
	return d.Spec.Enabled
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DiskList contains a list of Disk
type DiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Disk `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Disk{}, &DiskList{})
}

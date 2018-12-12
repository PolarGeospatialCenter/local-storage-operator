package pvctemplate

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type Getter interface {
	GetVolumeClaimTemplate() *corev1.PersistentVolumeClaim // Returns a deep copy of the claim.
	GetName() string
	GetVolumeName() string
	SetPvcMetadata(*corev1.PersistentVolumeClaim)
}

// CreatePVC returns a PersistentVolumeClaim from a PersistentVolume
func CreatePVC(g Getter, pv corev1.PersistentVolume) *corev1.PersistentVolumeClaim {
	pvc := g.GetVolumeClaimTemplate()
	pvc.Name = fmt.Sprintf("%s-%s", g.GetName(), pv.Name)
	pvc.Spec.VolumeName = pv.Name

	g.SetPvcMetadata(pvc)

	return pvc
}

func CreateVolume(g Getter, pvc corev1.PersistentVolumeClaim) corev1.Volume {
	v := corev1.Volume{}
	v.Name = g.GetVolumeName()
	v.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: pvc.Name,
	}

	return v
}

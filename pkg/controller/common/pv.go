package common

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
)

func checkPvEqual(pv1, pv2 *corev1.PersistentVolume) bool {
	pv1copy := pv1.DeepCopy()
	pv2copy := pv2.DeepCopy()
	pv1copy.Spec.ClaimRef = nil
	pv2copy.Spec.ClaimRef = nil
	return reflect.DeepEqual(pv1copy.Spec, pv2copy.Spec) &&
		reflect.DeepEqual(pv1copy.Labels, pv2copy.Labels)
}

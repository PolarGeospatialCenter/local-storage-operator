package common

import (
	"context"
	"fmt"

	"github.com/PolarGeospatialCenter/local-storage-operator/pkg/apis/localstorage/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type StorageObject interface {
	v1alpha1.PreparableStorageObject
	SetPreparePhase(v1alpha1.StoragePreparePhase)
	GetPreparePhase() v1alpha1.StoragePreparePhase
	GetEnabled() bool
	runtime.Object
}

type StorageObjectReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func NewStorageObjectReconciler(c client.Client, s *runtime.Scheme) *StorageObjectReconciler {
	return &StorageObjectReconciler{client: c, scheme: s}
}

func (r *StorageObjectReconciler) Reconcile(obj StorageObject) (reconcile.Result, error) {

	if !obj.GetEnabled() {
		return reconcile.Result{}, nil
	}

	if obj.GetPreparePhase() == v1alpha1.StoragePreparePhasePrepared {
		return reconcile.Result{}, r.createManagedPv(obj)
	}
	return reconcile.Result{}, r.prepareStorage(obj)
}

func (r *StorageObjectReconciler) createManagedPv(obj StorageObject) error {
	pv := obj.AsPv()
	existingPv := pv.DeepCopy()
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: existingPv.GetName()}, existingPv)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Error getting existing PV: %v", err)
	}

	notExists := errors.IsNotFound(err)

	if notExists {
		if err = controllerutil.SetControllerReference(obj, pv, r.scheme); err != nil {
			return err
		}
		err = r.client.Create(context.TODO(), pv)
		if err != nil {
			return fmt.Errorf("Error creating new PV: %v", err)
		}
		return nil
	}

	isDeleting := existingPv.DeletionTimestamp != nil
	released := existingPv.Status.Phase == "Released"

	if !notExists && !isDeleting && (released || !checkPvEqual(pv, existingPv)) {
		err = r.client.Delete(context.TODO(), existingPv)
		if err != nil {
			return fmt.Errorf("Error deleting existing PV: %v", err)
		}
	}

	return nil
}

func (r *StorageObjectReconciler) lookupPrepareJobTemplate(obj StorageObject) (*v1alpha1.StoragePrepareJobTemplate, error) {
	storagePrepareJobList := &v1alpha1.StoragePrepareJobTemplateList{}
	storagePrepareJobList.APIVersion = v1alpha1.SchemeGroupVersion.String()
	storagePrepareJobList.Kind = "StoragePrepareJobTemplate"

	err := r.client.List(context.TODO(), &client.ListOptions{}, storagePrepareJobList)

	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("error getting list of storagePrepareJobTemplates: %v", err)
	}

	return storagePrepareJobList.FindMatchingJob(obj)
}

func (r *StorageObjectReconciler) prepareStorage(obj StorageObject) error {
	switch obj.GetPreparePhase() {
	case v1alpha1.StoragePreparePhasePrepared:
		return nil
	case v1alpha1.StoragePreparePhaseDiscovered:
		prepJobTemplate, err := r.lookupPrepareJobTemplate(obj)
		if err != nil {
			return err
		}

		if prepJobTemplate == nil {
			return nil
		}

		prepJob := v1alpha1.NewStoragePrepareJob(prepJobTemplate, obj, prepJobTemplate.Namespace)

		if err = controllerutil.SetControllerReference(obj, prepJob, r.scheme); err != nil {
			return err
		}

		err = r.client.Create(context.TODO(), prepJob)
		if err != nil {
			return fmt.Errorf("Error creating storage prepare job: %v", err)
		}

		obj.SetPreparePhase(v1alpha1.StoragePreparePhasePreparing)
		err = r.client.Update(context.TODO(), obj)
		if err != nil {
			return err
		}

	case v1alpha1.StoragePreparePhasePreparing:
		prepJobTemplate, err := r.lookupPrepareJobTemplate(obj)

		if err != nil {
			return err
		}

		if prepJobTemplate == nil {
			return fmt.Errorf("no suitable StoragePrepareJobTemplate found.  This shouldn't happen, did you delete or update one?")
		}

		prepJob := v1alpha1.NewStoragePrepareJob(prepJobTemplate, obj, prepJobTemplate.Namespace)
		err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: prepJobTemplate.Namespace, Name: prepJob.Name}, prepJob)
		if err != nil {
			return err
		}

		switch prepJob.Status.Phase {
		case v1alpha1.StoragePrepareJobPhaseRunning:
			return nil
		case v1alpha1.StoragePrepareJobPhaseFailed:
			obj.SetPreparePhase(v1alpha1.StoragePreparePhaseFailed)
			return r.client.Update(context.TODO(), obj)
		case v1alpha1.StoragePrepareJobPhaseSucceeded:
			err = r.client.Delete(context.TODO(), prepJob, client.PropagationPolicy(metav1.DeletePropagationForeground))
			if err != nil {
				obj.SetPreparePhase(v1alpha1.StoragePreparePhaseFailed)
				r.client.Update(context.TODO(), obj)
				return fmt.Errorf("error deleting StoragePrepareJob %s: %v", prepJob.GetName(), err)
			}
			obj.SetPreparePhase(v1alpha1.StoragePreparePhasePrepared)
			return r.client.Update(context.TODO(), obj)
		}
	case v1alpha1.StoragePreparePhaseFailed:
		return nil
	}
	return nil
}

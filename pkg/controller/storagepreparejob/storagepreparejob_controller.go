package storagepreparejob

import (
	"context"
	"fmt"

	localstoragev1alpha1 "github.com/PolarGeospatialCenter/local-storage-operator/pkg/apis/localstorage/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_storagepreparejob")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new StoragePrepareJob Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileStoragePrepareJob{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("storagepreparejob-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource StoragePrepareJob
	err = c.Watch(&source.Kind{Type: &localstoragev1alpha1.StoragePrepareJob{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner StoragePrepareJob
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &localstoragev1alpha1.StoragePrepareJob{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileStoragePrepareJob{}

// ReconcileStoragePrepareJob reconciles a StoragePrepareJob object
type ReconcileStoragePrepareJob struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a StoragePrepareJob object and makes changes based on the state read
// and what is in the StoragePrepareJob.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileStoragePrepareJob) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling StoragePrepareJob")

	// Fetch the StoragePrepareJob instance
	instance := &localstoragev1alpha1.StoragePrepareJob{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	switch instance.Status.Phase {
	case localstoragev1alpha1.StoragePrepareJobPhasePending:

		for _, object := range instance.GetObjects() {
			if err = controllerutil.SetControllerReference(instance, object, r.scheme); err != nil {
				return reconcile.Result{}, err
			}
			err := r.client.Create(context.TODO(), object)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error creating %s while preparing %s: %v", object.GetObjectKind(), instance.GetName(), err)
			}
		}

		instance.Status.Phase = localstoragev1alpha1.StoragePrepareJobPhaseRunning
		return reconcile.Result{}, r.client.Update(context.TODO(), instance)

	case localstoragev1alpha1.StoragePrepareJobPhaseRunning:
		job := &instance.Spec.Job
		err := r.client.Get(context.TODO(), request.NamespacedName, job)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("Error getting job for %s even though status is running: %v", instance.Name, err)
		}

		if job.Status.Active > 0 {
			return reconcile.Result{}, nil
		}

		if job.Status.Failed > 0 {
			instance.Status.Phase = localstoragev1alpha1.StoragePrepareJobPhaseFailed
			return reconcile.Result{}, r.client.Update(context.TODO(), instance)
		}

		if job.Status.Succeeded > 0 {
			instance.Status.Phase = localstoragev1alpha1.StoragePrepareJobPhaseSucceeded
			return reconcile.Result{}, r.client.Update(context.TODO(), instance)
		}

	default:
		return reconcile.Result{}, nil

	}

	return reconcile.Result{}, nil
}

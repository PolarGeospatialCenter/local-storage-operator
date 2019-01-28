package filesystem

import (
	"context"

	localstoragev1alpha1 "github.com/PolarGeospatialCenter/local-storage-operator/pkg/apis/localstorage/v1alpha1"
	"github.com/PolarGeospatialCenter/local-storage-operator/pkg/controller/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_filesystem")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Filesystem Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileFilesystem{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("filesystem-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Filesystem
	err = c.Watch(&source.Kind{Type: &localstoragev1alpha1.Filesystem{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource StoragePrepareJob and requeue the owner Filesystem
	err = c.Watch(&source.Kind{Type: &localstoragev1alpha1.StoragePrepareJob{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &localstoragev1alpha1.Filesystem{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource PersistentVolume and requeue the owner Filesystem
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolume{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &localstoragev1alpha1.Disk{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileFilesystem{}

// ReconcileFilesystem reconciles a Filesystem object
type ReconcileFilesystem struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Filesystem object and makes changes based on the state read
// and what is in the Filesystem.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileFilesystem) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Filesystem")

	// Fetch the Filesystem instance
	instance := &localstoragev1alpha1.Filesystem{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name}, instance)
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

	return common.NewStorageObjectReconciler(r.client, r.scheme).Reconcile(instance)
}

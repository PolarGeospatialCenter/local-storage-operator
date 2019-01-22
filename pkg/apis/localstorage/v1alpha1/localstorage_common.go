package v1alpha1

type StoragePreparePhase string

const (
	StoragePreparePhaseDiscovered StoragePreparePhase = "discovered"
	StoragePreparePhasePreparing  StoragePreparePhase = "preparing"
	StoragePreparePhaseCleanup    StoragePreparePhase = "cleanup prep job"
	StoragePreparePhasePrepared   StoragePreparePhase = "prepared"
	StoragePreparePhaseFailed     StoragePreparePhase = "failed"
)

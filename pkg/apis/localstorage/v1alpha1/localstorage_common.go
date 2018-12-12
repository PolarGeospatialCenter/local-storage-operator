package v1alpha1

type StoragePreparePhase string

const (
	StoragePreparePhaseDiscovered StoragePreparePhase = "discovered"
	StoragePreparePhasePreparing  StoragePreparePhase = "preparing"
	StoragePreparePhasePrepared   StoragePreparePhase = "prepared"
	StoragePreparePhaseFailed     StoragePreparePhase = "failed"
)

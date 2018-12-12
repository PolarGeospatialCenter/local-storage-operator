package controller

import (
	"github.com/PolarGeospatialCenter/local-storage-operator/pkg/controller/filesystem"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, filesystem.Add)
}

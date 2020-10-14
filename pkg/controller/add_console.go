package controller

import (
	"github.com/tmax-cloud/console-operator/pkg/controller/console"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, console.Add)
}

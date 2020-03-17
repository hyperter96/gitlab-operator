package controller

import (
	"gitlab.com/ochienged/gitlab-operator/pkg/controller/runner"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, runner.Add)
}

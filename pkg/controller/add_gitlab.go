package controller

import (
	"gitlab.com/ochienged/gitlab-operator/pkg/controller/gitlab"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, gitlab.Add)
}

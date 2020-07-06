package controller

import (
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/controller/gitlab"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, gitlab.Add)
}

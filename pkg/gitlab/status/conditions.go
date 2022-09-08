package status

import (
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

const (
	ConditionInitialized gitlab.ConditionType = "Initialized"
	ConditionUpgrading   gitlab.ConditionType = "Upgrading"
	ConditionAvailable   gitlab.ConditionType = "Available"
)

const (
	PhasePreparing = "Preparing"
	PhaseRunning   = "Running"
)

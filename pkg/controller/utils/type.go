package utils

const (
	// RunnerType represents resource of type Runner
	RunnerType = "runner"

	// GitlabType represents resource of type Gitlab
	GitlabType = "gitlab"
)

var (
	// ConfigMapDefaultMode for configmap projected volume
	ConfigMapDefaultMode int32 = 420

	// ProjectedVolumeDefaultMode for projected volume
	ProjectedVolumeDefaultMode int32 = 256

	// SecretDefaultMode for secret projected volume
	SecretDefaultMode int32 = 288
)

// PasswordOptions provides parameters to be
// used when generating passwords
type PasswordOptions struct {
	// Length defines desired password length
	Length int
	// EnableSpecialCharacters adds special characters
	// to generated passwords
	EnableSpecialChars bool
}

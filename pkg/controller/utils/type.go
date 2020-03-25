package utils

const (
	// RunnerType represents resource of type Runner
	RunnerType = "runner"

	// GitlabType represents resource of type Gitlab
	GitlabType = "gitlab"
)

// PasswordOptions provides paramaters to be
// used when generating passwords
type PasswordOptions struct {
	// Length defines desired password length
	Length int
	// EnableSpecialCharacters adds special characters
	// to generated passwords
	EnableSpecialChars bool
}

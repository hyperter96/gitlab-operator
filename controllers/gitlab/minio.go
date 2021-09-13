package gitlab

const (
	internalMinioEnabled = "internalMinioEnabled"
)

// MinioEnabled returns `true` if enabled, and `false` if not.
func MinioEnabled(adapter CustomResourceAdapter) bool {
	enabled, _ := GetBoolValue(adapter.Values(), internalMinioEnabled, true)

	return enabled
}

package gitlab

const (
	globalMinioEnabled   = "global.minio.enabled"
	internalMinioEnabled = "internalMinioEnabled"
)

// MinioEnabled returns `true` if enabled, and `false` if not.
func MinioEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(internalMinioEnabled)
}

package settings

const (
	// LocalUser .
	LocalUser = "1000"

	// Region .
	Region = "us-east-1"

	// RegistryBucket .
	RegistryBucket = "registry"

	// AppConfigConnectionSecretName .
	AppConfigConnectionSecretName = "storage-config"

	// RegistryConnectionSecretName .
	RegistryConnectionSecretName = "registry-storage"

	// TaskRunnerConnectionSecretName .
	TaskRunnerConnectionSecretName = "s3cmd-config" // nolint:gosec
)

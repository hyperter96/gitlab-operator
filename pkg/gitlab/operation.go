package gitlab

// Operation represents the operations that are required to reconcile
// the underlying GitLab resource, for example install a new instance or
// upgrading/downgrading an existing instance.
type Operation interface {
	// IsInstall indicates if this GitLab resource needs to be installed. This
	// occurs when the GitLab instance with the specified name and version does
	// not exist.
	//
	// This function relies on both the specification and status of the GitLab
	// resource to make the decision.
	IsInstall() bool

	// IsUpgrade indicates if this GitLab resource needs upgrade. This occurs
	// when the GitLab instance already exists but its version is lower than
	// the specified version in GitLab resource specification.
	//
	// This function relies on both the specification and status of the GitLab
	// resource to make the decision.
	IsUpgrade() bool

	// IsDowngrade indicates if this GitLab resource needs downgrade. This
	// occurs when the GitLab instance already exists but its version is higher
	// than the specified version in GitLab resource specification.
	//
	// This function relies on both the specification and status of the GitLab
	// resource to make the decision.
	IsDowngrade() bool

	// CurrentVersion returns the current version of the GitLab instance or
	// an empty string when this is a new instance.
	//
	// This function uses the status of the GitLab resource.
	CurrentVersion() string

	// DesiredVersion returns the expected version of GitLab instance that is
	// specified.
	DesiredVersion() string
}

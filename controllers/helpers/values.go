package helpers

import (
	"fmt"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
)

// GetBoolValue returns the value of the specified key as boolean.
// Returns error if the key is not boolean.
func GetBoolValue(values helm.Values, key string) (bool, error) {
	x, err := values.GetValue(key)
	if err != nil {
		return false, err
	}

	if x == nil {
		return false, nil
	}

	b, ok := x.(bool)
	if !ok {
		return false, fmt.Errorf("Key %s is not a boolean value", key)
	}

	return b, nil
}

// GetStringValue returns the value of the specified key as string.
// Returns error if the key is not string.
func GetStringValue(values helm.Values, key string) (string, error) {
	x, err := values.GetValue(key)
	if err != nil {
		return "", err
	}

	if x == nil {
		return "", nil
	}

	s, ok := x.(string)
	if !ok {
		return "", fmt.Errorf("Key %s is not a string value", key)
	}

	return s, nil
}

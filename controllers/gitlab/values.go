package gitlab

import (
	"fmt"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

// GetBoolValue returns the value of the specified key as boolean.
// Returns error if the key is not boolean.
func GetBoolValue(values helm.Values, key string, defaults ...bool) (bool, error) {
	retVal := false
	if len(defaults) > 0 {
		retVal = defaults[0]
	}

	x, err := values.GetValue(key)
	if err != nil {
		return retVal, err
	}

	if x == nil {
		return retVal, nil
	}

	b, ok := x.(bool)
	if !ok {
		return retVal, fmt.Errorf("key %s is not a boolean value", key)
	}

	return b, nil
}

// GetStringValue returns the value of the specified key as string.
// Returns error if the key is not string.
func GetStringValue(values helm.Values, key string, defaults ...string) (string, error) {
	retVal := ""
	if len(defaults) > 0 {
		retVal = defaults[0]
	}

	x, err := values.GetValue(key)
	if err != nil {
		return retVal, err
	}

	if x == nil {
		return retVal, nil
	}

	s, ok := x.(string)
	if !ok {
		return retVal, fmt.Errorf("key %s is not a string value", key)
	}

	return s, nil
}

package helm

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/strvals"
	"sigs.k8s.io/yaml"
)

// Values stores a chart values.
type Values struct {
	container map[string]interface{}
}

// EmptyValues returns an empty value store.
func EmptyValues() *Values {
	return &Values{
		container: map[string]interface{}{},
	}
}

// AsMap returns values in a map.
func (v *Values) AsMap() map[string]interface{} {
	return v.container
}

// AddValue merges the specified value using the key.
func (v *Values) AddValue(key, value string) error {
	if err := strvals.ParseInto(fmt.Sprintf("%s=%s", key, value), v.container); err != nil {
		return errors.Wrapf(err, "failed to parse value assignment: %s=%s", key, value)
	}
	return nil
}

// AddStringValue merges the specified string value using the key.
func (v *Values) AddStringValue(key, value string) error {
	if err := strvals.ParseIntoString(fmt.Sprintf("%s=%s", key, value), v.container); err != nil {
		return errors.Wrapf(err, "failed to parse string value assignment: %s=%s", key, value)
	}
	return nil
}

// AddFileValue merges the content of the specified file using the key.
func (v *Values) AddFileValue(key, filePath string) error {
	reader := func(r []rune) (interface{}, error) {
		fileContent, err := ioutil.ReadFile(string(r))
		return string(fileContent), err
	}
	if err := strvals.ParseIntoFile(fmt.Sprintf("%s=%s", key, filePath), v.container, reader); err != nil {
		return errors.Wrapf(err, "failed to parse file value assignment: %s=%s", key, filePath)
	}
	return nil
}

// AddFromFile reads the specified file in YAML format and merges its content.
func (v *Values) AddFromFile(filePath string) error {
	newValues := map[string]interface{}{}

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(fileContent, &newValues); err != nil {
		return errors.Wrapf(err, "failed to parse value file: %s", filePath)
	}

	v.container = mergeMaps(v.container, newValues)

	return nil
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

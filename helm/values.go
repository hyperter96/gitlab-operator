package helm

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/strvals"
	"sigs.k8s.io/yaml"
)

// Values provides an interface to manipulate and store the values that are needed to render a Helm
// template.
type Values interface {
	// AsMap returns values in a map.
	AsMap() map[string]interface{}

	// AddValue merges the value using the key. The key conforms to Helm format.
	AddValue(key, value string) error

	// AddStringValue merges the string value using the key. The key conforms to Helm format.
	AddStringValue(key, value string) error

	// AddFileValue merges the content of the file using the key. The key conforms to Helm format.
	AddFileValue(key, filePath string) error

	// AddFromFile reads the specified file in YAML format and merges its content.
	AddFromFile(filePath string) error
}

// EmptyValues returns an empty value store.
func EmptyValues() Values {
	return &plainValues{
		container: map[string]interface{}{},
	}
}

type plainValues struct {
	container map[string]interface{}
}

func (v *plainValues) AsMap() map[string]interface{} {
	return v.container
}

func (v *plainValues) AddValue(key, value string) error {
	if err := strvals.ParseInto(fmt.Sprintf("%s=%s", key, value), v.container); err != nil {
		return errors.Wrapf(err, "failed to parse value assignment: %s=%s", key, value)
	}
	return nil
}

func (v *plainValues) AddStringValue(key, value string) error {
	if err := strvals.ParseIntoString(fmt.Sprintf("%s=%s", key, value), v.container); err != nil {
		return errors.Wrapf(err, "failed to parse string value assignment: %s=%s", key, value)
	}
	return nil
}

func (v *plainValues) AddFileValue(key, filePath string) error {
	reader := func(r []rune) (interface{}, error) {
		fileContent, err := ioutil.ReadFile(string(r))
		return string(fileContent), err
	}
	if err := strvals.ParseIntoFile(fmt.Sprintf("%s=%s", key, filePath), v.container, reader); err != nil {
		return errors.Wrapf(err, "failed to parse file value assignment: %s=%s", key, filePath)
	}
	return nil
}

func (v *plainValues) AddFromFile(filePath string) error {
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

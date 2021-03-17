package helm

import (
	"fmt"
	"io/ioutil"
	"strings"

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

	// GetValue returns the value of the specified key or an error when the key not present. The key
	// is limited to dot-separated field names and array indexing is not supported. It is the
	// responsibility of the caller to determine the type of the value.
	GetValue(key string) (interface{}, error)

	// SetValue sets the value of the specified key . Retruns error when the key is not writable. The
	// key is limited to dot-separated field names and array indexing is not supported. It is the
	// responsibility of the caller to pass the choose the right value type.
	SetValue(key string, value interface{}) error
}

// EmptyValues returns an empty value store.
func EmptyValues() Values {
	return &plainValues{
		container: map[string]interface{}{},
	}
}

// FromMap returns a value store that is populated with the provided map.
func FromMap(container map[string]interface{}) Values {
	if container == nil {
		return EmptyValues()
	}

	return &plainValues{
		container: container,
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

func (v *plainValues) GetValue(key string) (interface{}, error) {
	cursor := v.container
	elements := []string{}
	if key != "" {
		elements = strings.Split(key, ".")
	}

	var target interface{} = cursor
	for idx, elm := range elements {
		target = cursor[elm]

		if target == nil {
			if idx < len(elements)-1 {
				return nil, errors.Errorf("Missing element at %s for %s key", elm, key)
			}
		}

		if targetAsMap, ok := target.(map[string]interface{}); ok {
			cursor = targetAsMap
		} else {
			if idx < len(elements)-1 {
				return nil, errors.Errorf("Leaf element at %s for %s key", elm, key)
			}
		}
	}

	return target, nil
}

func (v *plainValues) SetValue(key string, value interface{}) error {
	if key == "" {
		return errors.Errorf("Can not set the root element")
	}

	cursor := v.container
	elements := strings.Split(key, ".")

	var target interface{} = cursor
	for idx, elm := range elements {
		target = cursor[elm]

		if target == nil {
			if idx < len(elements)-1 {
				target = map[string]interface{}{}
				cursor[elm] = target
			}
		}

		if targetAsMap, ok := target.(map[string]interface{}); ok {
			cursor = targetAsMap
		} else {
			if idx < len(elements)-1 {
				return errors.Errorf("Leaf element at %s for %s key", elm, key)
			}
		}
	}

	cursor[elements[len(elements)-1]] = value
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

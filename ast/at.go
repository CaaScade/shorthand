package ast

import (
	"fmt"
	"strings"

	"github.com/kr/pretty"
)

// MatchStringAt try to match a string.
func MatchStringAt(i interface{}, path string, expected string) bool {
	v, err := StringAt(i, path)
	return err == nil && v == expected
}

// At try to get the value at the given path.
func At(i interface{}, path string) (interface{}, error) {
	return at(i, strings.Split(path, "."))
}

func at(i interface{}, ks []string) (interface{}, error) {
	if len(ks) > 0 {
		k := ks[0]
		if m, ok := i.(map[string]interface{}); ok {
			if v, ok := m[k]; ok {
				return (at(v, ks[1:]))
			}

			return nil, fmt.Errorf(
				"no value for key %s", k)
		}

		return nil, pretty.Errorf(
			"expected a map (%# v)", i)
	}

	return i, nil
}

// SliceAt like At except it expects to find a slice.
func SliceAt(i interface{}, path string) ([]interface{}, error) {
	s, err := At(i, path)
	if err != nil {
		return nil, err
	}

	switch s := s.(type) {
	case []interface{}:
		return s, nil
	default:
		return nil, pretty.Errorf("expected slice (%# v)", s)
	}
}

// MapAt like At except it expects to find a map.
func MapAt(i interface{}, path string) (map[string]interface{}, error) {
	s, err := At(i, path)
	if err != nil {
		return nil, err
	}

	switch s := s.(type) {
	case map[string]interface{}:
		return s, nil
	default:
		return nil, pretty.Errorf("expected map (%# v)", s)
	}
}

// StringAt like At except it expects to find a string.
func StringAt(i interface{}, path string) (string, error) {
	s, err := At(i, path)
	if err != nil {
		return "", err
	}

	switch s := s.(type) {
	case string:
		return s, nil
	default:
		return "", pretty.Errorf("expected string (%# v)", s)
	}
}

// FloatAt like At except it expects to find a float64.
func FloatAt(i interface{}, path string) (float64, error) {
	s, err := At(i, path)
	if err != nil {
		return -12345, err
	}

	switch s := s.(type) {
	case float64:
		return s, nil
	default:
		return -12345, pretty.Errorf("expected float64 (%# v)", s)
	}
}

// CleanPath delete a path and remove any empty maps along the way.
func CleanPath(i interface{}, path string) error {
	ks := strings.Split(path, ".")
	if len(ks) == 0 {
		return nil
	}

	_, err := cleanPath(i, ks[0], ks[1:])

	return err
}

// Returns true if 'i' should be deleted from its parent.
func cleanPath(i interface{}, k string, ks []string) (bool, error) {
	switch i := i.(type) {
	case map[string]interface{}:
		if v, ok := i[k]; ok {
			var deleteV bool
			var err error
			if len(ks) > 0 {
				deleteV, err = cleanPath(v, ks[0], ks[1:])
			} else {
				deleteV = true
			}

			if err != nil {
				return false, err
			} else if deleteV {
				delete(i, k)
				return len(i) == 0, nil
			}
		}
		return false, nil
	default:
		return false, pretty.Errorf("not a map (%# v)", i)
	}
}

// InsertPath insert a value and create any missing maps along the way.
func InsertPath(i interface{}, path string, v interface{}) error {
	ks := strings.Split(path, ".")
	if len(ks) == 0 {
		return nil
	}

	return insertPath(i, ks[0], ks[1:], v)
}

func insertPath(i interface{}, k string, ks []string, v interface{}) error {
	switch i := i.(type) {
	case map[string]interface{}:
		if len(ks) == 0 {
			if _, ok := i[k]; ok {
				return pretty.Errorf("overwrite? at %v in (%# v)", k, i)
			}

			i[k] = v
			return nil
		}

		if ik, ok := i[k]; ok {
			return insertPath(ik, ks[0], ks[1:], v)
		}

		ik := make(map[string]interface{}, 1)
		i[k] = ik
		return insertPath(ik, ks[0], ks[1:], v)
	default:
		return pretty.Errorf("can't insert into (%# v)", i)
	}
}

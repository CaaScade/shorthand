package fs

import (
	"bytes"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/kr/pretty"
)

// ReadYamls read a yaml file of potentially multiple documents.
func ReadYamls(filename string) ([]interface{}, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	yamls := SplitYaml(contents)
	vs := make([]interface{}, len(yamls))
	for ix, y := range yamls {
		v := map[string]interface{}{}
		err = yaml.Unmarshal(y, &v)
		if err != nil {
			return nil, err
		}

		vs[ix] = v
	}

	return vs, nil
}

// SplitYaml split multi-document yaml file.
func SplitYaml(contents []byte) [][]byte {
	return bytes.Split(contents, []byte("\n---\n"))
}

// WriteYamls writes some objects (as yaml) to a buffer.
func WriteYamls(vs interface{}) (*bytes.Buffer, error) {
	switch vs := vs.(type) {
	case []interface{}:
		var b bytes.Buffer
		for ix, v := range vs {
			if ix > 0 {
				b.WriteString("\n---\n")
			}

			o, err := yaml.Marshal(v)
			if err != nil {
				return nil, err
			}

			b.Write(o)
		}

		return &b, nil
	default:
		return nil, pretty.Errorf("expected []interface{} (%# v)", vs)
	}
}

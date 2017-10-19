package isos

import (
	"fmt"
	"strings"

	log "github.com/koki/printline"
	"github.com/koki/shorthand/ast"
	"github.com/kr/pretty"
)

// ManifestIso shorthand format for an entire k8s manifest.
func ManifestIso() *ast.Iso {
	return &ast.Iso{Forward: shrink, Backward: grow}
}

func shrink(i interface{}) (interface{}, error) {
	var j interface{}
	var err error

	j, err = shrinkMounts(i)
	if err == nil {
		i = j
	}

	j, err = shrinkNodeSelector(i)
	if err == nil {
		i = j
	}

	j, err = shrinkMetadata(i)
	if err == nil {
		i = j
	}

	j, err = shrinkKind(i)
	if err == nil {
		i = j
	}

	return i, nil
}

func grow(j interface{}) (interface{}, error) {
	var i interface{}
	var err error
	i, err = growKind(j)
	if err == nil {
		j = i
	}

	i, err = growMetadata(j)
	if err == nil {
		j = i
	}

	i, err = growNodeSelector(j)
	if err == nil {
		j = i
	}

	i, err = growMounts(j)
	if err == nil {
		j = i
	}

	return j, nil
}

/*
From

kind: Pod
apiVersion: v1
metadata:
  name: pod1

To

Pod:
  name: pod1
*/

func shrinkKind(i interface{}) (interface{}, error) {
	switch i := i.(type) {
	case map[string]interface{}:
		kind, err := ast.StringAt(i, "kind")
		if err != nil {
			return nil, pretty.Errorf("no kind")
		}

		if _, ok := kindVersions[kind]; !ok {
			return nil, pretty.Errorf("please update kindVersions (shorthand source) for %s", kind)
		}

		delete(i, "kind")
		delete(i, "apiVersion")

		return map[string]interface{}{kind: i}, nil
	default:
		return nil, pretty.Errorf("root wasn't a map")
	}
}

var kindVersions = map[string]string{
	"Pod": "v1",
}

func growKind(i interface{}) (interface{}, error) {
	switch i := i.(type) {
	case map[string]interface{}:
		if len(i) != 1 {
			return nil, pretty.Errorf(
				"expected one top-level key (\"kind\")")
		}

		for kind, j := range i {
			if version, ok := kindVersions[kind]; ok {
				switch j := j.(type) {
				case map[string]interface{}:
					j["kind"] = kind
					j["apiVersion"] = version
					return j, nil
				default:
					return nil, pretty.Errorf(
						"child wasn't a map")
				}
			}

			return nil, pretty.Errorf(
				"unknown version for kind %v", kind)
		}

		return nil, pretty.Errorf("inconceivable")
	default:
		return nil, pretty.Errorf("root wasn't a map")
	}
}

func shrinkMetadata(i interface{}) (interface{}, error) {
	switch i := i.(type) {
	case map[string]interface{}:
		if metadata, ok := i["metadata"]; ok {
			switch metadata := metadata.(type) {
			case map[string]interface{}:
				if name, ok := metadata["name"]; ok {
					delete(metadata, "name")
					i["name"] = name

					if len(metadata) == 0 {
						delete(i, "metadata")
					}

					return i, nil
				}
			}

			return nil, pretty.Errorf("no metadata.name in (%# v)", i)
		}

		return nil, pretty.Errorf("no metadata in (%# v)", i)
	default:
		return nil, pretty.Errorf("root wasn't a map")
	}
}

func growMetadata(i interface{}) (interface{}, error) {
	switch i := i.(type) {
	case map[string]interface{}:
		if name, ok := i["name"]; ok {
			if metadata, ok := i["metadata"]; ok {
				switch metadata := metadata.(type) {
				case map[string]interface{}:
					metadata["name"] = name
					delete(i, "name")
					return i, nil
				default:
					return nil, pretty.Errorf("non-map metadata in (%# v)", i)
				}
			} else {
				i["metadata"] = map[string]interface{}{"name": name}
				delete(i, "name")
				return i, nil
			}
		}

		return nil, pretty.Errorf("no name in (%# v)", i)
	default:
		return nil, pretty.Errorf("root wasn't a map")
	}
}

/*
From

spec:
  affinity:
    nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            matchExpressions:
              key: type
              operator: eq
              values:
               - t2.micro

To

host_labels: type=t2.micro
*/

var nodeSelectorPath = "spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms.matchExpressions"

func shrinkNodeSelector(i interface{}) (interface{}, error) {
	matchx, err := ast.MapAt(i, nodeSelectorPath)
	if err != nil {
		return nil, err
	}

	var key, operator string
	var values []interface{}

	key, err = ast.StringAt(matchx, "key")
	if err != nil {
		return nil, err
	}

	operator, err = ast.StringAt(matchx, "operator")
	if err != nil {
		return nil, err
	}

	if operator != "eq" {
		return nil, pretty.Errorf("unrecognized operator (%# v)", i)
	}

	values, err = ast.SliceAt(matchx, "values")
	if len(values) == 0 {
		return nil, pretty.Errorf("no values (%# v)", i)
	}

	valueStrings := make([]string, 0, len(values))
	for _, v := range values {
		switch v := v.(type) {
		case string:
			valueStrings = append(valueStrings, v)
		default:
			return nil, pretty.Errorf("unrecognized value in (%# v)", i)
		}
	}

	valuesString := fmt.Sprintf("%v=%v", key, strings.Join(valueStrings, ","))

	err = ast.CleanPath(i, nodeSelectorPath)
	if err != nil {
		// Inconceivable!
		log.Fatal(err)
	}

	err = ast.InsertPath(i, "host_labels", valuesString)
	if err != nil {
		log.Fatal(err)
	}

	return i, nil
}

func growNodeSelector(i interface{}) (interface{}, error) {
	expr, err := ast.StringAt(i, "host_labels")
	segments := strings.Split(expr, "=")
	if len(segments) != 2 {
		return nil, pretty.Errorf("unrecognized host_labels %s", expr)
	}

	key := segments[0]
	values := strings.Split(segments[1], ",")

	matchx := map[string]interface{}{
		"key":      key,
		"operator": "eq",
		"values":   values,
	}

	err = ast.CleanPath(i, "host_labels")
	if err != nil {
		// Inconceivable!
		log.Fatal(err)
	}

	err = ast.InsertPath(i, nodeSelectorPath, matchx)
	if err != nil {
		log.Fatal(err)
	}

	return i, nil
}

/*
From

  containers:
   - name: master-container
     image: wlan0/postgresql-master:v0.1
     volumeMounts:
      - mountPath: /postgresql
        name: masterPV

To


  containers:
   - name: master-container
     image: wlan0/postgresql-master:v0.1
     mounts:
       - masterPV:/postgresql
*/

func shrinkMounts(i interface{}) (interface{}, error) {
	containers, err := ast.SliceAt(i, "spec.containers")
	if err != nil {
		return nil, err
	}

	transMount := func(i interface{}) (interface{}, error) {
		switch i := i.(type) {
		case map[string]interface{}:
			if len(i) == 2 {
				var mountPath, name string
				var err error
				mountPath, err = ast.StringAt(i, "mountPath")
				if err != nil {
					return nil, err
				}

				name, err = ast.StringAt(i, "name")
				if err != nil {
					return nil, err
				}

				return fmt.Sprintf("%s:%s", name, mountPath), nil
			}

			return nil, pretty.Errorf("expected 2 fields in (%# v)", i)
		default:
			return nil, pretty.Errorf("mount should be map (%# v)", i)
		}
	}

	transCont := func(i interface{}) (interface{}, error) {
		mounts, err := ast.SliceAt(i, "volumeMounts")
		if err != nil {
			return nil, err
		}

		var newMounts interface{}
		newMounts, err = ast.MultiplyTransform(transMount)(mounts)
		if err != nil {
			return nil, err
		}

		err = ast.CleanPath(i, "volumeMounts")
		if err != nil {
			return nil, err
		}

		err = ast.InsertPath(i, "mounts", newMounts)
		if err != nil {
			return nil, err
		}

		return i, nil
	}

	var newContainers interface{}
	newContainers, err = ast.MultiplyTransform(transCont)(containers)
	if err != nil {
		return nil, err
	}

	err = ast.CleanPath(i, "spec.containers")
	if err != nil {
		return nil, err
	}

	err = ast.InsertPath(i, "spec.containers", newContainers)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func growMounts(i interface{}) (interface{}, error) {
	containers, err := ast.SliceAt(i, "spec.containers")
	if err != nil {
		return nil, err
	}

	transMount := func(i interface{}) (interface{}, error) {
		switch i := i.(type) {
		case string:
			segments := strings.Split(i, ":")
			if len(segments) != 2 {
				return nil, pretty.Errorf("unrecognized mount (%s)", i)
			}

			name := segments[0]
			mountPath := segments[1]

			mount := map[string]interface{}{
				"name":      name,
				"mountPath": mountPath,
			}
			return mount, nil

		default:
			return nil, pretty.Errorf("unrecognized mount (%# v)", i)
		}
	}

	transCont := func(i interface{}) (interface{}, error) {
		mounts, err := ast.SliceAt(i, "mounts")
		if err != nil {
			return nil, err
		}

		var newMounts interface{}
		newMounts, err = ast.MultiplyTransform(transMount)(mounts)
		if err != nil {
			return nil, err
		}

		err = ast.CleanPath(i, "mounts")
		if err != nil {
			return nil, err
		}

		err = ast.InsertPath(i, "volumeMounts", newMounts)
		if err != nil {
			return nil, err
		}

		return i, nil
	}

	var newContainers interface{}
	newContainers, err = ast.MultiplyTransform(transCont)(containers)
	if err != nil {
		return nil, err
	}

	err = ast.CleanPath(i, "spec.containers")
	if err != nil {
		return nil, err
	}

	err = ast.InsertPath(i, "spec.containers", newContainers)
	if err != nil {
		return nil, err
	}

	return i, nil
}

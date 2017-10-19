package ast

import (
	"fmt"
)

// Transform .
type Transform func(interface{}) (interface{}, error)

// Iso doot.
type Iso struct {
	Forward  Transform
	Backward Transform
}

// IdentityIso identity isomorphism.
var IdentityIso = &Iso{IdentityTransform, IdentityTransform}

// IdentityTransform identity transform.
func IdentityTransform(i interface{}) (interface{}, error) {
	return i, nil
}

// FlipIso flip "forward" and "backward" directions.
func FlipIso(iso *Iso) *Iso {
	return &Iso{iso.Forward, iso.Backward}
}

// MultiplyTransform multiply a transform to apply it to many objects.
func MultiplyTransform(transform Transform) Transform {
	return func(objs interface{}) (interface{}, error) {
		switch objs := objs.(type) {
		case []interface{}:
			for ix, obj := range objs {
				newObj, err := transform(obj)
				if err == nil {
					objs[ix] = newObj
				}
			}

			return objs, nil
		default:
			return nil, fmt.Errorf(
				"multiply root wasn't a slice")
		}
	}
}

/*
func SequenceTransforms(transforms ...Transform) Transform {
	return func(i interface{}) (interface{}, error) {
		//var j map[string]interface{}
		//var err error
		for _, transform := range transforms {
			j, err := transform(i)
			if err == nil {

			}
		}
	}
}
*/

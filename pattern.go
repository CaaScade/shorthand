package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"log"
)

// Using "U" to indicate "unboxed" types.

// AbsentU indicates that a named Field was not present.
type AbsentU int

// Absent is the canonical value of type AbsentU.
const (
	Absent AbsentU = iota
)

// Field is a named hole.
type Field struct {
	name string

	value *Pattern
}

// Pattern is either a single hole or a list of named holes.
type Pattern struct {
	error error

	/* Field is optional. */
	fields []*Field
	/* Field is present IFF !fields */
	value interface{}
}

// P is shorthand for Pattern
type P map[string]interface{}

// Match doot.
func (p *Pattern) Match(i interface{}) {
	if p.fields == nil {
		p.value = i
	} else if m, ok := i.(map[string]interface{}); ok {
		matchFields(p.fields, m)
	} else {
		p.error = fmt.Errorf("expected map %v", i)
	}
}

// MatchFields doot.
func matchFields(fs []*Field, m map[string]interface{}) {
	for _, f := range fs {
		f.Match(m)
	}
}

// Match doot.
func (f *Field) Match(m map[string]interface{}) {
	if v, ok := m[f.name]; ok {
		f.value.Match(v)
	} else {
		f.value.Match(Absent)
	}
}

// MkPattern doot.
func MkPattern(fs ...*Field) *Pattern {
	// If no fields, then match the whole structure.

	// Fields should not have overlapping names.
	// TODO: Verify that ^
	return &Pattern{nil, fs, nil}
}

// ValPattern doot.
func ValPattern(i interface{}) *Pattern {
	return &Pattern{nil, nil, i}
}

// MkField doot.
func MkField(name string, fs ...*Field) *Field {
	// If no fields, then match the whole structure.
	return &Field{name, MkPattern(fs...)}
}

// ValField doot.
func ValField(name string, i interface{}) *Field {
	return &Field{name, ValPattern(i)}
}

// MkP doot.
func MkP(p P) *Pattern {
	fs := make([]*Field, 0, len(p))
	for name, v := range p {
		var f *Field
		switch v.(type) {
		case P:
			f = &Field{name, MkP(v.(P))}
		default:
			f = ValField(name, v)
		}

		fs = append(fs, f)
	}

	return MkPattern(fs...)
}

// removeFields - remove the matched structure from the nested map.
func removeFields(fs []*Field, m map[string]interface{}) interface{} {
	for _, f := range fs {
		removeField(f, m)
	}

	for k, v := range m {
		switch v.(type) {
		case AbsentU:
			delete(m, k)
		}
	}

	if len(m) > 0 {
		return m
	}

	// We deleted all the fields from this structure.
	return Absent
}

// removeField - remove a field from this map.
func removeField(f *Field, m map[string]interface{}) {
	if v, ok := m[f.name]; ok {
		m[f.name] = removePattern(f.value, v)
	} else {
		log.Fatal(spew.Sdump("field doesn't match map", f, m))
	}
}

// removePattern removes the captured fields from an object and returns the result.
func removePattern(p *Pattern, i interface{}) interface{} {
	if p.fields == nil {
		// We're removing the whole thing.
		return Absent
	} else if m, ok := i.(map[string]interface{}); ok {
		return removeFields(p.fields, m)
	} else {
		log.Fatal(spew.Sdump("pattern doesn't match object", p, i))
		return Absent
	}
}

// Erase a Pattern from an object and return the result. (Modifies the object.)
func (p *Pattern) Erase(i interface{}) interface{} {
	return removePattern(p, i)
}

// insertPattern merges the captured fields into an object and returns the result.
func insertPattern(p *Pattern, i interface{}) interface{} {
	if p.fields == nil {
		switch i.(type) {
		case AbsentU:
			return p.value
		default:
			log.Fatal(spew.Sdump("attempted to overwrite", p, i))
			return i
		}
	} else {
		return insertFields(p.fields, i)
	}
}

// insertFields merges the captured fields into an object and returns the result.
func insertFields(fs []*Field, i interface{}) interface{} {
	var m map[string]interface{}
	switch i.(type) {
	case AbsentU:
		m = nil
	case map[string]interface{}:
		m = i.(map[string]interface{})
	default:
		log.Fatal(spew.Sdump("expected map or Absent", i))
	}

	for _, f := range fs {
		// What's the old value?
		var v0 interface{}
		if m != nil {
			if mv, ok := m[f.name]; ok {
				v0 = mv
			} else {
				v0 = Absent
			}
		} else {
			v0 = Absent
		}

		// Get the new value and insert it (if present).
		v := insertPattern(f.value, v0)
		switch v.(type) {
		case AbsentU:
			// Do nothing.
		default:
			if m == nil {
				m = map[string]interface{}{f.name: v}
			} else {
				m[f.name] = v
			}
		}
	}

	return m
}

// Write a pattern to an object. (Modifies the object.)
func (p *Pattern) Write(i interface{}) interface{} {
	return insertPattern(p, i)
}

// Extract captured values from a Pattern.
func (p *Pattern) Extract() interface{} {
	switch {
	case p.error != nil:
		return p.error
	case p.fields == nil:
		return p.value
	default:
		return extractFields(p.fields)
	}
}

// Extract captured values from a Field.
func (f *Field) Extract() map[string]interface{} {
	return map[string]interface{}{f.name: f.value.Extract()}
}

// extractFields doot.
func extractFields(fs []*Field) map[string]interface{} {
	m := map[string]interface{}{}
	for _, f := range fs {
		m[f.name] = f.value.Extract()
	}

	return m
}

// At doot.
func At(i interface{}, ks ...string) (interface{}, error) {
	if len(ks) > 0 {
		k := ks[0]
		if m, ok := i.(map[string]interface{}); ok {
			if v, ok := m[k]; ok {
				return (At(v, ks[1:]...))
			}

			return nil, fmt.Errorf(
				"no value for key %s", k)
		}

		return nil, fmt.Errorf(
			"expected a map %v", i)
	}

	return i, nil
}

/*
// Set doot.
func Set(m map[string]interface{}, v interface{}, ks ...string) error {
	l := len(ks)
	if l > 0 {
		k := ks[0]

		if l == 1 {
			m[k] = v
			return nil
		}

		mk := m[k]

		if mm, ok := mk.(map[string]interface{}); ok {
			return Set(mm, v, ks[1:]...)
		}

		return fmt.Errorf("expected a map %v", mk)
	}

	return fmt.Errorf("can't set with zero keys")
}
*/

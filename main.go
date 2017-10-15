package main

import (
	"bytes"
	"encoding/json"
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
	error string

	/* Field is optional. */
	fields []*Field
	/* Field is present IFF !fields */
	value interface{}
}

func main() {
	v := map[string]interface{}{}
	blob := `{"doot":{},"boop":{"wat":"yes"}}`
	if err := json.Unmarshal([]byte(blob), &v); err != nil {
		log.Fatal(err)
	}

	var out bytes.Buffer
	if err := json.Indent(&out, []byte(blob), "", "  "); err != nil {
		log.Fatal(err)
	}

	blob1 := out.String()
	fmt.Println(blob1)
	p := MkPattern(MkField("boop", MkField("wat")))
	p.Match(v)
	spew.Dump(p)

	v1 := RemovePattern(p, v)
	blob2, err2 := json.MarshalIndent(v1, "", "  ")
	if err2 != nil {
		log.Fatal(err2)
	}
	fmt.Println(string(blob2))
	//spew.Dump(v)

	q := MkPattern(MkField("b", ValField("p", 1000)))
	v2 := InsertPattern(q, v1)
	blob3, err3 := json.MarshalIndent(v2, "", "  ")
	if err3 != nil {
		log.Fatal(err3)
	}
	fmt.Println(string(blob3))
	//spew.Dump(v2)
}

// Match doot.
func (p *Pattern) Match(i interface{}) {
	if p.fields == nil {
		p.value = i
	} else if m, ok := i.(map[string]interface{}); ok {
		MatchFields(p.fields, m)
	} else {
		p.error = fmt.Sprint("expected map", i)
	}
}

// MatchFields doot.
func MatchFields(fs []*Field, m map[string]interface{}) {
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
	return &Pattern{"", fs, nil}
}

// ValPattern doot.
func ValPattern(i interface{}) *Pattern {
	return &Pattern{"", nil, i}
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

// RemoveFields - remove the matched structure from the nested map.
func RemoveFields(fs []*Field, m map[string]interface{}) interface{} {
	for _, f := range fs {
		RemoveField(f, m)
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

// RemoveField - remove a field from this map.
func RemoveField(f *Field, m map[string]interface{}) {
	if v, ok := m[f.name]; ok {
		m[f.name] = RemovePattern(f.value, v)
	} else {
		log.Fatal("field doesn't match map", f, m)
	}
}

// RemovePattern removes the captured fields from an object and returns the result.
func RemovePattern(p *Pattern, i interface{}) interface{} {
	if p.fields == nil {
		// We're removing the whole thing.
		return Absent
	} else if m, ok := i.(map[string]interface{}); ok {
		return RemoveFields(p.fields, m)
	} else {
		log.Fatal("pattern doesn't match object", p, i)
		return Absent
	}
}

// InsertPattern merges the captured fields into an object and returns the result.
func InsertPattern(p *Pattern, i interface{}) interface{} {
	if p.fields == nil {
		switch i.(type) {
		case AbsentU:
			return p.value
		default:
			log.Fatal("attempted to overwrite", p, i)
			return i
		}
	} else {
		return InsertFields(p.fields, i)
	}
}

// InsertFields merges the captured fields into an object and returns the result.
func InsertFields(fs []*Field, i interface{}) interface{} {
	var m map[string]interface{}
	switch i.(type) {
	case AbsentU:
		m = nil
	case map[string]interface{}:
		m = i.(map[string]interface{})
	default:
		log.Fatal("expected map or Absent", i)
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
		v := InsertPattern(f.value, v0)
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

// Extract doot.
func Extract(p *Pattern) interface{} {
	return nil
}

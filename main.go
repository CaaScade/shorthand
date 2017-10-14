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
	blob := `{"doot":1,"boop":{"wat":"yes"}}`
	if err := json.Unmarshal([]byte(blob), &v); err != nil {
		log.Fatal(err)
	}

	var out bytes.Buffer
	if err := json.Indent(&out, []byte(blob), "", "  "); err != nil {
		log.Fatal(err)
	}

	blob1 := out.String()
	fmt.Println(blob1)
	/*
		p := FieldsPattern(
			SimpleField("doot"),
			PatternField("boop", FieldsPattern(SimpleField("wat"))))
	*/
	p := FieldsPattern(PatternField("boop", FieldsPattern(SimpleField("wat"))))
	p.Match(v)
	spew.Dump(p)

	v1 := RemovePattern(p, v)
	blob2, err2 := json.MarshalIndent(v1, "", "  ")
	if err2 != nil {
		log.Fatal(err2)
	}
	fmt.Println(string(blob2))
	spew.Dump(v)
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

// WholePattern doot.
func WholePattern() *Pattern {
	return &Pattern{"", nil, nil}
}

// FieldsPattern doot.
func FieldsPattern(fs ...*Field) *Pattern {
	// Fields should not have overlapping names.
	// TODO: Verify that ^
	return &Pattern{"", fs, nil}
}

// SimpleField doot.
func SimpleField(name string) *Field {
	return &Field{name, WholePattern()}
}

// PatternField doot.
func PatternField(name string, p *Pattern) *Field {
	return &Field{name, p}
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

// RemovePattern doot.
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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	//"os"
	//"reflect"
	"github.com/davecgh/go-spew/spew"
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

	for key, value := range v {
		fmt.Println("Key:", key, "Value:", value)
	}

	var p = FieldsPattern(
		SimpleField("doot"),
		PatternField("boop", FieldsPattern(SimpleField("wat"))),
		SimpleField("boop"))
	spew.Dump(p)
	p.Match(v)
	spew.Dump(p)
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
	v, ok := m[f.name]
	if ok {
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

// Fields doot.
func Fields(fs ...*Field) []*Field {
	return fs
}

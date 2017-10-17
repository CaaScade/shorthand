package main

import (
	"fmt"
	"github.com/kr/pretty"
	"log"
	"reflect"
)

// Using "U" to indicate "unboxed" types.

// AbsentU indicates that a named Field was not present.
type AbsentU int

// Absent is the canonical value of type AbsentU.
const (
	Absent AbsentU = iota
)

// WildcardU indicates that the pattern matches anything.
type WildcardU int

// Wild is the canonical value of type WildcardU.
// TODO: Typed wildcards.
const (
	AnyW WildcardU = iota
	StringW
	FloatW
)

// Field is a named hole.
type Field struct {
	name string

	value *Pattern
}

// Pattern is either a single hole or a list of named holes.
type Pattern struct {
	error   error
	capture interface{}

	/* Option 1. */
	fields    []*Field
	exclusive bool // !(Does it count as a match if other fields are present?)

	/* Option 2. Used if Option 1 is nil. */
	constant interface{}
}

// IsFields doot.
func (p *Pattern) IsFields() bool {
	return p.fields != nil
}

// P is shorthand for "Partial" Pattern
type P map[string]interface{}

// XP is shorthand for "Exclusive" Pattern
type XP map[string]interface{}

// ConstPattern doot.
func ConstPattern(i interface{}) *Pattern {
	return &Pattern{nil, Absent, nil, false, i}
}

// FieldsPattern doot.
func FieldsPattern(exclusive bool, fs ...*Field) *Pattern {
	// If no fields, then why are we here?
	if len(fs) == 0 {
		log.Fatal("no fields for FieldsPattern")
	}

	// Fields should not have overlapping names.
	// TODO: Verify that ^
	return &Pattern{nil, Absent, fs, exclusive, AnyW}
}

// AnyPattern doot.
func AnyPattern() *Pattern {
	return ConstPattern(AnyW)
}

// StringPattern doot.
func StringPattern() *Pattern {
	return ConstPattern(StringW)
}

// FloatPattern doot.
func FloatPattern() *Pattern {
	return ConstPattern(FloatW)
}

// ConstField doot.
func ConstField(name string, i interface{}) *Field {
	return &Field{name, ConstPattern(i)}
}

// FieldsField doot.
func FieldsField(name string, exclusive bool, fs ...*Field) *Field {
	// If no fields, then match the whole structure.
	return &Field{name, FieldsPattern(exclusive, fs...)}
}

// AnyField doot.
func AnyField(name string) *Field {
	return ConstField(name, AnyW)
}

// StringField doot.
func StringField(name string) *Field {
	return ConstField(name, StringW)
}

// FloatField doot.
func FloatField(name string) *Field {
	return ConstField(name, FloatW)
}

// MkP doot.
func MkP(p interface{}) *Pattern {
	switch p := p.(type) {
	case P:
		return MkFields(p, false)
	case XP:
		return MkFields(p, true)
	case int:
		return ConstPattern(float64(p))
	default:
		return ConstPattern(p)
	}
}

// MkFields doot.
func MkFields(m map[string]interface{}, exclusive bool) *Pattern {
	fs := make([]*Field, 0, len(m))
	for name, v := range m {
		f := &Field{name, MkP(v)}
		fs = append(fs, f)
	}

	return FieldsPattern(exclusive, fs...)
}

// Match a pattern to an object and capture its values.
func (p *Pattern) Match(i interface{}) {
	if p.IsFields() {
		p.matchFields(i)
	} else {
		p.capture = i
		switch c := p.constant.(type) {
		case WildcardU:
			p.checkWildcard(c, i)
		default:
			// This path also matches Absent.
			if !reflect.DeepEqual(p.constant, i) {
				p.error = fmt.Errorf(
					"constant and capture don't match")
			}
		}
	}
}

func (p *Pattern) checkWildcard(w WildcardU, i interface{}) {
	switch w {
	case AnyW:
		// No error
	case StringW:
		if _, ok := i.(string); !ok {
			p.error = pretty.Errorf("expected string (%# v)", i)
		}
	case FloatW:
		if _, ok := i.(float64); !ok {
			p.error = pretty.Errorf("expected float (%# v)", i)
		}
	}
}

func (p *Pattern) matchFields(i interface{}) {
	if m, ok := i.(map[string]interface{}); ok {
		for _, f := range p.fields {
			f.match(m)
		}

		if p.exclusive {
			ks := make(map[string]bool, len(p.fields))
			for _, f := range p.fields {
				ks[f.name] = true
			}

			for k := range m {
				if _, ok := ks[k]; !ok {
					p.error = fmt.Errorf("extraneous fields for exclusive pattern")
					p.capture = m
				}
			}
		}
	} else {
		p.error = fmt.Errorf("expected map")
		p.capture = i
	}
}

func (f *Field) match(m map[string]interface{}) {
	if v, ok := m[f.name]; ok {
		f.value.Match(v)
	} else {
		f.value.Match(Absent)
	}
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

// TODO: removal methods (Erase) should also capture errors?
//   i.e. if the pattern doesn't need to match perfectly

// removeField - remove a field from this map.
func removeField(f *Field, m map[string]interface{}) {
	if v, ok := m[f.name]; ok {
		m[f.name] = removePattern(f.value, v)
	} else {
		m[f.name] = removePattern(f.value, Absent)
	}
}

// removePattern removes the captured fields from an object and returns the result.
func removePattern(p *Pattern, i interface{}) interface{} {
	if !p.IsFields() {
		// We're removing the whole thing.
		return Absent
	}

	if m, ok := i.(map[string]interface{}); ok {
		return removeFields(p.fields, m)
	}

	log.Fatal(pretty.Sprint("pattern doesn't match object", p, i))
	return Absent
}

// Erase a Pattern from an object and return the result. (Modifies the object.)
func (p *Pattern) Erase(i interface{}) interface{} {
	return removePattern(p, i)
}

// insertPattern merges the captured fields into an object and returns the result.
func insertPattern(p *Pattern, i interface{}) (interface{}, error) {
	if p.IsFields() {
		return insertFields(p.fields, i)
	}

	switch p.constant.(type) {
	case WildcardU:
		ii, err := insertValue(p.capture, i)
		// We should be using constant patterns to write.
		if err != nil {
			err = fmt.Errorf("unexpected Wildcard (%v), also (%v)",
				p, err)
		} else {
			err = fmt.Errorf("unexpected Wildcard (%v)", p)
		}

		return ii, err
	default:
		return insertValue(p.constant, i)
	}
}

func insertValue(v interface{}, i interface{}) (interface{}, error) {
	switch i.(type) {
	case AbsentU:
		return v, nil
	default:
		return v, pretty.Errorf("wrote (%# v) over (%# v)", v, i)
	}
}

// insertFields merges the captured fields into an object and returns the result.
func insertFields(fs []*Field, i interface{}) (interface{}, error) {
	var m map[string]interface{}
	switch i.(type) {
	case AbsentU:
		m = nil
	case map[string]interface{}:
		m = i.(map[string]interface{})
	default:
		return i, pretty.Errorf("expected map or Absent (%# v)", i)
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
		v, err := insertPattern(f.value, v0)
		if err != nil {
			return i, fmt.Errorf("%s.%v", f.name, err)
		}

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

	return m, nil
}

// Write a pattern to an object. (Modifies the object.)
func (p *Pattern) Write(i interface{}) (interface{}, error) {
	return insertPattern(p, i)
}

// HasErrors returns true if the pattern didn't Match correctly.
func (p *Pattern) HasErrors() bool {
	if p.error != nil {
		return true
	}

	for _, f := range p.fields {
		if f.value.HasErrors() {
			return true
		}
	}

	return false
}

// ExtractString convenience method.
// TODO: With typed wildcards, this shouldn't fail in normal usage.
func (p *Pattern) ExtractString() string {
	x := p.Extract()
	if s, ok := x.(string); ok {
		return s
	}

	log.Fatal(pretty.Sprintf("not a string (%# v)", x))
	return ""
}

// ExtractFloat convenience method.
// TODO: With typed wildcards, this shouldn't fail in normal usage.
func (p *Pattern) ExtractFloat() float64 {
	x := p.Extract()
	if f, ok := x.(float64); ok {
		return f
	}

	log.Fatal("not a float", x)
	return -12345
}

// Extract captured values from a Pattern.
func (p *Pattern) Extract() interface{} {
	switch {
	case p.error != nil:
		return p.error
	case p.IsFields():
		return extractFields(p.fields)
	default:
		return p.capture
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

// At should always succeed. An error probably means programmer error.
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

		return nil, pretty.Errorf(
			"expected a map (%# v)", i)
	}

	return i, nil
}

// StringAt convenience method.
// In normal use, it could fail if the pattern matched something other than a string.
func StringAt(i interface{}, ks ...string) string {
	s, err := At(i, ks...)
	if err != nil {
		log.Fatal(err)
	}

	switch s := s.(type) {
	case string:
		return s
	default:
		log.Fatal("expected string", s)
		return ""
	}
}

// FloatAt convenience method.
// In normal use, it could fail if the pattern matched something other than a float64.
func FloatAt(i interface{}, ks ...string) float64 {
	s, err := At(i, ks...)
	if err != nil {
		log.Fatal(err)
	}

	switch s := s.(type) {
	case float64:
		return s
	default:
		log.Fatal("expected float64", s)
		return -12345
	}
}

// WildcardPath doot.
func (p *Pattern) WildcardPath() ([]string, error) {
	r, err := p.ReverseWildcardPath()
	if err != nil {
		return nil, err
	}

	l := len(r)

	ks := make([]string, l, l)
	for i, k := range r {
		ks[l-i-1] = k
	}

	return ks, nil
}

// ReverseWildcardPath doot.
func (p *Pattern) ReverseWildcardPath() ([]string, error) {
	if p.IsFields() {
		return ReverseWildcardPath(p.fields...)
	}

	switch p.constant.(type) {
	case WildcardU:
		return []string{}, nil
	default:
		return nil, fmt.Errorf("dead end")
	}
}

// ReverseWildcardPath doot.
func ReverseWildcardPath(fs ...*Field) ([]string, error) {
	for _, f := range fs {
		ks, err := f.value.ReverseWildcardPath()
		if err == nil {
			return append(ks, f.name), nil
		}
	}

	return nil, fmt.Errorf("dead ends")
}

// Clear captures from pattern (so it can be reused)
func (p *Pattern) Clear() {
	p.error = nil
	p.capture = Absent
	if p.IsFields() {
		for _, f := range p.fields {
			f.value.Clear()
		}
	}
}

// SetConst doot.
func (p *Pattern) SetConst(v interface{}, ks ...string) error {
	if len(ks) > 0 {
		if !p.IsFields() {
			return pretty.Errorf("no more keys at (%# v)", p)
		}

		return setConst(p.fields, ks[0], v, ks[1:])
	}

	if p.IsFields() {
		return pretty.Errorf(
			"dead end (%# v) with keys (%# v)", p, ks)
	}

	p.constant = v

	return nil
}

// setConst doot.
func setConst(fs []*Field, k string, v interface{}, ks []string) error {
	for _, f := range fs {
		if f.name == k {
			return f.value.SetConst(v, ks...)
		}
	}

	return pretty.Errorf("no value for (%# v):(%# v) in (%# v)", k, ks, fs)
}

// Clone doot.
func (p *Pattern) Clone() *Pattern {
	return &Pattern{p.error, p.capture, clone(p.fields), p.exclusive, p.constant}
}

func clone(fs []*Field) []*Field {
	if fs == nil {
		return nil
	}

	l := len(fs)
	ffs := make([]*Field, l, l)
	for i, f := range fs {
		ffs[i] = &Field{f.name, f.value.Clone()}
	}

	return ffs
}

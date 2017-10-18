package main

import (
//"github.com/kr/pretty"
//"log"
)

// Iso doot.
type Iso struct {
	forward  *Prism
	backward *Prism
}

// IdentityIso doot.
func IdentityIso() *Iso {
	return &Iso{IdentityPrism(), IdentityPrism()}
}

// ZoomIso doot.
func ZoomIso(telescope *Pattern, i *Iso) *Iso {
	return &Iso{ZoomPrism(telescope, i.forward), ZoomPrism(telescope, i.backward)}
}

// MultiplyIso doot.
func MultiplyIso(i *Iso) *Iso {
	return &Iso{MultiplyPrism(i.forward), MultiplyPrism(i.backward)}
}

// SequenceIsos doot.
func SequenceIsos(is ...*Iso) *Iso {
	l := len(is)
	fs := make([]*Prism, l, l)
	bs := make([]*Prism, l, l)
	for ix, i := range is {
		fs[ix] = i.forward
		bs[ix] = i.backward
	}

	return &Iso{SequencePrisms(fs...), SequencePrisms(bs...)}
}

// MkIso doot.
func MkIso(from *Pattern, to *Pattern, split func(*Pattern) (*Pattern, error), unsplit func(*Pattern) (*Pattern, error)) *Iso {
	return &Iso{&Prism{from, split}, &Prism{to, unsplit}}
}

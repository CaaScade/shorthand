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
	return &Iso{Identity(), Identity()}
}

// ZoomIso doot.
func ZoomIso(telescope *Pattern, i *Iso) *Iso {
	return &Iso{Zoom(telescope, i.forward), Zoom(telescope, i.backward)}
}

// MultiplyIso doot.
func MultiplyIso(i *Iso) *Iso {
	return &Iso{Multiply(i.forward), Multiply(i.backward)}
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

	return &Iso{Sequence(fs...), Sequence(bs...)}
}

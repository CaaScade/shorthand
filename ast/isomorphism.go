package ast

import ()

// Iso doot.
type Iso struct {
	Forward  *Prism
	Backward *Prism
}

// IdentityIso doot.
func IdentityIso() *Iso {
	return &Iso{IdentityPrism(), IdentityPrism()}
}

// ZoomIso doot.
func ZoomIso(telescope *Pattern, i *Iso) *Iso {
	return &Iso{ZoomPrism(telescope, i.Forward), ZoomPrism(telescope, i.Backward)}
}

// MultiplyIso doot.
func MultiplyIso(i *Iso) *Iso {
	return &Iso{MultiplyPrism(i.Forward), MultiplyPrism(i.Backward)}
}

// SequenceIsos doot.
func SequenceIsos(is ...*Iso) *Iso {
	l := len(is)
	fs := make([]*Prism, l)
	bs := make([]*Prism, l)
	for ix, i := range is {
		fs[ix] = i.Forward

		// Compose reverse prisms in reverse order.
		bs[l-ix-1] = i.Backward
	}

	return &Iso{SequencePrisms(fs...), SequencePrisms(bs...)}
}

// MkIso doot.
func MkIso(from *Pattern, to *Pattern, split func(*Pattern) (*Pattern, error), unsplit func(*Pattern) (*Pattern, error)) *Iso {
	return &Iso{&Prism{from, split}, &Prism{to, unsplit}}
}

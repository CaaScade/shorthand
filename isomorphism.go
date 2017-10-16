package main

import (
	"github.com/kr/pretty"
	"log"
)

// Prism knows how to split from one pattern to another (possibly multiple)
type Prism struct {
	from  func() *Pattern
	split func(from *Pattern) (*Pattern, error)
}

// View an object through this prism. If the Prism succeeds, it mutated the object.
func (p *Prism) View(i interface{}) (interface{}, error) {
	pat := p.from()
	pat.Match(i)
	if pat.HasErrors() {
		return nil, pretty.Errorf("failed match (%v) with (%v)", pat, i)
	}

	qat, err := p.split(pat)
	if err != nil {
		return nil, err
	}

	i = pat.Erase(i)
	i, err = qat.Write(i)
	if err != nil {
		// A failure during Write indicates a more serious issue
		// than a simple failure to Match.
		log.Fatal(err)

		// Unreachable, but:
		// We return success because the object has been mutated.
	}

	return i, nil
}

// Identity prism.
func Identity() *Prism {
	from := func() *Pattern { return ConstPattern(Wild) }
	split := func(from *Pattern) (*Pattern, error) {
		return ConstPattern(from.capture), nil
	}

	return &Prism{from, split}
}

// ServicePorts prism.
func ServicePorts(portsPrism *Prism) *Prism {
	from := func() *Pattern {
		return MkP(P{"kind": "Service", "spec": P{"ports": Wild}})
	}
	split := func(from *Pattern) (*Pattern, error) {
		// doot
		ports, err := At(from.Extract(), "spec", "ports")
		if err != nil {
			return nil, err
		}

		ports, err = portsPrism.View(ports)
		if err != nil {
			return nil, err
		}

		return MkP(P{"kind": "Service", "spec": P{"ports": ports}}), nil
	}

	return &Prism{from, split}
}

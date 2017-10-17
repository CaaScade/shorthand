package main

import (
	"github.com/kr/pretty"
	"log"
)

// Prism knows how to split from one pattern to another (possibly multiple)
type Prism struct {
	from  *Pattern
	split func(from *Pattern) (*Pattern, error)
}

// View an object through this prism. If the Prism succeeds, it mutated the object.
func (p *Prism) View(i interface{}) (interface{}, error) {
	pat := p.from
	pat.Clear()
	pat.Match(i)
	if pat.HasErrors() {
		return nil, pretty.Errorf("failed match (%# v) with (%# v)", pat, i)
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
	split := func(from *Pattern) (*Pattern, error) {
		return ConstPattern(from.capture), nil
	}

	return &Prism{ConstPattern(Wild), split}
}

// TaggedError doot.
type TaggedError struct {
	tag   interface{}
	error error
}

// Sequence doot.
func Sequence(ps ...*Prism) *Prism {
	if len(ps) == 0 {
		return Identity()
	}

	split := func(from *Pattern) (*Pattern, error) {
		errs := []*TaggedError{}
		i := from.capture
		for ix, p := range ps {
			ii, err := p.View(i)
			if err != nil {
				errs = append(errs, &TaggedError{ix, err})
			} else {
				i = ii
			}
		}

		if len(errs) == len(ps) {
			return nil, MergeErrors(errs)
		}

		return ConstPattern(i), nil
	}

	return &Prism{ConstPattern(Wild), split}
}

// Multiply a prism to apply it to an array.
func Multiply(p *Prism) *Prism {
	split := func(from *Pattern) (*Pattern, error) {
		errs := []*TaggedError{}
		i := from.capture
		switch i := i.(type) {
		case []interface{}:
			for ix, v := range i {
				vv, err := p.View(v)
				if err != nil {
					errs = append(errs, &TaggedError{ix, err})
				} else {
					i[ix] = vv
				}

			}

			if len(errs) == len(i) {
				return nil, MergeErrors(errs)
			}

			return ConstPattern(i), nil
		default:
			return nil, pretty.Errorf("expected slice at (%# v)", i)
		}
	}

	return &Prism{ConstPattern(Wild), split}
}

// MergeErrors combines multiple errors into one.
// TODO: Actually merge the errors.
func MergeErrors(errs []*TaggedError) error {
	if len(errs) > 0 {
		return errs[0].error
	}

	return nil
}

// Zoom doot.
func Zoom(telescope *Pattern, p *Prism) *Prism {
	ks, err := telescope.WildcardPath()
	if err != nil {
		log.Fatal(pretty.Sprint("not a telescope", telescope))
	}

	split := func(from *Pattern) (*Pattern, error) {
		v, err := At(from.Extract(), ks...)
		if err != nil {
			return nil, err
		}

		v, err = p.View(v)
		if err != nil {
			return nil, err
		}

		to := from.Clone()
		err = to.SetConst(v, ks...)
		if err != nil {
			log.Fatal(pretty.Sprintf(
				"couldn't set telescope (%# v):\n%v", to, err))
		}

		return to, nil
	}

	return &Prism{telescope, split}
}

package main

import (
	"bytes"
	"fmt"
)

// RoundTrip try to round-trip a file through the isomorphism.
func RoundTrip(filename string, iso *Iso) error {
	var ws interface{}
	var err error
	var b *bytes.Buffer

	// Load from file.
	ws, err = ReadYamls(filename)
	if err != nil {
		return err
	}

	// Pristine copy.
	b, err = WriteYamls(ws)
	if err != nil {
		return err
	}

	pristine := b.String()

	// Transformed copy.
	ws, err = iso.forward.View(ws)
	if err != nil {
		return err
	}

	b, err = WriteYamls(ws)
	if err != nil {
		return err
	}

	transformed := b.String()

	// Reverted copy.
	ws, err = iso.backward.View(ws)
	if err != nil {
		return err
	}

	b, err = WriteYamls(ws)
	if err != nil {
		return err
	}

	reverted := b.String()

	if pristine != reverted {
		return fmt.Errorf("failed round trip (\n%v) -> (\n%v) -> (\n%v)",
			pristine,
			transformed,
			reverted)
	}

	return nil
}

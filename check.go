package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// RoundTrip try to round-trip a file through the isomorphism.
func RoundTrip(filename string, iso *Iso) (pristine, transformed, reverted string, err error) {
	var ws interface{}
	var b *bytes.Buffer

	// Load from file.
	ws, err = ReadYamls(filename)
	if err != nil {
		return
	}

	// Pristine copy.
	b, err = WriteYamls(ws)
	if err != nil {
		return
	}

	pristine = b.String()

	// Transformed copy.
	ws, err = iso.forward.View(ws)
	if err != nil {
		return
	}

	b, err = WriteYamls(ws)
	if err != nil {
		return
	}

	transformed = b.String()

	// Reverted copy.
	ws, err = iso.backward.View(ws)
	if err != nil {
		return
	}

	b, err = WriteYamls(ws)
	if err != nil {
		return
	}

	reverted = b.String()

	if pristine != reverted {
		err = fmt.Errorf("failed round trip (\n%v) -> (\n%v) -> (\n%v)",
			pristine,
			transformed,
			reverted)
	}

	return
}

// WriteResults write out the results of the round-trip check.
func WriteResults(
	relPath string,
	outDir string,
	failDir string,
	pristine string,
	transformed string,
	reverted string,
	err error) error {
	var writeErr error
	if err != nil {
		failPath := filepath.Join(failDir, relPath)

		writeErr = writeResults(failPath, "-error", fmt.Sprint(err))
		if writeErr != nil {
			return writeErr
		}

		if len(pristine) > 0 {
			writeErr = writeResults(failPath, "-pristine", pristine)
			if writeErr != nil {
				return writeErr
			}
		}

		if len(transformed) > 0 {
			writeErr = writeResults(
				failPath, "-transformed", transformed)
			if writeErr != nil {
				return writeErr
			}
		}

		if len(reverted) > 0 {
			writeErr = writeResults(
				failPath, "-reverted", reverted)
			if writeErr != nil {
				return writeErr
			}
		}

		return nil
	}

	outPath := filepath.Join(outDir, relPath)
	return ioutil.WriteFile(outPath, []byte(transformed), 0644)
}

func writeResults(path string, tag string, contents string) error {
	return ioutil.WriteFile(
		InsertBeforeExt(path, tag),
		[]byte(contents), 0644)
}

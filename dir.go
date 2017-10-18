package main

import (
	"os"
	"path/filepath"
)

// YamlPathsInDir paths to all yaml files in a dir.
func YamlPathsInDir(root string) ([]string, error) {
	paths := []string{}
	err := filepath.Walk(root,
		func(path string, f os.FileInfo, err error) error {
			if filepath.Ext(path) == ".yaml" {
				paths = append(paths, path)
			}

			return nil
		})

	if err != nil {
		return nil, err
	}

	return paths, nil
}

// InsertBeforeExt inserts a tag just before the file extension.
func InsertBeforeExt(path string, tag string) string {
	ext := filepath.Ext(path)
	lext := len(ext)
	lpath := len(path)
	return path[:lpath-lext] + tag + ext
}

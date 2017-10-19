package fs

import (
	"io/ioutil"
	"path/filepath"
)

// StealYamlFiles steal (copy) yaml files from one directory to another.
func StealYamlFiles(fromDir, toDir string) error {
	paths, err := YamlPathsInDir(fromDir)
	for _, path := range paths {
		err = StealYamlFile(fromDir, toDir, path)
		if err != nil {
			return err
		}
	}

	return nil
}

// StealYamlFile steal a yaml file.
func StealYamlFile(fromDir, toDir, fromPath string) error {
	contents, err0 := ioutil.ReadFile(fromPath)
	if err0 != nil {
		return err0
	}

	relPath, err1 := filepath.Rel(fromDir, fromPath)
	if err1 != nil {
		return err1
	}

	toPath := filepath.Join(toDir, relPath)
	return MkdirWriteFile(toPath, contents)
}

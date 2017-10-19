package cmd

import (
	"path/filepath"

	"github.com/koki/shorthand/ast"
	"github.com/koki/shorthand/fs"
	"github.com/koki/shorthand/isos"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
)

// RootCmd root cobra command.
var RootCmd = &cobra.Command{
	Use:   "shorthand [command]",
	Short: "shorthand implements more readable k8s manifests",
}

var stealCmd = &cobra.Command{
	Use:   "steal [from dir] [to dir]",
	Short: "copy all the yaml files from one directory to another",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputDir := args[0]
		outputDir := args[1]
		return fs.StealYamlFiles(inputDir, outputDir)
	},
}

var shrinkCmd = &cobra.Command{
	Use:   "shrink [source dir] [output dir] [failures dir]",
	Short: "convert native manifest to shorthand manifest",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return applyIso(true, args[0], args[1], args[2])
	},
}

var growCmd = &cobra.Command{
	Use:   "grow [source dir] [output dir] [failures dir]",
	Short: "convert shorthand manifest to native manifest",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return applyIso(false, args[0], args[1], args[2])
	},
}

func applyIso(forward bool, inputDir, outputDir, failureDir string) error {
	paths, err := fs.YamlPathsInDir(inputDir)
	if err != nil {
		return err
	}

	iso := isos.ManifestIso()
	if !forward {
		iso = ast.FlipIso(iso)
	}

	for _, path := range paths {
		_, _ = pretty.Println(path)
		var relPath, pristine, transformed, reverted string
		var roundTripErr error
		pristine, transformed, reverted, roundTripErr = RoundTrip(path, iso)

		relPath, err = filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		err = WriteResults(
			relPath,
			outputDir,
			failureDir,
			pristine,
			transformed,
			reverted,
			roundTripErr)

		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	RootCmd.AddCommand(stealCmd, shrinkCmd, growCmd)
}

package cmd

import (
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
)

// RootCmd root cobra command.
var RootCmd = &cobra.Command{
	Use:   "shorthand",
	Short: "shorthand implements more readable k8s manifests",
	Run: func(cmd *cobra.Command, args []string) {
		_, _ = pretty.Println("yes, hello")
		load()
	},
}

func init() {
	RootCmd.AddCommand()
}

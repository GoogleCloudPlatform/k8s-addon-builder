package cmd

import (
	"github.com/spf13/cobra"
)

var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "git utility",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	PlyCmd.AddCommand(GitCmd)
}

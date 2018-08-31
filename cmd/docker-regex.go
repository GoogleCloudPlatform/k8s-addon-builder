package cmd

import (
	"github.com/spf13/cobra"
)

var DockerRegexCmd = &cobra.Command{
	Use:   "docker-regex",
	Short: "docker utility based on regexes",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	PlyCmd.AddCommand(DockerRegexCmd)
}

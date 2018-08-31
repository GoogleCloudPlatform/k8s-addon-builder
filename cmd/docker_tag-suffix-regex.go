package cmd

import (
	"github.com/spf13/cobra"
)

var DockerTagSuffixRegexCmd = &cobra.Command{
	Use: "tag-suffix-regex",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	DockerCmd.AddCommand(DockerTagSuffixRegexCmd)
}

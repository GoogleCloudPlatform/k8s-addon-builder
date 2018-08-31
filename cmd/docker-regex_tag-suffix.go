package cmd

import (
	"github.com/spf13/cobra"
)

var DockerRegexTagSuffixCmd = &cobra.Command{
	Use: "tag-suffix",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	DockerRegexCmd.AddCommand(DockerRegexTagSuffixCmd)
}

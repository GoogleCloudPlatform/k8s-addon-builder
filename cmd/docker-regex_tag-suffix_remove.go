package cmd

import (
	abd "github.com/GoogleCloudPlatform/addon-builder/pkg"

	"github.com/spf13/cobra"
)

var DockerRegexTagSuffixRemoveCmd = &cobra.Command{
	Use:  "remove <REGEX> <TAG_SUFFIX>",
	Args: cobra.ExactArgs(2),
	RunE: removeTagSuffixWrapper,
}

func init() {
	DockerRegexTagSuffixCmd.AddCommand(DockerRegexTagSuffixRemoveCmd)
}

func removeTagSuffixWrapper(cmd *cobra.Command, args []string) error {
	return abd.EditTagSuffixWrapper(cmd, args, false)
}

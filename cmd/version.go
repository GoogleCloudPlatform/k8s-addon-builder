package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VersionDate = "???"
var VersionGit = "???"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nUTC Timestamp: %s\n", VersionGit, VersionDate)
	},
}

func init() {
	ColaCmd.AddCommand(versionCmd)
}

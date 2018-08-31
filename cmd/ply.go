package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// PlyCmd represents the base command when called without any subcommands
var PlyCmd = &cobra.Command{
	Use:   "ply",
	Short: "utility for addon-builder",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the PlyCmd.
func Execute() {
	if err := PlyCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// No global flags (yet).
}

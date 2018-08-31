package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ColaCmd represents the base command when called without any subcommands
var ColaCmd = &cobra.Command{
	Use: "cola",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the ColaCmd.
func Execute() {
	if err := ColaCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// No global flags (yet).
}

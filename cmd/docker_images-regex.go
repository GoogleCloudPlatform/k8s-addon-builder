package cmd

import (
	"fmt"
	"regexp"

	abd "github.com/GoogleCloudPlatform/addon-builder/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var DockerImagesRegexCmd = &cobra.Command{
	Use:  "images-regex <REGEX>",
	Args: cobra.ExactArgs(1),
	RunE: listImages,
}

func init() {
	DockerCmd.AddCommand(DockerImagesRegexCmd)
}

func listImages(cmd *cobra.Command, args []string) error {
	regex := args[0]
	if regex == "" {
		//		cmd.Help()
		return fmt.Errorf("REGEX cannot be empty")
	}
	r, err := regexp.Compile(regex)
	if err != nil {
		return err
	}

	dcli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	images, err := abd.FindImages(dcli, r)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		fmt.Printf("No images match regex %v\n", regex)
		return nil
	}

	fmt.Println("Images found:")
	for k, _ := range images {
		fmt.Printf("  - %v\n", k)
	}
	return nil
}

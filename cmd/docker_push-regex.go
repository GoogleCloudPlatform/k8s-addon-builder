package cmd

import (
	"fmt"
	"os/exec"
	"regexp"

	abd "github.com/GoogleCloudPlatform/addon-builder/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var DockerPushRegexCmd = &cobra.Command{
	Use:  "push-regex <REGEX>",
	Args: cobra.ExactArgs(1),
	RunE: pushWrapper,
}

func init() {
	DockerCmd.AddCommand(DockerPushRegexCmd)
}

func pushWrapper(cmd *cobra.Command, args []string) error {
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
	found, err := abd.FindImages(dcli, r)
	if err != nil {
		return err
	}

	return pushImages(dcli, found)
}

func pushImages(dcli *client.Client, images abd.ImageMap) error {
	if len(images) == 0 {
		fmt.Println("No images to push")
		return nil
	}

	fmt.Println("Images to push:")
	for k, _ := range images {
		fmt.Printf("  - %v\n", k)
	}

	for k, _ := range images {
		cmd := exec.Command("docker", "push", k)
		cmdOut, err := cmd.Output()
		if err != nil {
			return err
		}
		fmt.Println(string(cmdOut))

	}
	return nil
}

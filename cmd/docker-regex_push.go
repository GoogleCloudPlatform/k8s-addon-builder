package cmd

import (
	"fmt"
	"os/exec"

	abd "github.com/GoogleCloudPlatform/addon-builder/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var DockerRegexPushCmd = &cobra.Command{
	Use:  "push <REGEX>",
	Args: cobra.ExactArgs(1),
	RunE: pushWrapper,
}

func init() {
	DockerRegexCmd.AddCommand(DockerRegexPushCmd)
}

func pushWrapper(cmd *cobra.Command, args []string) error {
	r, err := abd.MakeRegex(args[0])

	dcli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	found, err := abd.FindImages(dcli, r)
	if err != nil {
		return err
	}

	return pushImages(found)
}

func pushImages(images abd.ImageMap) error {
	if len(images) == 0 {
		fmt.Println("No images to push")
		return nil
	}

	fmt.Println("Images to push:")
	images.ShowPretty()

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

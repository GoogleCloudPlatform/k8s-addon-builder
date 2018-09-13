// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os/exec"

	abd "github.com/GoogleCloudPlatform/k8s-addon-builder/pkg"
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

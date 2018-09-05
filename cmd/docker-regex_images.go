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
	"regexp"

	abd "github.com/GoogleCloudPlatform/addon-builder/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var DockerRegexImagesCmd = &cobra.Command{
	Use:  "images <REGEX>",
	Args: cobra.ExactArgs(1),
	RunE: listImages,
}

func init() {
	DockerRegexCmd.AddCommand(DockerRegexImagesCmd)
}

func listImages(cmd *cobra.Command, args []string) error {
	regex := args[0]
	if regex == "" {
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

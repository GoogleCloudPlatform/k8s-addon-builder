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
	"strings"

	abd "github.com/GoogleCloudPlatform/k8s-addon-builder/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var DockerRegexLabelImagesCmd = &cobra.Command{
	Use:  "label-images <REGEX>",
	Args: cobra.ExactArgs(1),
	RunE: labelImages,
}

var Labels []string

func init() {
	DockerRegexCmd.AddCommand(DockerRegexLabelImagesCmd)
	DockerRegexLabelImagesCmd.Flags().StringSliceVarP(&Labels, "label", "l", nil, "label to append into an image (can be specified multiple times; required)")
	DockerRegexLabelImagesCmd.MarkFlagRequired("label")
}

func labelImages(cmd *cobra.Command, args []string) error {
	r, err := abd.MakeRegex(args[0])
	if err != nil {
		return err
	}

	dcli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	found, err := abd.FindImages(dcli, r)
	if err != nil {
		return err
	}

	fmt.Println("Labels to add:")
	fmt.Println(Labels)
	labels := make(map[string]string)
	for _, label := range Labels {
		kv := strings.Split(label, "=")
		if len(kv) != 2 {
			return fmt.Errorf("invalid label '%v' (must be of the form 'key=value')", kv)
		}
		if len(kv[0]) == 0 || len(kv[1]) == 0 {
			return fmt.Errorf("invalid label '%v' (must be of the form 'key=value'; both key and value must be non-empty strings)", kv)
		}
		labels[kv[0]] = kv[1]
	}

	if len(labels) == 0 {
		fmt.Println("No labels defined; nothing to do")
		return nil
	}

	fmt.Println("Images to add labels to:")
	found.ShowPretty()

	for _, image := range found.SortedNames() {
		dockerfileContents := "FROM " + image
		fmt.Println(dockerfileContents)
		tags := []string{image}
		abd.BuildImage(dcli, []byte(dockerfileContents), labels, tags)
	}

	return nil
}

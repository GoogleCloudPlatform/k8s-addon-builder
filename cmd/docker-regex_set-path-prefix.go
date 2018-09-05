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
	"strings"

	abd "github.com/GoogleCloudPlatform/addon-builder/pkg"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var DockerRegexSetPathPrefixCmd = &cobra.Command{
	Use:  "set-path-prefix <REGEX> <PATH_PREFIX>",
	Args: cobra.ExactArgs(2),
	RunE: setRegistryWrapper,
}

func init() {
	DockerRegexCmd.AddCommand(DockerRegexSetPathPrefixCmd)
}

func setRegistryWrapper(cmd *cobra.Command, args []string) error {
	regex := args[0]
	pathPrefix := args[1]
	if regex == "" {
		return fmt.Errorf("REGEX cannot be empty")
	}
	if pathPrefix == "" {
		return fmt.Errorf("PATH_PREFIX cannot be empty")
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

	return setPathPrefix(dcli, found, pathPrefix)
}

func setPathPrefix(dcli *client.Client, images abd.ImageMap, pathPrefix string) error {

	imageNames := images.SortedNames()

	tagOps := make([]abd.TagOp, 0)
	for _, imageName := range imageNames {
		ref, err := reference.ParseNormalizedNamed(imageName)
		if err != nil {
			return err
		}

		_, tag, err := abd.GetImageAndTag(imageName)
		if err != nil {
			return err
		}

		refTagged, err := reference.WithTag(ref, tag)
		if err != nil {
			return err
		}

		_, lp, err := splitLastPath(ref.String())
		if err != nil {
			return err
		}
		newRef, err := reference.ParseNormalizedNamed(pathPrefix + "/" + lp)
		if err != nil {
			return err
		}

		newRefTagged, err := reference.WithTag(newRef, tag)
		if err != nil {
			return err
		}

		if newRef == ref {
			fmt.Println("Skipping NOP retag:", ref)
			continue
		}
		tagOps = append(tagOps, abd.TagOp{From: refTagged.String(), To: newRefTagged.String()})
	}

	if len(tagOps) == 0 {
		fmt.Println("No images to modify")
		return nil
	}

	fmt.Println("Images to change:")
	for _, tagOp := range tagOps {
		fmt.Printf("  - %v -> %v\n", tagOp.From, tagOp.To)
	}

	for _, tagOp := range tagOps {
		err := abd.MoveTag(dcli, tagOp)
		if err != nil {
			return err
		}
	}
	return nil
}

// Split a string into 2 parts: everything before and after the last "/".
func splitLastPath(fullName string) (string, string, error) {
	if fullName == "" {
		return "", "", fmt.Errorf("cannot split last path from empty string")
	}
	parts := strings.Split(fullName, "/")
	if len(parts) == 1 {
		return "", parts[0], nil
	}
	pathPrefix := strings.Join(parts[:(len(parts)-1)], "/")
	lastPath := parts[(len(parts) - 1)]
	return pathPrefix, lastPath, nil
}

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

package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

func EditTagSuffixWrapper(cmd *cobra.Command, args []string, appendOrRemove bool) error {
	tagSuffix := args[1]

	if tagSuffix == "" {
		return fmt.Errorf("TAG_SUFFIX cannot be empty")
	}

	r, err := MakeRegex(args[0])
	if err != nil {
		return err
	}

	dcli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	return editTagSuffix(dcli, tagSuffix, appendOrRemove, r)
}

func GetImageAndTag(repoTag string) (string, string, error) {
	imageAndTag := strings.Split(repoTag, ":")
	if len(imageAndTag) != 2 {
		return "", "", fmt.Errorf("divisor ':' not found in RepoTag %v", repoTag)
	}
	return imageAndTag[0], imageAndTag[1], nil
}

// "A tag name must be valid ASCII and may contain lowercase and uppercase
// letters, digits, underscores, periods and dashes. A tag name may not start
// with a period or a dash and may contain a maximum of 128 characters." [1]
//
// [1]: https://docs.docker.com/engine/reference/commandline/tag/
func isValidTag(tag string) bool {
	r, _ := MakeRegex("^[a-zA-Z0-9_][a-zA-Z0-9_.-]+$")
	if !r.MatchString(tag) {
		return false
	}
	if len(tag) > 128 {
		return false
	}
	return true
}

func repoTagExists(dcli *client.Client, needle string) bool {
	images, err := dcli.ImageList(context.Background(), types.ImageListOptions{All: true})
	if err != nil {
		return false
	}

	for _, image := range images {
		for _, repoTag := range image.RepoTags {
			if repoTag == needle {
				return true
			}
		}
	}

	return false
}

type TagOp struct {
	From string
	To   string
}

func appendTag(tagOps []TagOp, dcli *client.Client, tagSuffix string, repoTag string) ([]TagOp, error) {
	imageName, tag, err := GetImageAndTag(repoTag)
	if err != nil {
		return tagOps, err
	}
	// Skip implicit "latest" tag. Images should not be named
	// "latest-<suffix>" (or seen another way, have a "latest-" tag
	// prefix).
	if tag == "latest" {
		fmt.Printf("skipping %v (avoid tagging '%v-%v')\n", repoTag, tag, tagSuffix)
		return tagOps, nil
	}
	if strings.HasSuffix(repoTag, "-"+tagSuffix) {
		fmt.Printf("skipping %v (already has suffix '-%v')\n", repoTag, tagSuffix)
		return tagOps, nil
	}
	var newTag string = tag + "-" + tagSuffix
	if !isValidTag(newTag) {
		return tagOps, fmt.Errorf("new tag %v is invalid", newTag)
	}
	var newRepoTag string = imageName + ":" + newTag
	if repoTagExists(dcli, newRepoTag) {
		fmt.Printf("skipping %v (already suffixed to '-%v')\n", repoTag, tagSuffix)
		return tagOps, nil
	}
	tagOps = append(tagOps, TagOp{repoTag, newRepoTag})
	return tagOps, nil
}

func removeTag(tagOps []TagOp, dcli *client.Client, tagSuffix string, repoTag string) ([]TagOp, error) {
	imageName, tag, err := GetImageAndTag(repoTag)
	if err != nil {
		return tagOps, err
	}
	var newTag string = strings.TrimSuffix(tag, "-"+tagSuffix)
	var newRepoTag string = imageName + ":" + newTag
	if newRepoTag == repoTag {
		fmt.Printf("skipping %v (suffix '-%v' not found)\n", repoTag, tagSuffix)
		return tagOps, nil
	}
	tagOps = append(tagOps, TagOp{repoTag, newRepoTag})
	return tagOps, nil
}

func mkTaggingOperations(dcli *client.Client, tagSuffix string, r *regexp.Regexp, appendOrRemove bool) ([]TagOp, error) {
	images, err := FindImages(dcli, r)
	if err != nil {
		return nil, err
	}

	tagOps := make([]TagOp, 0)
	for _, image := range images {
		// Skip untagged (dangling) images.
		if image.RepoTags[0] == "<none>:<none>" {
			continue
		}
		for _, repoTag := range image.RepoTags {
			if !r.MatchString(repoTag) {
				continue
			}
			if appendOrRemove {
				tagOps, err = appendTag(tagOps, dcli, tagSuffix, repoTag)
			} else {
				tagOps, err = removeTag(tagOps, dcli, tagSuffix, repoTag)
			}
			if err != nil {
				return nil, err
			}
		}
	}

	return tagOps, nil
}

func editTagSuffix(dcli *client.Client, tagSuffix string, appendOrRemove bool, r *regexp.Regexp) error {
	ops, err := mkTaggingOperations(dcli, tagSuffix, r, appendOrRemove)
	if err != nil {
		return err
	}

	if len(ops) == 0 {
		fmt.Printf("Nothing to do.\n")
		return nil
	}

	for _, op := range ops {
		return MoveTag(dcli, op)
	}

	return nil
}

func MoveTag(dcli *client.Client, tagOp TagOp) error {
	err := dcli.ImageTag(context.Background(), tagOp.From, tagOp.To)
	if err != nil {
		return err
	}
	fmt.Printf("tagged from:%v\n         to:%v\n", tagOp.From, tagOp.To)

	responses, err := dcli.ImageRemove(context.Background(), tagOp.From, types.ImageRemoveOptions{})
	for _, res := range responses {
		if len(res.Deleted) > 0 {
			fmt.Printf("deleted: %v\n", res.Deleted)
		}
		if len(res.Untagged) > 0 {
			fmt.Printf("untagged: %v\n", res.Untagged)
		}
	}
	return nil
}

type ImageMap map[string]types.ImageSummary

func FindImages(dcli *client.Client, r *regexp.Regexp) (ImageMap, error) {
	found := make(ImageMap)
	images, err := dcli.ImageList(context.Background(), types.ImageListOptions{All: true})
	if err != nil {
		return nil, err
	}

	for _, image := range images {
		if len(image.RepoTags) == 0 || image.RepoTags[0] == "<none>:<none>" {
			continue
		}
		for _, repoTag := range image.RepoTags {
			if r.MatchString(repoTag) {
				found[repoTag] = image
			}
		}
	}

	return found, nil
}

func BuildImage(dcli *client.Client, dockerFileContents []byte, labels map[string]string, tags []string) error {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerFileContents)),
	}
	err := tw.WriteHeader(tarHeader)
	if err != nil {
		return err
	}
	_, err = tw.Write(dockerFileContents)
	if err != nil {
		return err
	}
	dockerFileTarReader := bytes.NewReader(buf.Bytes())
	ctx := context.Background()

	imageBuildResponse, err := dcli.ImageBuild(
		ctx,
		dockerFileTarReader,
		types.ImageBuildOptions{
			Context: dockerFileTarReader,
			Labels:  labels,
			Tags:    tags,
			// Remove: whether to remove the intermediate container used during
			// the build.
			Remove: true})
	if err != nil {
		return err
	}

	err = PrintStream(ctx, imageBuildResponse.Body)
	if err != nil {
		return err
	}
	return nil
}

type TextStream struct {
	Stream string `json:"stream"`
}

func PrintStream(ctx context.Context, stream io.ReadCloser) error {
	decoder := json.NewDecoder(stream)
	var s TextStream
	for {
		select {
		case <-ctx.Done():
			stream.Close()
			return nil
		default:
			if err := decoder.Decode(&s); err == io.EOF {
				return nil
			} else if err != nil {
				return err
			}
		}
		fmt.Print(s.Stream)
	}
}

func MakeRegex(regex string) (*regexp.Regexp, error) {
	if regex == "" {
		return nil, fmt.Errorf("REGEX cannot be empty")
	}
	return regexp.Compile(regex)
}

func (images ImageMap) SortedNames() []string {
	imageNames := make([]string, 0)
	for imageName, _ := range images {
		imageNames = append(imageNames, imageName)
	}
	sort.Strings(imageNames)
	return imageNames
}

func (images ImageMap) ShowPretty() {
	for _, imageName := range images.SortedNames() {
		fmt.Printf("  - %v\n", imageName)
	}
}

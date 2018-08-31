package docker

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

func EditTagSuffixWrapper(cmd *cobra.Command, args []string, appendOrRemove bool) error {
	regex := args[0]
	tagSuffix := args[1]

	if regex == "" {
		return fmt.Errorf("REGEX cannot be empty")
	}
	if tagSuffix == "" {
		return fmt.Errorf("TAG_SUFFIX cannot be empty")
	}

	r, err := regexp.Compile(regex)
	if err != nil {
		return err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	return editTagSuffix(cli, tagSuffix, appendOrRemove, r)
}

func getImageAndTag(repoTag string) (string, string, error) {
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
	r, _ := regexp.Compile("^[a-zA-Z0-9_][a-zA-Z0-9_.-]+$")
	if !r.MatchString(tag) {
		return false
	}
	if len(tag) > 128 {
		return false
	}
	return true
}

func repoTagExists(cli *client.Client, needle string) bool {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{All: true})
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

type tagOp struct {
	from string
	to   string
}

func appendTag(tagOps []tagOp, cli *client.Client, tagSuffix string, repoTag string) ([]tagOp, error) {
	imageName, tag, err := getImageAndTag(repoTag)
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
	if repoTagExists(cli, newRepoTag) {
		fmt.Printf("skipping %v (already suffixed to '-%v')\n", repoTag, tagSuffix)
		return tagOps, nil
	}
	tagOps = append(tagOps, tagOp{repoTag, newRepoTag})
	return tagOps, nil
}

func removeTag(tagOps []tagOp, cli *client.Client, tagSuffix string, repoTag string) ([]tagOp, error) {
	imageName, tag, err := getImageAndTag(repoTag)
	if err != nil {
		return tagOps, err
	}
	var newTag string = strings.TrimSuffix(tag, "-"+tagSuffix)
	var newRepoTag string = imageName + ":" + newTag
	if newRepoTag == repoTag {
		fmt.Printf("skipping %v (suffix '-%v' not found)\n", repoTag, tagSuffix)
		return tagOps, nil
	}
	tagOps = append(tagOps, tagOp{repoTag, newRepoTag})
	return tagOps, nil
}

func mkTaggingOperations(cli *client.Client, tagSuffix string, r *regexp.Regexp, appendOrRemove bool) ([]tagOp, error) {
	images, err := FindImages(cli, r)
	if err != nil {
		return nil, err
	}

	tagOps := make([]tagOp, 0)
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
				tagOps, err = appendTag(tagOps, cli, tagSuffix, repoTag)
			} else {
				tagOps, err = removeTag(tagOps, cli, tagSuffix, repoTag)
			}
			if err != nil {
				return nil, err
			}
		}
	}

	return tagOps, nil
}

func editTagSuffix(cli *client.Client, tagSuffix string, appendOrRemove bool, r *regexp.Regexp) error {
	ops, err := mkTaggingOperations(cli, tagSuffix, r, appendOrRemove)
	if err != nil {
		return err
	}

	if len(ops) == 0 {
		fmt.Printf("Nothing to do.\n")
		return nil
	}

	for _, op := range ops {
		err := cli.ImageTag(context.Background(), op.from, op.to)
		if err != nil {
			return err
		}
		fmt.Printf("tagged from:%v\n         to:%v\n", op.from, op.to)

		responses, err := cli.ImageRemove(context.Background(), op.from, types.ImageRemoveOptions{})
		for _, res := range responses {
			if len(res.Deleted) > 0 {
				fmt.Printf("deleted: %v\n", res.Deleted)
			}
			if len(res.Untagged) > 0 {
				fmt.Printf("untagged: %v\n", res.Untagged)
			}
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

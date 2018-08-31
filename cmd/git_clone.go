package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var GitCloneCmd = &cobra.Command{
	Use:  "clone <REPO_URL>",
	Args: cobra.ExactArgs(1),
	RunE: cloneAndCheckout,
}

var Dir string
var Rev string

func init() {
	GitCmd.AddCommand(GitCloneCmd)
	GitCloneCmd.Flags().StringVarP(&Dir, "dir", "d", "", "directory to clone into (by default uses the name of the git repo)")
	GitCloneCmd.Flags().StringVarP(&Rev, "rev", "r", "", "revision to check out after the clone (defaults to master branch)")
}

func cloneAndCheckout(cmd *cobra.Command, args []string) error {
	repoUrl := args[0]
	var ecmd *exec.Cmd
	var path string
	if len(Dir) > 0 {
		path = Dir
	} else {
		path = getRepoName(repoUrl)
	}
	ecmd = exec.Command("git", "clone", repoUrl, path)

	if err := ecmd.Run(); err != nil {
		return err
	}

	err := checkout(cmd, path)
	if err != nil {
		return err
	}

	return nil
}

func checkout(cmd *cobra.Command, path string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	if len(Rev) > 0 {
		worktree, err := repo.Worktree()
		if err != nil {
			return err
		}
		var checkoutOptions git.CheckoutOptions

		if isHash(Rev) {
			fmt.Println("checking out hash:", Rev)
			checkoutOptions = git.CheckoutOptions{Hash: plumbing.NewHash(Rev)}
		} else {
			fmt.Println("checking out branch:", Rev)
			checkoutOptions = git.CheckoutOptions{Branch: plumbing.ReferenceName("refs/remotes/origin/" + Rev)}
		}
		err = worktree.Checkout(&checkoutOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

func isHash(rev string) bool {
	r, err := regexp.Compile("^[a-fA-F0-9]{40}$")
	if err != nil {
		return false
	}
	return r.MatchString(rev)
}

// Get the repository name; drop the last ".git" suffix if found.
func getRepoName(repoUrl string) string {
	parts := strings.Split(repoUrl, "/")
	last := parts[len(parts)-1]
	if strings.HasSuffix(last, ".git") {
		return last[:len(last)-4]
	}
	return last
}

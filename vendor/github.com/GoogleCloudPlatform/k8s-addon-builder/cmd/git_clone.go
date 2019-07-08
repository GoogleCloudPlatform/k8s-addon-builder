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

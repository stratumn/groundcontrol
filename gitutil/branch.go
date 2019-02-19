// Copyright 2019 Stratumn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitutil

import (
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// CurrentBranch returns the current branch of a repository.
// It is possible that there isn't a current branch, in which case it returns null.
func CurrentBranch(repo *git.Repository, defaultMerge plumbing.ReferenceName) (*config.Branch, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	if !head.Name().IsBranch() {
		return nil, nil
	}

	branchName := RefBranchName(head)

	branch, err := repo.Branch(branchName)

	if err == git.ErrBranchNotFound {
		// Branch tracking is not configured.
		return &config.Branch{
			Remote: "origin",
			Name:   branchName,
			Merge:  defaultMerge,
		}, nil
	}

	return branch, err
}

// RefBranchName returns the branch name of a reference.
// It assumes that the ref has a branch type.
func RefBranchName(ref *plumbing.Reference) string {
	return RefBranchNameStr(ref.String())
}

// RefBranchNameStr returns the branch name of a reference string.
// It assumes that the ref has a branch type.
func RefBranchNameStr(str string) string {
	parts := strings.Split(str, "/")
	return strings.Join(parts[2:], "/")
}

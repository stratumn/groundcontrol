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

package model

import (
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"groundcontrol/relay"
)

// NewCommitFromGit creates a new Commit model from a Git commit.
func NewCommitFromGit(commit *object.Commit) *Commit {
	return &Commit{
		ID:       relay.EncodeID(NodeTypeCommit, commit.Hash.String()),
		Hash:     Hash(commit.Hash.String()),
		Headline: strings.Split(commit.Message, "\n")[0],
		Message:  commit.Message,
		Author:   commit.Author.Name,
		Date:     DateTime(commit.Author.When),
	}
}

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
	"bytes"
	"context"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

// HasAncestor returns whether a commit contains another commit in its history.
// It returns false if the two hashes are the same.
func HasAncestor(ctx context.Context, repo *git.Repository, child plumbing.Hash, ancestor plumbing.Hash) (bool, error) {
	if bytes.Equal(child[:], ancestor[:]) {
		return false, nil
	}
	iter, err := repo.Log(&git.LogOptions{From: child})
	if err != nil {
		return false, err
	}
	defer iter.Close()
	has := false
	err = iter.ForEach(func(commit *object.Commit) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if bytes.Equal(ancestor[:], commit.Hash[:]) {
			has = true
			return storer.ErrStop
		}
		return nil
	})
	return has, err
}

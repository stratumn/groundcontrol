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

package queue

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPriorities(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var results []int

	q := New(1)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go q.Do(func() {
		results = append(results, 1)
		wg.Done()
	})

	go q.DoHi(func() {
		results = append(results, 2)
		wg.Done()
	})

	time.Sleep(10 * time.Millisecond)

	go q.Work(ctx)

	go func() {
		wg.Wait()
		cancel()
	}()

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("jobs didn't run")
	}

	assert.Equal(t, []int{2, 1}, results, "job order")
}

func TestDoError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := New(1)
	go q.Work(ctx)

	errCh := make(chan error)

	go func() {
		errCh <- q.DoError(func() error {
			return errors.New("")
		})
	}()

	select {
	case <-time.After(time.Second):
		t.Fatal("jobs didn't run")
	case err := <-errCh:
		assert.Error(t, err)
	}
}

func TestDoErrorHi(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := New(1)
	go q.Work(ctx)

	errCh := make(chan error)

	go func() {
		errCh <- q.DoErrorHi(func() error {
			return errors.New("")
		})
	}()

	select {
	case <-time.After(time.Second):
		t.Fatal("jobs didn't run")
	case err := <-errCh:
		assert.Error(t, err)
	}
}

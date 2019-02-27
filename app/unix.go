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

// +build !windows

package app

import (
	"context"
	"groundcontrol/model"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func initHooks(ctx context.Context) error {
	incNoFile(ctx)
	return nil
}

func incNoFile(ctx context.Context) {
	modelCtx := model.GetContext(ctx)
	log := modelCtx.Log
	systemID := modelCtx.SystemID
	limit := syscall.Rlimit{}

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		log.WarningWithOwner(
			ctx,
			systemID,
			"failed to get maximum number of open files because %s",
			err.Error(),
		)
		return
	}

	max, err := maxNoFile()
	if err != nil {
		log.WarningWithOwner(
			ctx,
			systemID,
			"failed to get maximum number of open files because %s",
			err.Error(),
		)
		return
	}

	was := limit.Cur
	limit.Cur = max

	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		log.WarningWithOwner(
			ctx,
			systemID,
			"failed to set maximum number of open files because %s",
			err.Error(),
		)
		return
	}

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		log.WarningWithOwner(
			ctx,
			systemID,
			"failed to get maximum number of open files because %s",
			err.Error(),
		)
		return
	}

	log.InfoWithOwner(
		ctx,
		systemID,
		"maximum number of open files increased from %d to %d",
		was,
		limit.Cur,
	)
}

func maxNoFile() (uint64, error) {
	n, err := ulimit()
	if err != nil {
		limit := syscall.Rlimit{}
		if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
			return 0, err
		}
		n = limit.Max
	}

	return n, nil
}

func ulimit() (uint64, error) {
	cmd := exec.Command("ulimit", "-n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	n, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, err
	}

	return n, nil
}

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
	"syscall"

	"groundcontrol/appcontext"
)

func initHooks(ctx context.Context) error {
	incNofile(ctx)
	return nil
}

// incNofile attempts to increase the maximum allowed number of open files,
// which is especially helpful when file watchers are used by services.
func incNofile(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID
	limit := syscall.Rlimit{}
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		log.WarningWithOwner(ctx, systemID, "failed to get maximum number of open files because %s", err.Error())
		return
	}
	was := limit.Cur
	limit.Cur = limit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		// Max on some macOS.
		limit.Cur = 12288
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
			log.WarningWithOwner(ctx, systemID, "failed to set maximum number of open files because %s", err.Error())
			return
		}
	}
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		log.WarningWithOwner(ctx, systemID, "failed to get maximum number of open files because %s", err.Error())
		return
	}
	log.InfoWithOwner(ctx, systemID, "maximum number of open files increased from %d to %d", was, limit.Cur)
}

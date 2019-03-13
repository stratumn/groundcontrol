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

package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
)

func execCmd(ctx context.Context, path string, args []string) error {
	moduleCtx, _ := interp.FromModuleContext(ctx)
	if path == "" {
		fmt.Fprintf(moduleCtx.Stderr, "%q: executable file not found in $PATH\n", args[0])
		return interp.ExitStatus(127)
	}
	cmd := createCmd(ctx, path, args)
	err := cmd.Start()
	if err == nil {
		exitCh := make(chan struct{}, 1)
		go func() {
			<-ctx.Done()
			go func() {
				select {
				case <-exitCh:
					return
				case <-time.After(moduleCtx.KillTimeout):
					_ = sendSignalToCmd(cmd, os.Kill)
				}
			}()
			_ = sendSignalToCmd(cmd, os.Interrupt)
		}()
		err = cmd.Wait()
		close(exitCh)
	}
	return handleCmdErr(ctx, err)
}

func createCmd(ctx context.Context, path string, args []string) *exec.Cmd {
	moduleCtx, _ := interp.FromModuleContext(ctx)
	return &exec.Cmd{
		Path:        path,
		Args:        args,
		Env:         execEnv(moduleCtx.Env),
		Dir:         moduleCtx.Dir,
		Stdin:       moduleCtx.Stdin,
		Stdout:      moduleCtx.Stdout,
		Stderr:      moduleCtx.Stderr,
		SysProcAttr: createCmdSysProcAttr(),
	}
}

func handleCmdErr(ctx context.Context, err error) error {
	moduleCtx, _ := interp.FromModuleContext(ctx)
	switch x := err.(type) {
	case *exec.ExitError:
		if status, ok := x.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() && ctx.Err() != nil {
				return ctx.Err()
			}
			return interp.ExitStatus(status.ExitStatus())
		}
		return interp.ExitStatus(1)
	case *exec.Error:
		fmt.Fprintf(moduleCtx.Stderr, "%v\n", err)
		return interp.ExitStatus(127)
	}
	return err
}

func execEnv(env expand.Environ) []string {
	list := make([]string, 0, 32)
	env.Each(func(name string, vr expand.Variable) bool {
		if vr.Exported {
			list = append(list, name+"="+vr.String())
		}
		return true
	})
	return list
}

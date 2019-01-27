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

package models

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/stratumn/groundcontrol/date"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
)

// ProcessManager manages creating and running jobs.
type ProcessManager struct {
	nodes *NodeManager
	log   *Logger
	subs  *pubsub.PubSub

	systemID string

	nextID    uint64
	processes map[string]*os.Process

	runningCounter int64
	doneCounter    int64
	failedCounter  int64
}

// NewProcessManager creates a ProcessManager.
func NewProcessManager(
	nodes *NodeManager,
	log *Logger,
	subs *pubsub.PubSub,
	systemID string,
) *ProcessManager {
	return &ProcessManager{
		nodes:     nodes,
		log:       log,
		subs:      subs,
		systemID:  systemID,
		processes: map[string]*os.Process{},
	}
}

// CreateGroup creates a new ProcessGroup and returns its ID.
func (p *ProcessManager) CreateGroup(taskID string) string {
	id := relay.EncodeID(
		NodeTypeProcessGroup,
		fmt.Sprint(atomic.AddUint64(&p.nextID, 1)),
	)

	group := ProcessGroup{
		ID:        id,
		CreatedAt: date.NowFormatted(),
		TaskID:    taskID,
	}

	p.nodes.MustStoreProcessGroup(group)
	p.subs.Publish(ProcessGroupUpserted, id)

	p.nodes.MustLockSystem(p.systemID, func(system System) {
		system.ProcessGroupIDs = append(
			[]string{id},
			system.ProcessGroupIDs...,
		)

		p.nodes.MustStoreSystem(system)
	})

	return id
}

// Run launches a new Process and adds it to a ProcessGroup.
func (p *ProcessManager) Run(
	command string,
	processGroupID string,
	projectID string,
	dir string,
) string {
	id := relay.EncodeID(
		NodeTypeProcess,
		fmt.Sprint(atomic.AddUint64(&p.nextID, 1)),
	)

	meta := struct {
		ProcessID      string
		ProcessGroupID string
		ProjectID      string
		Dir            string
		Command        string
		Error          string
	}{
		id,
		processGroupID,
		projectID,
		dir,
		command,
		"",
	}

	process := Process{
		ID:             id,
		Command:        command,
		ProcessGroupID: processGroupID,
		ProjectID:      projectID,
	}

	stdout := CreateLineWriter(p.log.Info, meta)
	stderr := CreateLineWriter(p.log.Warning, meta)
	cmd := exec.Command("bash", "-l", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Start()

	if err == nil {
		process.Status = ProcessStatusRunning
		atomic.AddInt64(&p.runningCounter, 1)
	} else {
		process.Status = ProcessStatusFailed
		meta.Error = err.Error()
		atomic.AddInt64(&p.failedCounter, 1)
	}

	p.nodes.MustStoreProcess(process)
	p.nodes.MustLockProcessGroup(processGroupID, func(processGroup ProcessGroup) {
		processGroup.ProcessIDs = append([]string{id}, processGroup.ProcessIDs...)
		p.nodes.MustStoreProcessGroup(processGroup)
	})

	p.subs.Publish(ProcessUpserted, id)
	p.subs.Publish(ProcessGroupUpserted, processGroupID)
	p.publishMetrics()

	if err != nil {
		p.log.Error("Process Failed", meta)
		stdout.Close()
		stderr.Close()
		return id
	}

	p.log.Info("Process Running", meta)
	// cmd.Process

	go func() {
		err := cmd.Wait()

		p.nodes.MustLockProcess(id, func(process Process) {
			if err == nil {
				process.Status = ProcessStatusRunning
				atomic.AddInt64(&p.doneCounter, 1)
				p.log.Info("Process Done", meta)
			} else {
				process.Status = ProcessStatusFailed
				meta.Error = err.Error()
				atomic.AddInt64(&p.failedCounter, 1)
				p.log.Error("Process Failed", meta)
			}

			atomic.AddInt64(&p.runningCounter, -1)
			p.nodes.MustStoreProcess(process)
		})

		p.subs.Publish(ProcessUpserted, id)
		p.subs.Publish(ProcessGroupUpserted, processGroupID)
		p.publishMetrics()

		stdout.Close()
		stderr.Close()
	}()

	return id
}

func (p *ProcessManager) publishMetrics() {
	system := p.nodes.MustLoadSystem(p.systemID)

	p.nodes.MustLockProcessMetrics(system.ProcessMetricsID, func(metrics ProcessMetrics) {
		metrics.Running = int(atomic.LoadInt64(&p.runningCounter))
		metrics.Done = int(atomic.LoadInt64(&p.doneCounter))
		metrics.Failed = int(atomic.LoadInt64(&p.failedCounter))
		p.nodes.MustStoreProcessMetrics(metrics)
	})

	p.subs.Publish(ProcessMetricsUpdated, system.ProcessMetricsID)
}

// CreateLineWriter creates a writer with a line splitter.
// Remember to call close().
func CreateLineWriter(
	write func(string, interface{}) string,
	meta interface{},
) io.WriteCloser {
	r, w := io.Pipe()
	scanner := bufio.NewScanner(r)

	go func() {
		for scanner.Scan() {
			write(scanner.Text(), meta)

			// Don't kill the poor browser.
			time.Sleep(10 * time.Millisecond)
		}
	}()

	return w
}

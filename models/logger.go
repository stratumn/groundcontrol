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
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"groundcontrol/relay"
	"groundcontrol/util"
)

var logLevelPriorities = map[LogLevel]int{
	LogLevelDebug:   0,
	LogLevelInfo:    1,
	LogLevelWarning: 2,
	LogLevelError:   3,
}

// Logger logs messages.
type Logger struct {
	cap   int
	level LogLevel

	lastID      uint64
	logEntryIDs []string
	head        int

	debugCounter   int64
	infoCounter    int64
	warningCounter int64
	errorCounter   int64

	stdoutLog *log.Logger
	stderrLog *log.Logger
}

// NewLogger creates a Logger with given capacity and level.
func NewLogger(cap int, level LogLevel) *Logger {
	return &Logger{
		cap:         cap,
		level:       level,
		logEntryIDs: make([]string, cap*2),
		stdoutLog:   log.New(os.Stdout, "", log.LstdFlags),
		stderrLog:   log.New(os.Stderr, "", log.LstdFlags),
	}
}

// Add adds a log entry.
func (l *Logger) Add(
	ctx context.Context,
	level LogLevel,
	ownerID string,
	message string,
) (string, error) {
	modelCtx := GetModelContext(ctx)

	if logLevelPriorities[level] < logLevelPriorities[l.level] {
		return "", nil
	}

	log := l.stdoutLog
	if logLevelPriorities[level] >= logLevelPriorities[LogLevelWarning] {
		log = l.stderrLog
	}

	if ownerID != "" {
		log.Printf("%s\t<%s> %s", level, ownerID, message)
	} else {
		log.Printf("%s\t%s", level, message)
	}

	id := atomic.AddUint64(&l.lastID, 1)
	now := DateTime(time.Now())
	logEntry := LogEntry{
		ID:        relay.EncodeID(NodeTypeLogEntry, fmt.Sprint(id)),
		Level:     level,
		CreatedAt: now,
		Message:   message,
		OwnerID:   ownerID,
	}
	l.matchSourceFile(ctx, &logEntry)
	logEntry.MustStore(ctx)

	MustLockSystem(ctx, modelCtx.SystemID, func(system System) {
		if l.head >= l.cap*2 {
			copy(l.logEntryIDs, l.logEntryIDs[l.cap:])
			l.head = l.cap

			for _, oldEntryID := range l.logEntryIDs[l.head:] {
				oldEntry := MustLoadLogEntry(ctx, oldEntryID)

				switch oldEntry.Level {
				case LogLevelDebug:
					atomic.AddInt64(&l.debugCounter, -1)
				case LogLevelInfo:
					atomic.AddInt64(&l.infoCounter, -1)
				case LogLevelWarning:
					atomic.AddInt64(&l.warningCounter, -1)
				case LogLevelError:
					atomic.AddInt64(&l.errorCounter, -1)
				}

				MustDeleteLogEntry(ctx, oldEntryID)
			}
		}

		l.logEntryIDs[l.head] = logEntry.ID

		l.head++

		system.LogEntryIDs = l.logEntryIDs[:l.head]
		system.MustStore(ctx)
	})

	switch level {
	case LogLevelDebug:
		atomic.AddInt64(&l.debugCounter, 1)
	case LogLevelInfo:
		atomic.AddInt64(&l.infoCounter, 1)
	case LogLevelWarning:
		atomic.AddInt64(&l.warningCounter, 1)
	case LogLevelError:
		atomic.AddInt64(&l.errorCounter, 1)
	}

	l.updateMetrics(ctx)

	return logEntry.ID, nil
}

// Debug adds a debug entry.
func (l *Logger) Debug(ctx context.Context, message string, a ...interface{}) string {
	id, err := l.Add(ctx, LogLevelDebug, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// Info adds an info entry.
func (l *Logger) Info(ctx context.Context, message string, a ...interface{}) string {
	id, err := l.Add(ctx, LogLevelInfo, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// Warning adds a warning entry.
func (l *Logger) Warning(ctx context.Context, message string, a ...interface{}) string {
	id, err := l.Add(ctx, LogLevelWarning, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// Error adds an error entry.
func (l *Logger) Error(ctx context.Context, message string, a ...interface{}) string {
	id, err := l.Add(ctx, LogLevelError, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// DebugWithOwner adds a debug entry with an owner.
func (l *Logger) DebugWithOwner(
	ctx context.Context,
	ownerID string,
	message string,
	a ...interface{},
) string {
	id, err := l.Add(ctx, LogLevelDebug, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// InfoWithOwner adds an info entry with an owner.
func (l *Logger) InfoWithOwner(
	ctx context.Context,
	ownerID string,
	message string,
	a ...interface{},
) string {
	id, err := l.Add(ctx, LogLevelInfo, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// WarningWithOwner adds a warning entry with an owner.
func (l *Logger) WarningWithOwner(
	ctx context.Context,
	ownerID string,
	message string,
	a ...interface{},
) string {
	id, err := l.Add(ctx, LogLevelWarning, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// ErrorWithOwner adds an error entry with an owner.
func (l *Logger) ErrorWithOwner(
	ctx context.Context,
	ownerID string,
	message string,
	a ...interface{},
) string {
	id, err := l.Add(ctx, LogLevelError, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

func (l *Logger) updateMetrics(ctx context.Context) {
	modelCtx := GetModelContext(ctx)
	system := MustLoadSystem(ctx, modelCtx.SystemID)

	MustLockLogMetrics(ctx, system.LogMetricsID, func(metrics LogMetrics) {
		metrics.Debug = int(atomic.LoadInt64(&l.debugCounter))
		metrics.Info = int(atomic.LoadInt64(&l.infoCounter))
		metrics.Warning = int(atomic.LoadInt64(&l.warningCounter))
		metrics.Error = int(atomic.LoadInt64(&l.errorCounter))
		metrics.MustStore(ctx)
	})
}

func (l *Logger) matchSourceFile(ctx context.Context, entry *LogEntry) {
	modelCtx := GetModelContext(ctx)

	if entry.OwnerID == "" {
		return
	}

	node, ok := modelCtx.Nodes.Load(entry.OwnerID)
	if !ok {
		return
	}

	project, ok := node.(Project)
	if !ok {
		return
	}

	sourceFile, begin, end, err := util.MatchSourceFile(entry.Message)
	if err != nil {
		return
	}

	sourceParts := strings.Split(sourceFile, ":")
	fileName := sourceParts[0]

	workspace := MustLoadWorkspace(ctx, project.WorkspaceID)
	projectPath := modelCtx.GetProjectPath(workspace.Slug, project.Slug)

	if !filepath.IsAbs(fileName) {
		fileName, err = filepath.Abs(filepath.Join(projectPath, fileName))
		if err != nil {
			return
		}
	}

	fileName = filepath.Clean(fileName)
	sourceFile = strings.Join(append([]string{fileName}, sourceParts[1:]...), ":")

	if !util.FileExists(fileName) {
		return
	}

	if util.IsDirectory(fileName) {
		return
	}

	entry.SourceFile = &sourceFile
	entry.SourceFileBegin = &begin
	entry.SourceFileEnd = &end
}

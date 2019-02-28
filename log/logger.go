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

package log

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"groundcontrol/appcontext"
	"groundcontrol/model"
	"groundcontrol/relay"
	"groundcontrol/util"
)

var logLevelPriorities = map[model.LogLevel]int{
	model.LogLevelDebug:   0,
	model.LogLevelInfo:    1,
	model.LogLevelWarning: 2,
	model.LogLevelError:   3,
}

// Logger logs messages.
type Logger struct {
	cap   int
	level model.LogLevel

	lastID        uint64
	logEntriesIDs []string
	head          int

	debugCounter   int64
	infoCounter    int64
	warningCounter int64
	errorCounter   int64

	stdoutLog *log.Logger
	stderrLog *log.Logger
}

// NewLogger creates a Logger with given capacity and level.
func NewLogger(cap int, level model.LogLevel) *Logger {
	return &Logger{
		cap:           cap,
		level:         level,
		logEntriesIDs: make([]string, cap*2),
		stdoutLog:     log.New(os.Stdout, "", log.LstdFlags),
		stderrLog:     log.New(os.Stderr, "", log.LstdFlags),
	}
}

// Debug adds a debug entry.
func (l *Logger) Debug(ctx context.Context, message string, a ...interface{}) string {
	id, err := l.add(ctx, model.LogLevelDebug, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// Info adds an info entry.
func (l *Logger) Info(ctx context.Context, message string, a ...interface{}) string {
	id, err := l.add(ctx, model.LogLevelInfo, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// Warning adds a warning entry.
func (l *Logger) Warning(ctx context.Context, message string, a ...interface{}) string {
	id, err := l.add(ctx, model.LogLevelWarning, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// Error adds an error entry.
func (l *Logger) Error(ctx context.Context, message string, a ...interface{}) string {
	id, err := l.add(ctx, model.LogLevelError, "", fmt.Sprintf(message, a...))
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
	id, err := l.add(ctx, model.LogLevelDebug, ownerID, fmt.Sprintf(message, a...))
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
	id, err := l.add(ctx, model.LogLevelInfo, ownerID, fmt.Sprintf(message, a...))
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
	id, err := l.add(ctx, model.LogLevelWarning, ownerID, fmt.Sprintf(message, a...))
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
	id, err := l.add(ctx, model.LogLevelError, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

func (l *Logger) updateMetrics(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	system := model.MustLoadSystem(ctx, appCtx.SystemID)

	model.MustLockLogMetrics(ctx, system.LogMetricsID, func(metrics *model.LogMetrics) {
		metrics.Debug = int(atomic.LoadInt64(&l.debugCounter))
		metrics.Info = int(atomic.LoadInt64(&l.infoCounter))
		metrics.Warning = int(atomic.LoadInt64(&l.warningCounter))
		metrics.Error = int(atomic.LoadInt64(&l.errorCounter))
		metrics.MustStore(ctx)
	})
}

func (l *Logger) matchSourceFile(ctx context.Context, entry *model.LogEntry) {
	appCtx := appcontext.Get(ctx)

	if entry.OwnerID == "" {
		return
	}

	node, ok := appCtx.Nodes.Load(entry.OwnerID)
	if !ok {
		return
	}

	project, ok := node.(*model.Project)
	if !ok {
		return
	}

	sourceFile, begin, end, err := util.MatchSourceFile(entry.Message)
	if err != nil {
		return
	}

	sourceParts := strings.Split(sourceFile, ":")
	fileName := sourceParts[0]

	workspace := model.MustLoadWorkspace(ctx, project.WorkspaceID)
	projectPath := appCtx.GetProjectPath(workspace.Slug, project.Slug)

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

func (l *Logger) add(
	ctx context.Context,
	level model.LogLevel,
	ownerID string,
	message string,
) (string, error) {
	appCtx := appcontext.Get(ctx)

	if logLevelPriorities[level] < logLevelPriorities[l.level] {
		return "", nil
	}

	log := l.stdoutLog
	if logLevelPriorities[level] >= logLevelPriorities[model.LogLevelWarning] {
		log = l.stderrLog
	}

	if ownerID != "" {
		log.Printf("%s\t<%s> %s", level, ownerID, message)
	} else {
		log.Printf("%s\t%s", level, message)
	}

	id := atomic.AddUint64(&l.lastID, 1)
	now := model.DateTime(time.Now())
	logEntry := model.LogEntry{
		ID:        relay.EncodeID(model.NodeTypeLogEntry, fmt.Sprint(id)),
		Level:     level,
		CreatedAt: now,
		Message:   message,
		OwnerID:   ownerID,
	}
	l.matchSourceFile(ctx, &logEntry)
	logEntry.MustStore(ctx)

	model.MustLockSystem(ctx, appCtx.SystemID, func(system *model.System) {
		if l.head >= l.cap*2 {
			copy(l.logEntriesIDs, l.logEntriesIDs[l.cap:])
			l.head = l.cap

			for _, oldEntryID := range l.logEntriesIDs[l.head:] {
				oldEntry := model.MustLoadLogEntry(ctx, oldEntryID)

				switch oldEntry.Level {
				case model.LogLevelDebug:
					atomic.AddInt64(&l.debugCounter, -1)
				case model.LogLevelInfo:
					atomic.AddInt64(&l.infoCounter, -1)
				case model.LogLevelWarning:
					atomic.AddInt64(&l.warningCounter, -1)
				case model.LogLevelError:
					atomic.AddInt64(&l.errorCounter, -1)
				}

				model.MustDeleteLogEntry(ctx, oldEntryID)
			}
		}

		l.logEntriesIDs[l.head] = logEntry.ID

		l.head++

		system.LogEntriesIDs = l.logEntriesIDs[:l.head]
		system.MustStore(ctx)
	})

	switch level {
	case model.LogLevelDebug:
		atomic.AddInt64(&l.debugCounter, 1)
	case model.LogLevelInfo:
		atomic.AddInt64(&l.infoCounter, 1)
	case model.LogLevelWarning:
		atomic.AddInt64(&l.warningCounter, 1)
	case model.LogLevelError:
		atomic.AddInt64(&l.errorCounter, 1)
	}

	l.updateMetrics(ctx)

	return logEntry.ID, nil
}

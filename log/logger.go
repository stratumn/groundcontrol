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
	"sync/atomic"
	"time"

	"groundcontrol/appcontext"
	"groundcontrol/model"
	"groundcontrol/relay"
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
func (l *Logger) DebugWithOwner(ctx context.Context, ownerID string, message string, a ...interface{}) string {
	id, err := l.add(ctx, model.LogLevelDebug, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// InfoWithOwner adds an info entry with an owner.
func (l *Logger) InfoWithOwner(ctx context.Context, ownerID string, message string, a ...interface{}) string {
	id, err := l.add(ctx, model.LogLevelInfo, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// WarningWithOwner adds a warning entry with an owner.
func (l *Logger) WarningWithOwner(ctx context.Context, ownerID string, message string, a ...interface{}) string {
	id, err := l.add(ctx, model.LogLevelWarning, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// ErrorWithOwner adds an error entry with an owner.
func (l *Logger) ErrorWithOwner(ctx context.Context, ownerID string, message string, a ...interface{}) string {
	id, err := l.add(ctx, model.LogLevelError, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

func (l *Logger) add(ctx context.Context, level model.LogLevel, ownerID string, message string) (string, error) {
	if logLevelPriorities[level] < logLevelPriorities[l.level] {
		return "", nil
	}
	l.printToStdLog(level, ownerID, message)
	id := atomic.AddUint64(&l.lastID, 1)
	relayID := relay.EncodeID(model.NodeTypeLogEntry, fmt.Sprint(id))
	now := model.DateTime(time.Now())
	l.append(ctx, &model.LogEntry{
		ID:        relayID,
		Level:     level,
		CreatedAt: now,
		Message:   message,
		OwnerID:   ownerID,
	})
	return relayID, nil
}

func (l *Logger) append(ctx context.Context, entry *model.LogEntry) {
	entry.MustStore(ctx)
	systemID := appcontext.Get(ctx).SystemID
	model.MustLockSystem(ctx, systemID, func(system *model.System) {
		if l.head >= l.cap*2 {
			l.deleteOldEntries(ctx)
		}
		l.logEntriesIDs[l.head] = entry.ID
		l.head++
		system.LogEntriesIDs = l.logEntriesIDs[:l.head]
		system.MustStore(ctx)
	})
	l.incCounter(entry.Level)
	l.storeMetrics(ctx)
}

func (l *Logger) deleteOldEntries(ctx context.Context) {
	copy(l.logEntriesIDs, l.logEntriesIDs[l.cap:])
	l.head = l.cap
	for _, oldEntryID := range l.logEntriesIDs[l.head:] {
		oldEntry := model.MustLoadLogEntry(ctx, oldEntryID)
		model.MustDeleteLogEntry(ctx, oldEntryID)
		l.decCounter(oldEntry.Level)
	}
}

func (l *Logger) printToStdLog(level model.LogLevel, ownerID, message string) {
	log := l.stdoutLog
	if logLevelPriorities[level] >= logLevelPriorities[model.LogLevelWarning] {
		log = l.stderrLog
	}
	if ownerID != "" {
		log.Printf("%s\t<%s> %s", level, ownerID, message)
		return
	}
	log.Printf("%s\t%s", level, message)
}

func (l *Logger) incCounter(level model.LogLevel) {
	l.addToCounter(level, 1)
}

func (l *Logger) decCounter(level model.LogLevel) {
	l.addToCounter(level, -1)
}

func (l *Logger) addToCounter(level model.LogLevel, delta int64) {
	switch level {
	case model.LogLevelDebug:
		atomic.AddInt64(&l.debugCounter, delta)
	case model.LogLevelInfo:
		atomic.AddInt64(&l.infoCounter, delta)
	case model.LogLevelWarning:
		atomic.AddInt64(&l.warningCounter, delta)
	case model.LogLevelError:
		atomic.AddInt64(&l.errorCounter, delta)
	}
}

func (l *Logger) storeMetrics(ctx context.Context) {
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

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
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"groundcontrol/appcontext"
	"groundcontrol/util"
)

// String is a string representation for the type instance.
func (n *LogEntry) String() string {
	return n.Message
}

// LongString is a long string representation for the type instance.
func (n *LogEntry) LongString(ctx context.Context) string {
	nodes := appcontext.Get(ctx).Nodes
	if n.OwnerID != "" {
		if owner, ok := nodes.Load(n.OwnerID); ok {
			if owner, ok := owner.(LongStringer); ok {
				return fmt.Sprintf("%-7s  %s  %s", n.Level, owner.LongString(ctx), n)
			}
			return fmt.Sprintf("%-7s  %s  %s", n.Level, owner, n)
		}
		return fmt.Sprintf("%-7s  %s  %s", n.Level, n.OwnerID, n)
	}
	return fmt.Sprintf("%-7s  %s", n.Level, n)
}

// BeforeStore parses a source file in the message before storing the node.
func (n *LogEntry) BeforeStore(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	if err := n.ParseSourceFile(ctx); err != nil {
		log := appcontext.Get(ctx).Log
		log.WarningWithOwner(ctx, appCtx.SystemID, "failed to parse source file because %s", err.Error())
	}
}

// ParseSourceFile looks for a source file in the message and sets the source file fields.
func (n *LogEntry) ParseSourceFile(ctx context.Context) error {
	sourceFile, begin, end, err := util.MatchSourceFile(n.Message)
	if err == util.ErrNoMatch {
		return nil
	}
	if err != nil {
		return err
	}
	sourceParts := strings.Split(sourceFile, ":")
	absFilename, err := n.AbsolutePath(ctx, sourceParts[0])
	if err != nil {
		return err
	}
	if !util.FileExists(absFilename) {
		return nil
	}
	if util.IsDirectory(absFilename) {
		return nil
	}
	sourceParts[0] = absFilename
	sourceFile = strings.Join(sourceParts, ":")
	n.SourceFile = &sourceFile
	n.SourceFileBegin = &begin
	n.SourceFileEnd = &end
	return nil
}

// OwnerPath returns a filepath associated with the Owner if approriate.
func (n *LogEntry) OwnerPath(ctx context.Context) string {
	switch node := n.Owner(ctx).(type) {
	case *Project:
		return node.Path(ctx)
	}
	return ""
}

// AbsolutePath returns the absolute path of a filename.
// If the given filename is relative, it will be joined with the path of the entry's owner (if any).
func (n *LogEntry) AbsolutePath(ctx context.Context, filename string) (string, error) {
	if !filepath.IsAbs(filename) {
		full := filepath.Join(n.OwnerPath(ctx), filename)
		abs, err := filepath.Abs(full)
		if err != nil {
			return "", err
		}
		filename = abs
	}
	return filepath.Clean(filename), nil
}

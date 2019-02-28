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
	"io"
	"strconv"
	"time"
)

// DateFormat is the date format used throughout the app.
const DateFormat = "2006-01-02T15:04:05-0700"

// DateTime holds a date.
type DateTime time.Time

// UnmarshalGQL implements the graphql.Marshaler interface.
func (d *DateTime) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return ErrType
	}
	str, err := strconv.Unquote(str)
	if err != nil {
		return err
	}
	t, err := time.Parse(DateFormat, str)
	if err != nil {
		return err
	}
	*d = DateTime(t)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface.
func (d DateTime) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(strconv.Quote(time.Time(d).Format(DateFormat))))
}

// Hash holds a Git hash.
// TODO: change to bytes.
type Hash string

// UnmarshalGQL implements the graphql.Marshaler interface.
func (h *Hash) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return ErrType
	}
	*h = Hash(str)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface.
func (h Hash) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(strconv.Quote(string(h))))
}

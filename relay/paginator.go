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

package relay

import (
	"errors"
	"reflect"
)

// Pagination errors.
var (
	ErrPaginateFirst = errors.New("first cannot be negative")
	ErrPaginateLast  = errors.New("last cannot be negative")
)

// Paginator helps paginate lists for Relay.
//
// See: https://facebook.github.io/relay/graphql/connections.htm
type Paginator struct {
	// GetID must return the ID of a list value.
	GetID func(node interface{}) string

	// CreateEdge must create an edge given a cursor and a value.
	CreateEdge func(cursor string, node interface{}) interface{}

	// GetEdgeCursor must return the cursor assigned to an edge.
	GetEdgeCursor func(edge interface{}) string

	// EdgeType is the zero value of an edge.
	EdgeType interface{}
}

// PaginationConnection represents the result of a pagination.
type PaginationConnection struct {
	Edges    interface{}
	PageInfo PageInfo
}

// PageInfo contains fields related to pagination.
type PageInfo struct {
	HasPreviousPage bool   `json:"hasPreviousPage"`
	HasNextPage     bool   `json:"hasNextPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

// Paginate paginates a slice of nodes given query parameters.
func (p Paginator) Paginate(slice interface{}, after, before *string, first, last *int) (*PaginationConnection, error) {
	edgeSlice, hadMore := p.applyCursors(slice, after, before)
	edgeSliceV := reflect.ValueOf(edgeSlice)
	edgeSliceLen := edgeSliceV.Len()

	pageInfo := PageInfo{}

	if first != nil {
		firstValue := *first
		if firstValue < 0 {
			return nil, ErrPaginateFirst
		}
		if firstValue > edgeSliceLen {
			firstValue = edgeSliceLen
		}
		if firstValue < edgeSliceLen {
			pageInfo.HasNextPage = true
		} else if before != nil {
			pageInfo.HasNextPage = hadMore
		}
		edgeSliceV = edgeSliceV.Slice(0, firstValue)
		edgeSliceLen = edgeSliceV.Len()
	} else if before != nil {
		pageInfo.HasNextPage = hadMore
	}

	if last != nil {
		lastValue := *last
		if lastValue < 0 {
			return nil, ErrPaginateLast
		}
		if lastValue > edgeSliceLen {
			lastValue = edgeSliceLen
		}
		if lastValue < edgeSliceLen {
			pageInfo.HasPreviousPage = true
		} else if after != nil {
			pageInfo.HasPreviousPage = hadMore
		}
		end := edgeSliceLen - lastValue
		edgeSliceV = edgeSliceV.Slice(end, edgeSliceLen)
		edgeSliceLen = edgeSliceV.Len()
	} else if after != nil {
		pageInfo.HasPreviousPage = hadMore
	}

	if edgeSliceLen > 0 {
		pageInfo.StartCursor = p.GetEdgeCursor(edgeSliceV.Index(0).Interface())
		pageInfo.EndCursor = p.GetEdgeCursor(edgeSliceV.Index(edgeSliceLen - 1).Interface())
	}

	return &PaginationConnection{
		Edges:    edgeSliceV.Interface(),
		PageInfo: pageInfo,
	}, nil
}

func (p Paginator) applyCursors(slice interface{}, after, before *string) (interface{}, bool) {
	sliceV := reflect.ValueOf(slice)
	edgeT := reflect.TypeOf(p.EdgeType)
	edgeSliceT := reflect.SliceOf(edgeT)
	edgeSliceV := reflect.MakeSlice(edgeSliceT, 0, 10)
	hadMore := false

	if after != nil {
		index := p.indexOf(slice, *after)
		if index < 0 {
			return nil, false
		}
		hadMore = index > 0
		for i := index + 1; i < sliceV.Len(); i++ {
			node := sliceV.Index(i).Interface()
			edgeSliceV = reflect.Append(edgeSliceV, reflect.ValueOf(p.CreateEdge(
				p.GetID(node),
				node,
			)))
		}
		return edgeSliceV.Interface(), hadMore
	}

	if before != nil {
		index := p.indexOf(slice, *before)
		if index < 0 {
			return nil, false
		}
		hadMore = index < sliceV.Len()-1
		for i := 0; i < index-1; i++ {
			node := sliceV.Index(i).Interface()
			edgeSliceV = reflect.Append(edgeSliceV, reflect.ValueOf(p.CreateEdge(
				p.GetID(node),
				node,
			)))
		}
		return edgeSliceV.Interface(), hadMore
	}

	for i := 0; i < sliceV.Len(); i++ {
		node := sliceV.Index(i).Interface()
		edgeSliceV = reflect.Append(edgeSliceV, reflect.ValueOf(p.CreateEdge(
			p.GetID(node),
			node,
		)))
	}

	return edgeSliceV.Interface(), hadMore
}

func (p Paginator) indexOf(slice interface{}, id string) int {
	sliceV := reflect.ValueOf(slice)

	for i := 0; i < sliceV.Len(); i++ {
		elemV := sliceV.Index(i)
		elemID := p.GetID(elemV.Interface())

		if elemID == id {
			return i
		}
	}

	return -1
}

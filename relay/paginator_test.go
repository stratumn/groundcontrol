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
	"fmt"
	"reflect"
	"testing"
)

type TestModel struct {
	ID   string
	Name string
}

type TestEdge struct {
	Cursor string
	Node   TestModel
}

var testPaginator = Paginator{
	GetID: func(node interface{}) string {
		return node.(TestModel).ID
	},
	CreateEdge: func(cursor string, node interface{}) interface{} {
		return TestEdge{
			Cursor: cursor,
			Node:   node.(TestModel),
		}
	},
	GetEdgeCursor: func(edge interface{}) string {
		return edge.(TestEdge).Cursor
	},
	EdgeType: TestEdge{},
}

func TestPaginator_Paginate(t *testing.T) {
	var models []TestModel
	for i := 0; i < 10; i++ {
		models = append(models, TestModel{
			ID:   fmt.Sprint(i),
			Name: fmt.Sprint(i),
		})
	}

	var edges []TestEdge
	for i := 0; i < 10; i++ {
		edges = append(edges, TestEdge{
			Cursor: fmt.Sprint(i),
			Node:   models[i],
		})
	}

	five := 5

	type args struct {
		slice  interface{}
		after  *string
		before *string
		first  *int
		last   *int
	}
	tests := []struct {
		name    string
		args    args
		want    *PaginationConnection
		wantErr bool
	}{{
		"all",
		args{models, nil, nil, nil, nil},
		&PaginationConnection{
			Edges: edges,
			PageInfo: PageInfo{
				HasPreviousPage: false,
				HasNextPage:     false,
				StartCursor:     "0",
				EndCursor:       "9",
			},
		},
		false,
	}, {
		"after",
		args{models, &models[6].ID, nil, nil, nil},
		&PaginationConnection{
			Edges: edges[7:],
			PageInfo: PageInfo{
				HasPreviousPage: true,
				HasNextPage:     false,
				StartCursor:     "7",
				EndCursor:       "9",
			},
		},
		false,
	}, {
		"before",
		args{models, nil, &models[3].ID, nil, nil},
		&PaginationConnection{
			Edges: edges[:2],
			PageInfo: PageInfo{
				HasPreviousPage: false,
				HasNextPage:     true,
				StartCursor:     "0",
				EndCursor:       "1",
			},
		},
		false,
	}, {
		"first",
		args{models, nil, nil, &five, nil},
		&PaginationConnection{
			Edges: edges[:5],
			PageInfo: PageInfo{
				HasPreviousPage: false,
				HasNextPage:     true,
				StartCursor:     "0",
				EndCursor:       "4",
			},
		},
		false,
	}, {
		"last",
		args{models, nil, nil, nil, &five},
		&PaginationConnection{
			Edges: edges[5:],
			PageInfo: PageInfo{
				HasPreviousPage: true,
				HasNextPage:     false,
				StartCursor:     "5",
				EndCursor:       "9",
			},
		},
		false,
	}, {
		"after first",
		args{models, &models[2].ID, nil, &five, nil},
		&PaginationConnection{
			Edges: edges[3:8],
			PageInfo: PageInfo{
				HasPreviousPage: true,
				HasNextPage:     true,
				StartCursor:     "3",
				EndCursor:       "7",
			},
		},
		false,
	}, {
		"after last",
		args{models, &models[2].ID, nil, nil, &five},
		&PaginationConnection{
			Edges: edges[5:10],
			PageInfo: PageInfo{
				HasPreviousPage: true,
				HasNextPage:     false,
				StartCursor:     "5",
				EndCursor:       "9",
			},
		},
		false,
	}, {
		"before first",
		args{models, nil, &models[7].ID, &five, nil},
		&PaginationConnection{
			Edges: edges[0:5],
			PageInfo: PageInfo{
				HasPreviousPage: false,
				HasNextPage:     true,
				StartCursor:     "0",
				EndCursor:       "4",
			},
		},
		false,
	}, {
		"before last",
		args{models, nil, &models[7].ID, nil, &five},
		&PaginationConnection{
			Edges: edges[1:6],
			PageInfo: PageInfo{
				HasPreviousPage: true,
				HasNextPage:     true,
				StartCursor:     "1",
				EndCursor:       "5",
			},
		},
		false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testPaginator.Paginate(tt.args.slice, tt.args.after, tt.args.before, tt.args.first, tt.args.last)
			if (err != nil) != tt.wantErr {
				t.Errorf("Paginator.Paginate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Paginator.Paginate() = %v, want %v", got, tt.want)
			}
		})
	}
}

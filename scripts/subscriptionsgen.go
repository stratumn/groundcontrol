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

//+build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	subs := flag.String("s", "", "A comma separated list of subscriptions")
	filename := flag.String("o", "", "A filename to output the generated code to")
	flag.Parse()

	vars := struct {
		Added    []string
		Updated  []string
		Upserted []string
	}{}

	for _, sub := range strings.Split(*subs, ",") {
		switch {
		case strings.HasSuffix(sub, "Added"):
			vars.Added = append(vars.Added, strings.TrimSuffix(sub, "Added"))
		case strings.HasSuffix(sub, "Updated"):
			vars.Updated = append(vars.Updated, strings.TrimSuffix(sub, "Updated"))
		case strings.HasSuffix(sub, "Upserted"):
			vars.Upserted = append(vars.Upserted, strings.TrimSuffix(sub, "Upserted"))
		default:
			fmt.Println("subscriptions must be of the form TypeAdded, TypeUpdated, or TypeUpserted")
			os.Exit(1)
		}
	}

	w := os.Stdout

	if *filename != "" {
		f, err := os.OpenFile(*filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		checkError(err)

		defer f.Close()
		defer f.Sync()

		w = f
	}

	t, err := template.New("tmpl").Parse(tmpl)
	checkError(err)

	err = t.Execute(w, vars)
	checkError(err)
}

var tmpl = `// Code generated by github.com/stratumn/groundcontrol/scripts/subscriptionsgen.go, DO NOT EDIT.

package resolvers

import (
	"context"
	
	"github.com/stratumn/groundcontrol/models"
)

{{range $index, $type := .Added}}
func (r *subscriptionResolver) {{$type}}Added(ctx context.Context) (<-chan models.{{$type}}, error) {
	go func() {
		<-ctx.Done()
	}()

	ch := make(chan models.{{$type}}, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.{{$type}}Added, func(msg interface{}) {
		nodeID := msg.(string)
		select {
		case ch <- r.Nodes.MustLoad{{$type}}(nodeID):
		default:
		}
	})

	return ch, nil
}
{{end -}}

{{range $index, $type := .Updated}}
func (r *subscriptionResolver) {{$type}}Updated(ctx context.Context, id *string) (<-chan models.{{$type}}, error) {
	go func() {
		<-ctx.Done()
	}()

	ch := make(chan models.{{$type}}, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.{{$type}}Updated, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoad{{$type}}(nodeID):
		default:
		}
	})

	return ch, nil
}
{{end -}}

{{range $index, $type := .Upserted}}
func (r *subscriptionResolver) {{$type}}Upserted(ctx context.Context, id *string) (<-chan models.{{$type}}, error) {
	go func() {
		<-ctx.Done()
	}()

	ch := make(chan models.{{$type}}, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.{{$type}}Upserted, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoad{{$type}}(nodeID):
		default:
		}
	})

	return ch, nil
}
{{end -}}
`

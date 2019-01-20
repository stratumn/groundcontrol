//+build ignore

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"github.com/stratumn/groundcontrol"
)

func main() {
	err := vfsgen.Generate(groundcontrol.UI, vfsgen.Options{
		PackageName:  "groundcontrol",
		BuildTags:    "release",
		VariableName: "UI",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

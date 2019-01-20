//+build release

package main

import "github.com/stratumn/groundcontrol"

func init() {
	// When build for release, embed the UI.
	ui = groundcontrol.UI
}

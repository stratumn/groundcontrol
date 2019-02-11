package main

import (
	astilectron "github.com/asticode/go-astilectron"
)

func main() {
	// Initialize astilectron
	a, err := astilectron.New(astilectron.Options{
		AppName: "Ground Control",
		//AppIconDefaultPath: "assets", // If path is relative, it must be relative to the data directory
		//AppIconDarwinPath:  "assets", // Same here
		BaseDirectoryPath: "assets",
		SingleInstance:    true,
	})
	if err != nil {
		panic(err)
	}

	defer a.Close()

	// Start astilectron
	if err := a.Start(); err != nil {
		panic(err)
	}

	//gc := app.New(app.OptUI(embeddedUI), app.OptOpenBrowser(false))

	//go gc.Start(context.Background())

	// Create a new window
	w, err := a.NewWindow("http://127.0.0.1:3335", &astilectron.WindowOptions{
		Center:    astilectron.PtrBool(true),
		Width:     astilectron.PtrInt(900),
		Height:    astilectron.PtrInt(600),
		MinWidth:  astilectron.PtrInt(440),
		MinHeight: astilectron.PtrInt(480),
	})
	if err != nil {
		panic(err)
	}

	if err := w.Create(); err != nil {
		panic(err)
	}

	// Blocking pattern
	a.Wait()
}

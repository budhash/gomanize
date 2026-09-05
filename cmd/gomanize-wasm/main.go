//go:build js && wasm

// Command gomanize-wasm is the WebAssembly entry point for the browser demo in
// web/. It exposes one global function to JavaScript:
//
//	gomanizeTranslit(text, options) -> string
//
// where options is an object of booleans (see webdemo.FlagNames). The engine
// and all its embedded models are compiled in via go:embed, so the page needs
// no network calls beyond fetching the .wasm itself.
//
// Build with `make wasm`. This file only compiles for GOOS=js GOARCH=wasm, so
// it is invisible to the normal `go build ./...` / `make ci` toolchain.
package main

import (
	"syscall/js"

	"github.com/budhash/gomanize"
	"github.com/budhash/gomanize/webdemo"
)

func main() {
	g, err := gomanize.New("hindi")
	if err != nil {
		panic(err)
	}

	// The WASM runtime dispatches JS calls one at a time on a single event
	// loop, so mutating the shared instance's options per call is safe here.
	translit := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 {
			return ""
		}
		text := args[0].String()

		flags := make(map[string]bool, len(webdemo.FlagNames()))
		if len(args) >= 2 && args[1].Type() == js.TypeObject {
			for _, name := range webdemo.FlagNames() {
				flags[name] = args[1].Get(name).Truthy()
			}
		}

		g.SetOptions(webdemo.Options(flags))
		return g.Translit(text)
	})

	js.Global().Set("gomanizeTranslit", translit)
	js.Global().Get("console").Call("log", "gomanize wasm ready")

	// Keep the exported function callable for the lifetime of the page.
	select {}
}

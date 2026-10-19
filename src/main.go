//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
)

func main() {
	fmt.Println("LEXIO Go WebAssembly Engine initializing...")

	initUI()
	initDocumentManager()
	initSpeechEngine()
	initAuth()
	initDirectorStudio()

	fmt.Println("LEXIO Go WebAssembly Engine ready.")

	// Keep the Go runtime active to handle browser events
	select {}
}

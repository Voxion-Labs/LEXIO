package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type PageConfig struct {
	Title       string
	Description string
	Version     string
	Year        int
	Author      string
}

var defaultConfig = PageConfig{
	Title:       "LEXIO - Document Speech Studio",
	Description: "LEXIO - Client-side multilingual document speech studio powered by Go WebAssembly with synchronized phrase tracking.",
	Version:     "2.6.0",
	Year:        time.Now().Year(),
	Author:      "Voxion Labs",
}

func main() {
	serveFlag := flag.Bool("serve", false, "Start a local development server")
	portFlag := flag.String("port", "3000", "Port for local dev server")
	wasmOnly := flag.Bool("wasm", false, "Only build the WebAssembly binary")
	tmplOnly := flag.Bool("tmpl", false, "Only build the HTML from Go templates")
	flag.Parse()

	if *serveFlag {
		buildAll()
		startServer(*portFlag)
		return
	}

	if *wasmOnly {
		buildWasm()
		return
	}

	if *tmplOnly {
		buildTemplates()
		return
	}

	buildAll()
}

func buildAll() {
	fmt.Println("==================================================")
	fmt.Println(" LEXIO Go Build & Template Engine")
	fmt.Println("==================================================")

	buildTemplates()
	buildWasm()

	fmt.Println("Build complete. Ready for GitHub Pages deployment.")
}

func buildTemplates() {
	fmt.Println("--> Compiling Go templates from templates/*.tmpl...")

	tmplPattern := filepath.Join("templates", "*.tmpl")
	t, err := template.ParseGlob(tmplPattern)
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "layout", defaultConfig)
	if err != nil {
		log.Fatalf("Error executing layout template: %v", err)
	}

	// Compress HTML by cleaning up redundant blank lines
	rawHTML := buf.String()
	cleanHTML := compressHTML(rawHTML)

	err = os.WriteFile("index.html", []byte(cleanHTML), 0644)
	if err != nil {
		log.Fatalf("Error writing index.html: %v", err)
	}

	fmt.Printf("    Generated index.html (%d bytes)\n", len(cleanHTML))
}

func compressHTML(input string) string {
	// Collapse multiple consecutive empty lines to a single newline
	reBlankLines := regexp.MustCompile(`\n\s*\n\s*\n+`)
	cleaned := reBlankLines.ReplaceAllString(input, "\n\n")
	return strings.TrimSpace(cleaned) + "\n"
}

func buildWasm() {
	fmt.Println("--> Compiling Go WebAssembly engine (src/ -> wasm/app.wasm)...")

	// Ensure wasm output directory exists
	if err := os.MkdirAll("wasm", 0755); err != nil {
		log.Fatalf("Error creating wasm directory: %v", err)
	}

	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", "wasm/app.wasm", "./src")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error compiling WebAssembly: %v", err)
	}

	fi, err := os.Stat("wasm/app.wasm")
	if err != nil {
		log.Fatalf("Error reading wasm output file: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("    Compiled wasm/app.wasm: %.2f MB in %v\n", float64(fi.Size())/(1024*1024), duration.Round(time.Millisecond))
}

func startServer(port string) {
	fmt.Printf("\nServing LEXIO at http://localhost:%s\n", port)
	fmt.Println("MIME types configured for WebAssembly (application/wasm).")
	fmt.Println("Press Ctrl+C to stop.")

	fs := http.FileServer(http.Dir("."))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set proper MIME types for WASM
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.URL.Path, ".svg") {
			w.Header().Set("Content-Type", "image/svg+xml")
		}

		// Disable aggressive caching during local development
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		fs.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(":"+port, handler))
}

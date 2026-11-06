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
	stylesOnly := flag.Bool("styles", false, "Only build CSS from Go templates")
	flag.Parse()

	if *serveFlag {
		buildAll()
		startServer(*portFlag)
		return
	}

	if *wasmOnly {
		buildWasm()
		ensureWasmExec()
		return
	}

	if *tmplOnly {
		buildTemplates()
		return
	}

	if *stylesOnly {
		buildStyles()
		return
	}

	buildAll()
}

func buildAll() {
	fmt.Println("==================================================")
	fmt.Println(" LEXIO Go Build & Template Engine")
	fmt.Println("==================================================")

	buildTemplates()
	buildStyles()
	buildWasm()
	ensureWasmExec()

	fmt.Println("Build complete. Ready for GitHub Pages deployment.")
}

func buildTemplates() {
	fmt.Println("--> Compiling Go templates (templates/*.tmpl -> index.html)...")

	// Parse only HTML templates (exclude styles, components, and wasm_exec tmpls)
	htmlTmpls := []string{
		"templates/layout.tmpl",
		"templates/head.tmpl",
		"templates/loading.tmpl",
		"templates/landing.tmpl",
		"templates/modal.tmpl",
		"templates/app_header.tmpl",
		"templates/app_toolbar.tmpl",
		"templates/dashboard.tmpl",
		"templates/reader.tmpl",
		"templates/studio.tmpl",
		"templates/panels.tmpl",
		"templates/scripts.tmpl",
	}

	t, err := template.ParseFiles(htmlTmpls...)
	if err != nil {
		log.Fatalf("Error parsing HTML templates: %v", err)
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

func buildStyles() {
	fmt.Println("--> Compiling CSS stylesheets from Go templates...")

	// Render styles.css from templates/styles.tmpl
	stylesContent, err := os.ReadFile("templates/styles.tmpl")
	if err != nil {
		log.Fatalf("Error reading templates/styles.tmpl: %v", err)
	}
	err = os.WriteFile("styles.css", stylesContent, 0644)
	if err != nil {
		log.Fatalf("Error writing styles.css: %v", err)
	}
	fmt.Printf("    Generated styles.css (%d bytes)\n", len(stylesContent))

	// Render css/components.css from templates/components.tmpl
	componentsContent, err := os.ReadFile("templates/components.tmpl")
	if err != nil {
		log.Fatalf("Error reading templates/components.tmpl: %v", err)
	}
	if err := os.MkdirAll("css", 0755); err != nil {
		log.Fatalf("Error creating css directory: %v", err)
	}
	err = os.WriteFile(filepath.Join("css", "components.css"), componentsContent, 0644)
	if err != nil {
		log.Fatalf("Error writing css/components.css: %v", err)
	}
	fmt.Printf("    Generated css/components.css (%d bytes)\n", len(componentsContent))
}

func compressHTML(input string) string {
	reBlankLines := regexp.MustCompile(`\n\s*\n\s*\n+`)
	cleaned := reBlankLines.ReplaceAllString(input, "\n\n")
	return strings.TrimSpace(cleaned) + "\n"
}

func buildWasm() {
	fmt.Println("--> Compiling Go WebAssembly engine (src/ -> wasm/app.wasm)...")

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

func ensureWasmExec() {
	if err := os.MkdirAll("wasm", 0755); err != nil {
		log.Fatalf("Error creating wasm directory: %v", err)
	}

	target := filepath.Join("wasm", "wasm_exec.js")

	// First try to copy from local template
	if content, err := os.ReadFile("templates/wasm_exec.tmpl"); err == nil {
		if err := os.WriteFile(target, content, 0644); err == nil {
			fmt.Println("    Provisioned wasm/wasm_exec.js from templates/wasm_exec.tmpl")
			return
		}
	}

	// Fallback to Go installation GOROOT
	cmd := exec.Command("go", "env", "GOROOT")
	out, err := cmd.Output()
	if err == nil {
		goroot := strings.TrimSpace(string(out))
		candidate1 := filepath.Join(goroot, "lib", "wasm", "wasm_exec.js")
		candidate2 := filepath.Join(goroot, "misc", "wasm", "wasm_exec.js")

		var srcPath string
		if _, err := os.Stat(candidate1); err == nil {
			srcPath = candidate1
		} else if _, err := os.Stat(candidate2); err == nil {
			srcPath = candidate2
		}

		if srcPath != "" {
			if content, err := os.ReadFile(srcPath); err == nil {
				if err := os.WriteFile(target, content, 0644); err == nil {
					fmt.Printf("    Provisioned wasm/wasm_exec.js from %s\n", srcPath)
					return
				}
			}
		}
	}
}

func startServer(port string) {
	fmt.Printf("\nServing LEXIO at http://localhost:%s\n", port)
	fmt.Println("MIME types configured for WebAssembly (application/wasm).")
	fmt.Println("Press Ctrl+C to stop.")

	fs := http.FileServer(http.Dir("."))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.URL.Path, ".svg") {
			w.Header().Set("Content-Type", "image/svg+xml")
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		fs.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(":"+port, handler))
}

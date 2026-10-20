//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"syscall/js"
	"unicode"
)

type DocSection struct {
	ID        string
	Label     string
	Kind      string
	Text      string
	WordCount int
	StartIdx  int
}

var (
	currentRawText     = ""
	currentCleanedText = ""
	currentSections    []DocSection
	currentDocWords    []string
	activeWordElement  js.Value
)

func initDocumentManager() {
	doc := js.Global().Get("document")

	inputArea := doc.Call("getElementById", "inputText")
	if !inputArea.IsNull() {
		inputArea.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			raw := inputArea.Get("value").String()
			ProcessAndRenderDocument(raw, false)
			return nil
		}))
	}

	// Demo sample button
	demoBtn := doc.Call("getElementById", "demoBtn")
	if !demoBtn.IsNull() {
		demoBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			loadSampleDocument()
			return nil
		}))
	}

	// Dashboard sample button
	dashSampleBtn := doc.Call("getElementById", "dashLoadSampleBtn")
	if !dashSampleBtn.IsNull() {
		dashSampleBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			loadSampleDocument()
			return nil
		}))
	}

	// Clear button
	clearBtn := doc.Call("getElementById", "clearBtn")
	if !clearBtn.IsNull() {
		clearBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ClearDocument()
			return nil
		}))
	}

	// Setup File Import Drop Zone & File Input
	initFileUploader()
}

func loadSampleDocument() {
	sample := `# Document Reading & Narration

1. Introduction
Welcome to LEXIO, a client-side document speech studio designed for professionals and learners. It reads documents by natural phrases rather than word by word, giving you smooth and clear delivery.

2. Multilingual & Regional Accents
Whether you are reviewing research papers in English, contracts in Spanish, reports in Italian, or guides in Hindi and Arabic, you can customize the voice, pitch, and listening speed to match your focus.

3. Zero Backend Privacy
All file processing, speech synthesis, and document structure analysis happen inside your browser. No files are uploaded to remote servers, ensuring complete confidentiality for your work.`

	doc := js.Global().Get("document")
	inputArea := doc.Call("getElementById", "inputText")
	if !inputArea.IsNull() {
		inputArea.Set("value", sample)
	}
	ProcessAndRenderDocument(sample, true)
	ShowToast("Sample document loaded", "success")
}

func ClearDocument() {
	StopReading()
	currentRawText = ""
	currentCleanedText = ""
	currentSections = nil
	currentDocWords = nil

	doc := js.Global().Get("document")
	inputArea := doc.Call("getElementById", "inputText")
	if !inputArea.IsNull() {
		inputArea.Set("value", "")
	}

	renderEmptyDocumentSurface()
	updateDocumentMetrics(0, 0, 0, 0, "en")
	UpdateDocumentReadability("")
	ShowToast("Workspace cleared", "info")
}

func ProcessAndRenderDocument(text string, isClean bool) {
	currentRawText = text

	cleaned := text
	if !isClean {
		cleaned = cleanText(text)
	}
	currentCleanedText = cleaned

	words := strings.Fields(cleaned)
	currentDocWords = words

	sections := buildSections(cleaned)
	currentSections = sections

	renderDocumentSurface(sections)
	renderDocumentOutline(sections)

	// Language detection
	detectedLang := detectLanguage(cleaned)

	// Update Metrics
	wordCount := len(words)
	charCount := len([]rune(cleaned))
	sectionCount := len(sections)
	readMinutes := float64(wordCount) / 150.0
	readTimeSeconds := int(readMinutes * 60)

	updateDocumentMetrics(wordCount, charCount, sectionCount, readTimeSeconds, detectedLang)
	UpdateDocumentReadability(cleaned)
}

func cleanText(raw string) string {
	res := strings.ReplaceAll(raw, "\r\n", "\n")
	res = strings.ReplaceAll(res, "\r", "\n")

	// Collapse multiple spaces into single space
	spaceRe := regexp.MustCompile(`[ \t]+`)
	res = spaceRe.ReplaceAllString(res, " ")

	// Collapse 3+ newlines to 2
	nlRe := regexp.MustCompile(`\n{3,}`)
	res = nlRe.ReplaceAllString(res, "\n\n")

	return strings.TrimSpace(res)
}

func buildSections(text string) []DocSection {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	paragraphs := strings.Split(text, "\n\n")
	var sections []DocSection
	wordIdx := 0

	for i, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		pWords := strings.Fields(p)
		secID := fmt.Sprintf("sec-%d", i+1)
		label := fmt.Sprintf("Section %d", i+1)
		kind := "Section"

		// Detect markdown heading or custom label
		if strings.HasPrefix(p, "#") {
			lines := strings.Split(p, "\n")
			firstLine := strings.TrimLeft(lines[0], "# \t")
			if firstLine != "" {
				label = firstLine
				kind = "Heading"
			}
		} else if len(pWords) > 0 && len(pWords) <= 8 {
			label = p
			kind = "Topic"
		}

		sections = append(sections, DocSection{
			ID:        secID,
			Label:     label,
			Kind:      kind,
			Text:      p,
			WordCount: len(pWords),
			StartIdx:  wordIdx,
		})

		wordIdx += len(pWords)
	}

	return sections
}

func renderDocumentSurface(sections []DocSection) {
	doc := js.Global().Get("document")
	surface := doc.Call("getElementById", "documentSurface")
	if surface.IsNull() {
		return
	}

	if len(sections) == 0 {
		renderEmptyDocumentSurface()
		return
	}

	var sb strings.Builder
	sb.WriteString("<div class=\"document-flow\">")

	globalWordOffset := 0

	for _, sec := range sections {
		sb.WriteString(fmt.Sprintf("<section class=\"doc-section\" data-section-id=\"%s\">", sec.ID))
		sb.WriteString("<header class=\"doc-section-header\">")
		sb.WriteString(fmt.Sprintf("<span class=\"doc-section-kind\">%s</span>", html.EscapeString(sec.Kind)))
		sb.WriteString(fmt.Sprintf("<h4>%s</h4>", html.EscapeString(sec.Label)))
		sb.WriteString(fmt.Sprintf("<span class=\"doc-section-meta\">%d words</span>", sec.WordCount))
		sb.WriteString("</header>")

		sb.WriteString("<div class=\"doc-section-body\">")
		// Tokenize section text
		words := strings.Fields(sec.Text)
		for _, w := range words {
			sb.WriteString(fmt.Sprintf("<button type=\"button\" class=\"doc-word\" data-word-index=\"%d\">%s</button> ", globalWordOffset, html.EscapeString(w)))
			globalWordOffset++
		}
		sb.WriteString("</div>")
		sb.WriteString("</section>")
	}

	sb.WriteString("</div>")
	surface.Set("innerHTML", sb.String())

	// Attach click listeners to all words
	wordButtons := surface.Call("querySelectorAll", ".doc-word")
	count := wordButtons.Get("length").Int()
	for i := 0; i < count; i++ {
		btn := wordButtons.Index(i)
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			idxAttr := this.Call("getAttribute", "data-word-index").String()
			var clickedIdx int
			if n, _ := fmt.Sscanf(idxAttr, "%d", &clickedIdx); n == 1 {
				StartReading(clickedIdx)
			}
			return nil
		}))
	}
}

func renderDocumentOutline(sections []DocSection) {
	doc := js.Global().Get("document")
	outlinePanel := doc.Call("getElementById", "outlinePanel")
	if outlinePanel.IsNull() {
		return
	}

	if len(sections) == 0 {
		outlinePanel.Set("innerHTML", "<p class=\"empty-copy\">Document outline appears here once text is loaded.</p>")
		return
	}

	var sb strings.Builder
	for _, sec := range sections {
		sb.WriteString(fmt.Sprintf("<button type=\"button\" class=\"outline-item\" data-sec-target=\"%s\">", sec.ID))
		sb.WriteString(fmt.Sprintf("<span class=\"outline-label\">%s</span>", html.EscapeString(sec.Label)))
		sb.WriteString(fmt.Sprintf("<span class=\"outline-meta\">%d words</span>", sec.WordCount))
		sb.WriteString("</button>")
	}
	outlinePanel.Set("innerHTML", sb.String())

	// Outline jump click listener
	items := outlinePanel.Call("querySelectorAll", ".outline-item")
	for i := 0; i < items.Get("length").Int(); i++ {
		item := items.Index(i)
		item.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			secTarget := this.Call("getAttribute", "data-sec-target").String()
			targetEl := doc.Call("querySelector", fmt.Sprintf("[data-section-id=\"%s\"]", secTarget))
			if !targetEl.IsNull() {
				opts := js.Global().Get("Object").New()
				opts.Set("behavior", "smooth")
				opts.Set("block", "center")
				targetEl.Call("scrollIntoView", opts)
			}
			return nil
		}))
	}
}

func renderEmptyDocumentSurface() {
	doc := js.Global().Get("document")
	surface := doc.Call("getElementById", "documentSurface")
	if surface.IsNull() {
		return
	}
	surface.Set("innerHTML", `
        <div class="empty-state">
            <i class="fas fa-file-lines"></i>
            <h4>Import a file or paste content</h4>
            <p>LEXIO will render your sections here and highlight spoken words in real time.</p>
        </div>
    `)
}

func SetActiveWord(wordIdx int) {
	ClearActiveWord()

	doc := js.Global().Get("document")
	surface := doc.Call("getElementById", "documentSurface")
	if surface.IsNull() {
		return
	}

	target := surface.Call("querySelector", fmt.Sprintf("[data-word-index=\"%d\"]", wordIdx))
	if !target.IsNull() && !target.IsUndefined() {
		target.Get("classList").Call("add", "active-word")
		activeWordElement = target

		autoScrollCheck := doc.Call("getElementById", "autoScrollHighlight")
		shouldScroll := true
		if !autoScrollCheck.IsNull() {
			shouldScroll = autoScrollCheck.Get("checked").Bool()
		}

		if shouldScroll {
			opts := js.Global().Get("Object").New()
			opts.Set("behavior", "smooth")
			opts.Set("block", "center")
			opts.Set("inline", "nearest")
			target.Call("scrollIntoView", opts)
		}
	}
}

func ClearActiveWord() {
	if !activeWordElement.IsNull() && !activeWordElement.IsUndefined() {
		activeWordElement.Get("classList").Call("remove", "active-word")
		activeWordElement = js.Null()
	}
}

func updateDocumentMetrics(words, chars, sections, readTimeSec int, lang string) {
	doc := js.Global().Get("document")

	timeStr := fmt.Sprintf("%ds", readTimeSec)
	if readTimeSec >= 60 {
		timeStr = fmt.Sprintf("%dm %ds", readTimeSec/60, readTimeSec%60)
	}

	setElText(doc, "wordCount", fmt.Sprintf("%d", words))
	setElText(doc, "charCount", fmt.Sprintf("%d", chars))
	setElText(doc, "sectionCount", fmt.Sprintf("%d", sections))
	setElText(doc, "timeEstimate", timeStr)

	setElText(doc, "charCounter", fmt.Sprintf("%d chars", chars))
	setElText(doc, "wordCountChip", fmt.Sprintf("%d words", words))

	langDisplay := "English"
	switch lang {
	case "hi":
		langDisplay = "Hindi"
	case "es":
		langDisplay = "Spanish"
	case "it":
		langDisplay = "Italian"
	case "ar":
		langDisplay = "Arabic"
	}
	setElText(doc, "detectedLanguage", langDisplay)

	// Update reading stats overview
	setElText(doc, "docWordsStat", fmt.Sprintf("%d", words))
	setElText(doc, "docSectionsStat", fmt.Sprintf("%d", sections))
	setElText(doc, "docPaceStat", "150 wpm")
}

func setElText(doc js.Value, id, val string) {
	el := doc.Call("getElementById", id)
	if !el.IsNull() {
		el.Set("textContent", val)
	}
}

func detectLanguage(text string) string {
	sample := text
	if len(sample) > 1000 {
		sample = sample[:1000]
	}

	hindiCount := 0
	arabicCount := 0
	spanishCount := 0

	for _, r := range sample {
		if unicode.In(r, unicode.Devanagari) {
			hindiCount++
		} else if unicode.In(r, unicode.Arabic) {
			arabicCount++
		} else if r == 'ñ' || r == 'á' || r == 'é' || r == 'í' || r == 'ó' || r == 'ú' || r == '¿' || r == '¡' {
			spanishCount++
		}
	}

	if hindiCount > 5 {
		return "hi"
	}
	if arabicCount > 5 {
		return "ar"
	}
	if spanishCount > 3 {
		return "es"
	}
	return "en"
}

func GetCurrentDocumentWords() []string {
	return currentDocWords
}

func initFileUploader() {
	doc := js.Global().Get("document")
	uploadBtn := doc.Call("getElementById", "uploadBtn")

	// Create hidden file input
	fileInput := doc.Call("createElement", "input")
	fileInput.Set("type", "file")
	fileInput.Set("id", "hiddenFileInput")
	fileInput.Set("accept", ".txt,.md,.json,.csv,.rtf,.html,.pdf,.docx,.xlsx")
	fileInput.Get("style").Set("display", "none")
	doc.Get("body").Call("appendChild", fileInput)

	if !uploadBtn.IsNull() {
		uploadBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			fileInput.Call("click")
			return nil
		}))
	}

	fileInput.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		files := fileInput.Get("files")
		if files.Get("length").Int() > 0 {
			file := files.Index(0)
			handleImportedFile(file)
		}
		fileInput.Set("value", "")
		return nil
	}))
}

func handleImportedFile(file js.Value) {
	fileName := file.Get("name").String()
	lowerName := strings.ToLower(fileName)

	doc := js.Global().Get("document")
	fileBadge := doc.Call("getElementById", "fileBadge")
	if !fileBadge.IsNull() {
		fileBadge.Set("textContent", fileName)
	}

	// Plain text types: .txt, .md, .csv, .json, .rtf, .html
	if strings.HasSuffix(lowerName, ".txt") || strings.HasSuffix(lowerName, ".md") ||
		strings.HasSuffix(lowerName, ".csv") || strings.HasSuffix(lowerName, ".json") ||
		strings.HasSuffix(lowerName, ".rtf") || strings.HasSuffix(lowerName, ".html") {

		reader := js.Global().Get("FileReader").New()
		reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			content := args[0].Get("target").Get("result").String()
			inputArea := doc.Call("getElementById", "inputText")
			if !inputArea.IsNull() {
				inputArea.Set("value", content)
			}
			ProcessAndRenderDocument(content, false)
			ShowToast("Imported "+fileName, "success")
			return nil
		}))
		reader.Call("readAsText", file)
		return
	}

	// PDF parsing via pdf.js
	if strings.HasSuffix(lowerName, ".pdf") {
		reader := js.Global().Get("FileReader").New()
		reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			typedArray := js.Global().Get("Uint8Array").New(args[0].Get("target").Get("result"))
			pdfjs := js.Global().Get("pdfjsLib")
			if pdfjs.IsNull() || pdfjs.IsUndefined() {
				ShowToast("PDF library is loading, please try again", "error")
				return nil
			}

			loadTask := pdfjs.Call("getDocument", typedArray)
			loadTask.Get("promise").Call("then", js.FuncOf(func(pThis js.Value, pArgs []js.Value) interface{} {
				pdfDoc := pArgs[0]
				numPages := pdfDoc.Get("numPages").Int()
				var fullText strings.Builder

				var extractPage func(int)
				extractPage = func(pageNum int) {
					if pageNum > numPages {
						extracted := fullText.String()
						inputArea := doc.Call("getElementById", "inputText")
						if !inputArea.IsNull() {
							inputArea.Set("value", extracted)
						}
						ProcessAndRenderDocument(extracted, false)
						ShowToast(fmt.Sprintf("Imported %d pages from %s", numPages, fileName), "success")
						return
					}
					pdfDoc.Call("getPage", pageNum).Call("then", js.FuncOf(func(pgThis js.Value, pgArgs []js.Value) interface{} {
						page := pgArgs[0]
						page.Call("getTextContent").Call("then", js.FuncOf(func(tcThis js.Value, tcArgs []js.Value) interface{} {
							items := tcArgs[0].Get("items")
							fullText.WriteString(fmt.Sprintf("\n\n--- Page %d ---\n", pageNum))
							for i := 0; i < items.Get("length").Int(); i++ {
								str := items.Index(i).Get("str").String()
								fullText.WriteString(str + " ")
							}
							extractPage(pageNum + 1)
							return nil
						}))
						return nil
					}))
				}
				extractPage(1)
				return nil
			}))
			return nil
		}))
		reader.Call("readAsArrayBuffer", file)
		return
	}

	// DOCX parsing via mammoth
	if strings.HasSuffix(lowerName, ".docx") {
		reader := js.Global().Get("FileReader").New()
		reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			mammoth := js.Global().Get("mammoth")
			if mammoth.IsNull() || mammoth.IsUndefined() {
				ShowToast("DOCX parser is loading, please try again", "error")
				return nil
			}
			opts := js.Global().Get("Object").New()
			opts.Set("arrayBuffer", args[0].Get("target").Get("result"))
			mammoth.Call("extractRawText", opts).Call("then", js.FuncOf(func(mThis js.Value, mArgs []js.Value) interface{} {
				docxText := mArgs[0].Get("value").String()
				inputArea := doc.Call("getElementById", "inputText")
				if !inputArea.IsNull() {
					inputArea.Set("value", docxText)
				}
				ProcessAndRenderDocument(docxText, false)
				ShowToast("Imported "+fileName, "success")
				return nil
			}))
			return nil
		}))
		reader.Call("readAsArrayBuffer", file)
		return
	}

	ShowToast("Unsupported file type: "+fileName, "error")
}

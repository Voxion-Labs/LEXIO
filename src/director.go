//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"regexp"
	"strings"
	"syscall/js"
)

type PronunciationRule struct {
	Term   string `json:"term"`
	Spoken string `json:"spoken"`
}

var (
	lexiconRules         []PronunciationRule
	teleprompterFontSize = 28
	teleprompterTheme    = "dark" // "dark", "light", "contrast"
)

func initDirectorStudio() {
	doc := js.Global().Get("document")

	// Load saved lexicon from localStorage
	loadLexiconRules()
	renderLexiconTable()

	// Lexicon add rule button
	addRuleBtn := doc.Call("getElementById", "addLexiconBtn")
	if !addRuleBtn.IsNull() {
		addRuleBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			termInput := doc.Call("getElementById", "lexiconTerm")
			spokenInput := doc.Call("getElementById", "lexiconSpoken")
			if !termInput.IsNull() && !spokenInput.IsNull() {
				term := strings.TrimSpace(termInput.Get("value").String())
				spoken := strings.TrimSpace(spokenInput.Get("value").String())
				if term != "" && spoken != "" {
					AddPronunciationRule(term, spoken)
					termInput.Set("value", "")
					spokenInput.Set("value", "")
					ShowToast(fmt.Sprintf("Added rule: %s → %s", term, spoken), "success")
				} else {
					ShowToast("Please enter both term and spoken form", "error")
				}
			}
			return nil
		}))
	}

	// Lexicon preset button
	presetBtn := doc.Call("getElementById", "loadLexiconPresetsBtn")
	if !presetBtn.IsNull() {
		presetBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			loadDefaultLexiconPresets()
			ShowToast("Standard tech & pronunciation presets loaded", "success")
			return nil
		}))
	}

	// Teleprompter font size slider
	fontSlider := doc.Call("getElementById", "teleprompterFontSlider")
	if !fontSlider.IsNull() {
		fontSlider.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			val := fontSlider.Get("value").String()
			var size int
			if n, _ := fmt.Sscanf(val, "%d", &size); n == 1 && size >= 18 {
				teleprompterFontSize = size
				applyTeleprompterStyles()
			}
			return nil
		}))
	}

	// Teleprompter theme buttons
	darkThemeBtn := doc.Call("getElementById", "teleprompterThemeDark")
	if !darkThemeBtn.IsNull() {
		darkThemeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			setTeleprompterTheme("dark")
			return nil
		}))
	}

	lightThemeBtn := doc.Call("getElementById", "teleprompterThemeLight")
	if !lightThemeBtn.IsNull() {
		lightThemeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			setTeleprompterTheme("light")
			return nil
		}))
	}

	contrastThemeBtn := doc.Call("getElementById", "teleprompterThemeContrast")
	if !contrastThemeBtn.IsNull() {
		contrastThemeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			setTeleprompterTheme("contrast")
			return nil
		}))
	}

	// Teleprompter Play/Pause/Resume/Stop triggers
	teleStartBtn := doc.Call("getElementById", "teleStartBtn")
	if !teleStartBtn.IsNull() {
		teleStartBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			StartReading(0)
			return nil
		}))
	}

	telePauseBtn := doc.Call("getElementById", "telePauseBtn")
	if !telePauseBtn.IsNull() {
		telePauseBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			PauseReading()
			return nil
		}))
	}

	teleResumeBtn := doc.Call("getElementById", "teleResumeBtn")
	if !teleResumeBtn.IsNull() {
		teleResumeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ResumeReading()
			return nil
		}))
	}

	teleStopBtn := doc.Call("getElementById", "teleStopBtn")
	if !teleStopBtn.IsNull() {
		teleStopBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			StopReading()
			return nil
		}))
	}

	// Subtitle & Caption Exporters (SRT & WebVTT)
	exportSrtBtn := doc.Call("getElementById", "exportSrtBtn")
	if !exportSrtBtn.IsNull() {
		exportSrtBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ExportSubtitles("srt")
			return nil
		}))
	}

	exportVttBtn := doc.Call("getElementById", "exportVttBtn")
	if !exportVttBtn.IsNull() {
		exportVttBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ExportSubtitles("vtt")
			return nil
		}))
	}

	applyTeleprompterStyles()
}

func UpdateTeleprompter(activeIdx int) {
	doc := js.Global().Get("document")
	stage := doc.Call("getElementById", "teleprompterContent")
	if stage.IsNull() {
		return
	}

	words := GetCurrentDocumentWords()
	if len(words) == 0 || activeIdx < 0 || activeIdx >= len(words) {
		stage.Set("innerHTML", "<p class=\"teleprompter-empty\">Load a document and press Start to run the Teleprompter Deck.</p>")
		return
	}

	// Window of words around active word
	windowStart := activeIdx - 6
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := activeIdx + 12
	if windowEnd > len(words) {
		windowEnd = len(words)
	}

	var sb strings.Builder
	for i := windowStart; i < windowEnd; i++ {
		w := html.EscapeString(words[i])
		if i == activeIdx {
			sb.WriteString(fmt.Sprintf("<span class=\"teleprompter-word-active\">%s</span> ", w))
		} else {
			sb.WriteString(fmt.Sprintf("<span class=\"teleprompter-word\">%s</span> ", w))
		}
	}

	stage.Set("innerHTML", sb.String())

	meta := doc.Call("getElementById", "teleprompterMeta")
	if !meta.IsNull() {
		meta.Set("textContent", fmt.Sprintf("Word %d of %d", activeIdx+1, len(words)))
	}
}

func setTeleprompterTheme(theme string) {
	teleprompterTheme = theme
	doc := js.Global().Get("document")
	stage := doc.Call("getElementById", "teleprompterStage")
	if stage.IsNull() {
		return
	}

	stage.Get("classList").Call("remove", "theme-dark", "theme-light", "theme-contrast")
	stage.Get("classList").Call("add", "theme-"+theme)
}

func applyTeleprompterStyles() {
	doc := js.Global().Get("document")
	content := doc.Call("getElementById", "teleprompterContent")
	if !content.IsNull() {
		content.Get("style").Set("fontSize", fmt.Sprintf("%dpx", teleprompterFontSize))
	}
}

func ApplyPronunciationLexicon(text string) string {
	res := text
	for _, rule := range lexiconRules {
		if rule.Term == "" || rule.Spoken == "" {
			continue
		}
		pattern := `(?i)\b` + regexp.QuoteMeta(rule.Term) + `\b`
		re, err := regexp.Compile(pattern)
		if err == nil {
			res = re.ReplaceAllString(res, rule.Spoken)
		}
	}
	return res
}

func AddPronunciationRule(term, spoken string) {
	for i, r := range lexiconRules {
		if strings.EqualFold(r.Term, term) {
			lexiconRules[i].Spoken = spoken
			saveLexiconRules()
			renderLexiconTable()
			return
		}
	}

	lexiconRules = append(lexiconRules, PronunciationRule{
		Term:   term,
		Spoken: spoken,
	})
	saveLexiconRules()
	renderLexiconTable()
}

func DeletePronunciationRule(term string) {
	var updated []PronunciationRule
	for _, r := range lexiconRules {
		if !strings.EqualFold(r.Term, term) {
			updated = append(updated, r)
		}
	}
	lexiconRules = updated
	saveLexiconRules()
	renderLexiconTable()
}

func saveLexiconRules() {
	storage := js.Global().Get("localStorage")
	if !storage.IsNull() && !storage.IsUndefined() {
		data, _ := json.Marshal(lexiconRules)
		storage.Call("setItem", "lexio_pronunciation_lexicon", string(data))
	}
}

func loadLexiconRules() {
	storage := js.Global().Get("localStorage")
	if storage.IsNull() || storage.IsUndefined() {
		return
	}
	raw := storage.Call("getItem", "lexio_pronunciation_lexicon")
	if !raw.IsNull() && !raw.IsUndefined() && raw.String() != "" {
		json.Unmarshal([]byte(raw.String()), &lexiconRules)
	}

	if len(lexiconRules) == 0 {
		loadDefaultLexiconPresets()
	}
}

func loadDefaultLexiconPresets() {
	lexiconRules = []PronunciationRule{
		{Term: "WASM", Spoken: "Waz-um"},
		{Term: "SQL", Spoken: "Sequel"},
		{Term: "API", Spoken: "A-P-I"},
		{Term: "GUI", Spoken: "Goo-ey"},
		{Term: "VoIP", Spoken: "Voyp"},
		{Term: "LaTeX", Spoken: "Lay-tek"},
		{Term: "etc.", Spoken: "et cetera"},
		{Term: "e.g.", Spoken: "for example"},
		{Term: "i.e.", Spoken: "that is"},
	}
	saveLexiconRules()
	renderLexiconTable()
}

func renderLexiconTable() {
	doc := js.Global().Get("document")
	tbody := doc.Call("getElementById", "lexiconTableBody")
	if tbody.IsNull() {
		return
	}

	if len(lexiconRules) == 0 {
		tbody.Set("innerHTML", "<tr><td colspan=\"3\" style=\"text-align:center;color:var(--text-muted);padding:14px;\">No custom pronunciation rules yet. Add one above or load standard presets!</td></tr>")
		return
	}

	var sb strings.Builder
	for _, r := range lexiconRules {
		sb.WriteString(fmt.Sprintf(`
            <tr>
                <td><strong>%s</strong></td>
                <td><span class="badge badge-primary">%s</span></td>
                <td style="text-align: right;">
                    <button type="button" class="btn-icon delete-lex-btn" data-term="%s" title="Delete rule" style="width:30px;height:30px;">
                        <i class="fas fa-trash" style="font-size:0.8rem;color:var(--error);"></i>
                    </button>
                </td>
            </tr>
        `, html.EscapeString(r.Term), html.EscapeString(r.Spoken), html.EscapeString(r.Term)))
	}
	tbody.Set("innerHTML", sb.String())

	delButtons := tbody.Call("querySelectorAll", ".delete-lex-btn")
	for i := 0; i < delButtons.Get("length").Int(); i++ {
		btn := delButtons.Index(i)
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			term := this.Call("getAttribute", "data-term").String()
			DeletePronunciationRule(term)
			ShowToast("Removed pronunciation rule", "info")
			return nil
		}))
	}
}

// ============================================================================
// LINGUISTIC READABILITY & CADENCE TELEMETRY
// ============================================================================

func countSyllablesInWord(word string) int {
	w := strings.ToLower(strings.TrimSpace(word))
	// Strip non-alpha
	var clean []rune
	for _, r := range w {
		if (r >= 'a' && r <= 'z') || r == '\'' {
			clean = append(clean, r)
		}
	}
	w = string(clean)
	if len(w) <= 3 {
		return 1
	}

	// Remove trailing 'e', 'es', 'ed' unless preceded by 'le'
	if strings.HasSuffix(w, "e") && !strings.HasSuffix(w, "le") {
		w = w[:len(w)-1]
	}

	vowels := "aeiouy"
	count := 0
	inVowelGroup := false
	for _, ch := range w {
		isVowel := strings.ContainsRune(vowels, ch)
		if isVowel && !inVowelGroup {
			count++
			inVowelGroup = true
		} else if !isVowel {
			inVowelGroup = false
		}
	}

	if count == 0 {
		return 1
	}
	return count
}

func UpdateDocumentReadability(text string) {
	doc := js.Global().Get("document")
	if strings.TrimSpace(text) == "" {
		setElText(doc, "fleschScore", "--")
		setElText(doc, "fleschBadge", "No Document")
		setElText(doc, "gunningFogScore", "--")
		setElText(doc, "syllableCount", "0 syllables")
		setElText(doc, "complexWordCount", "0 complex")
		setElText(doc, "pacePresTime", "0s")
		setElText(doc, "paceNormalTime", "0s")
		setElText(doc, "paceFastTime", "0s")
		return
	}

	words := strings.Fields(text)
	totalWords := len(words)
	if totalWords == 0 {
		return
	}

	// Count sentences
	sentenceRe := regexp.MustCompile(`[.!?]+(?:\s+|\n+|$)`)
	matches := sentenceRe.FindAllStringIndex(text, -1)
	totalSentences := len(matches)
	if totalSentences == 0 {
		totalSentences = 1
	}

	totalSyllables := 0
	complexWords := 0
	for _, w := range words {
		syl := countSyllablesInWord(w)
		totalSyllables += syl
		if syl >= 3 {
			complexWords++
		}
	}

	asl := float64(totalWords) / float64(totalSentences)
	asw := float64(totalSyllables) / float64(totalWords)

	// Flesch Reading Ease Formula
	flesch := 206.835 - (1.015 * asl) - (84.6 * asw)
	if flesch < 0 {
		flesch = 0
	} else if flesch > 100 {
		flesch = 100
	}

	rating := "Standard"
	if flesch >= 80 {
		rating = "Easy / Conversational"
	} else if flesch >= 60 {
		rating = "Standard / Clear"
	} else if flesch >= 45 {
		rating = "Professional / Technical"
	} else {
		rating = "Dense / Academic"
	}

	// Gunning Fog Index
	fog := 0.4 * (asl + 100.0*(float64(complexWords)/float64(totalWords)))
	fogGrade := fmt.Sprintf("Grade %.1f", fog)
	if fog >= 16 {
		fogGrade = "College / Graduate"
	}

	// Cadence times
	calcTime := func(wpm float64) string {
		sec := int(math.Round((float64(totalWords) / wpm) * 60.0))
		if sec < 60 {
			return fmt.Sprintf("%ds", sec)
		}
		return fmt.Sprintf("%dm %ds", sec/60, sec%60)
	}

	setElText(doc, "fleschScore", fmt.Sprintf("%.1f / 100", flesch))
	setElText(doc, "fleschBadge", rating)
	setElText(doc, "gunningFogScore", fogGrade)
	setElText(doc, "syllableCount", fmt.Sprintf("%d syllables", totalSyllables))
	setElText(doc, "complexWordCount", fmt.Sprintf("%d complex (3+ syl)", complexWords))
	setElText(doc, "pacePresTime", calcTime(125.0))
	setElText(doc, "paceNormalTime", calcTime(150.0))
	setElText(doc, "paceFastTime", calcTime(180.0))
}

// ============================================================================
// SUBTITLE & CLOSED CAPTION EXPORTERS (.SRT & .VTT)
// ============================================================================

func ExportSubtitles(format string) {
	if strings.TrimSpace(currentCleanedText) == "" {
		ShowToast("Please load or paste a document first", "error")
		return
	}

	// Split text into distinct sentences
	splitRe := regexp.MustCompile(`(?m)([^.!?\n]+[.!?]+(?:\s+|\n+|$)|[^\n]+)`)
	rawSentences := splitRe.FindAllString(currentCleanedText, -1)

	var sentences []string
	for _, s := range rawSentences {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			sentences = append(sentences, trimmed)
		}
	}

	if len(sentences) == 0 {
		ShowToast("No sentences found to export", "error")
		return
	}

	wpm := 150.0 // Base pace
	wordsPerSec := wpm / 60.0

	currentTimeSec := 0.0
	var sb strings.Builder

	if format == "vtt" {
		sb.WriteString("WEBVTT - Generated by LEXIO Document Speech Studio\n\n")
	}

	for idx, s := range sentences {
		sWords := len(strings.Fields(s))
		dur := float64(sWords) / wordsPerSec
		if dur < 1.4 {
			dur = 1.4
		}

		startTime := currentTimeSec
		endTime := currentTimeSec + dur

		startFormatted := formatTimestamp(startTime, format)
		endFormatted := formatTimestamp(endTime, format)

		if format == "srt" {
			sb.WriteString(fmt.Sprintf("%d\n", idx+1))
			sb.WriteString(fmt.Sprintf("%s --> %s\n", startFormatted, endFormatted))
			sb.WriteString(fmt.Sprintf("%s\n\n", s))
		} else {
			sb.WriteString(fmt.Sprintf("%d\n", idx+1))
			sb.WriteString(fmt.Sprintf("%s --> %s\n", startFormatted, endFormatted))
			sb.WriteString(fmt.Sprintf("%s\n\n", s))
		}

		currentTimeSec = endTime + 0.25 // slight natural pause between sentences
	}

	content := sb.String()
	filename := "lexio-narration-subtitles.srt"
	mime := "text/plain;charset=utf-8"
	if format == "vtt" {
		filename = "lexio-narration-captions.vtt"
		mime = "text/vtt;charset=utf-8"
	}

	downloadFile(filename, content, mime)
	ShowToast(fmt.Sprintf("Exported narration %s (%d cues)", strings.ToUpper(format), len(sentences)), "success")
}

func formatTimestamp(seconds float64, format string) string {
	totalMs := int(math.Round(seconds * 1000.0))
	hrs := totalMs / 3600000
	totalMs %= 3600000
	mins := totalMs / 60000
	totalMs %= 60000
	secs := totalMs / 1000
	ms := totalMs % 1000

	if format == "srt" {
		return fmt.Sprintf("%02d:%02d:%02d,%03d", hrs, mins, secs, ms)
	}
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hrs, mins, secs, ms)
}

func downloadFile(filename, content, mimeType string) {
	doc := js.Global().Get("document")
	window := js.Global().Get("window")

	parts := js.Global().Get("Array").New()
	parts.Call("push", content)

	opts := js.Global().Get("Object").New()
	opts.Set("type", mimeType)

	blob := js.Global().Get("Blob").New(parts, opts)
	url := window.Get("URL").Call("createObjectURL", blob)

	a := doc.Call("createElement", "a")
	a.Set("href", url)
	a.Set("download", filename)
	doc.Get("body").Call("appendChild", a)
	a.Call("click")
	doc.Get("body").Call("removeChild", a)
	window.Get("URL").Call("revokeObjectURL", url)
}

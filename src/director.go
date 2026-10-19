//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"syscall/js"
)

type PronunciationRule struct {
	Term   string `json:"term"`
	Spoken string `json:"spoken"`
}

var (
	lexiconRules        []PronunciationRule
	teleprompterFontSize = 28
	teleprompterTheme    = "dark" // "dark", "light", "high-contrast"
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

	// Teleprompter Play/Pause/Stop triggers
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

	// Find window around active word (e.g. 5 words before, 10 words after)
	windowStart := activeIdx - 6
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := activeIdx + 10
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

	// Also update teleprompter meta index
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
	// Check existing
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

	// Attach delete event listeners
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

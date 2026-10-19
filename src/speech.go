//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"strings"
	"syscall/js"
)

type RegionAccent struct {
	Code  string
	Label string
}

type BrowserVoice struct {
	Index    int
	Name     string
	Lang     string
	IsFemale bool
	IsMale   bool
}

var (
	availableVoices []BrowserVoice
	currentVoiceIdx = 0
	isReading       = false
	isPaused        = false
	currentWordIdx  = -1

	// Regional Accents map per language
	regionalAccents = map[string][]RegionAccent{
		"en": {
			{Code: "all", Label: "All English Regions"},
			{Code: "US", Label: "United States (en-US)"},
			{Code: "GB", Label: "United Kingdom (en-GB)"},
			{Code: "IN", Label: "India (en-IN)"},
			{Code: "AU", Label: "Australia (en-AU)"},
			{Code: "CA", Label: "Canada (en-CA)"},
			{Code: "IE", Label: "Ireland (en-IE)"},
			{Code: "NZ", Label: "New Zealand (en-NZ)"},
			{Code: "ZA", Label: "South Africa (en-ZA)"},
			{Code: "SG", Label: "Singapore (en-SG)"},
			{Code: "PH", Label: "Philippines (en-PH)"},
		},
		"hi": {
			{Code: "all", Label: "All Hindi Regions"},
			{Code: "IN", Label: "India (hi-IN)"},
		},
		"es": {
			{Code: "all", Label: "All Spanish Regions"},
			{Code: "ES", Label: "Spain (es-ES)"},
			{Code: "MX", Label: "Mexico (es-MX)"},
			{Code: "AR", Label: "Argentina (es-AR)"},
			{Code: "CO", Label: "Colombia (es-CO)"},
			{Code: "US", Label: "United States (es-US)"},
			{Code: "CL", Label: "Chile (es-CL)"},
			{Code: "PE", Label: "Peru (es-PE)"},
		},
		"ar": {
			{Code: "all", Label: "All Arabic Regions"},
			{Code: "SA", Label: "Saudi Arabia (ar-SA)"},
			{Code: "AE", Label: "United Arab Emirates (ar-AE)"},
			{Code: "EG", Label: "Egypt (ar-EG)"},
			{Code: "MA", Label: "Morocco (ar-MA)"},
			{Code: "JO", Label: "Jordan (ar-JO)"},
			{Code: "QA", Label: "Qatar (ar-QA)"},
		},
		"it": {
			{Code: "all", Label: "All Italian Regions"},
			{Code: "IT", Label: "Italy (it-IT)"},
			{Code: "CH", Label: "Switzerland (it-CH)"},
		},
	}
)

func initSpeechEngine() {
	synth := js.Global().Get("speechSynthesis")
	if synth.IsNull() || synth.IsUndefined() {
		ShowToast("Speech synthesis is not supported in this browser", "error")
		return
	}

	loadVoices()

	// Browser voices load asynchronously in Chromium
	synth.Set("onvoiceschanged", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		loadVoices()
		return nil
	}))

	bindSpeechControls()
}

func loadVoices() {
	synth := js.Global().Get("speechSynthesis")
	if synth.IsNull() || synth.IsUndefined() {
		return
	}

	rawVoices := synth.Call("getVoices")
	count := rawVoices.Get("length").Int()

	availableVoices = make([]BrowserVoice, 0, count)
	for i := 0; i < count; i++ {
		v := rawVoices.Index(i)
		name := v.Get("name").String()
		lang := v.Get("lang").String()

		isFemale, isMale := detectVoiceGender(name)

		availableVoices = append(availableVoices, BrowserVoice{
			Index:    i,
			Name:     name,
			Lang:     lang,
			IsFemale: isFemale,
			IsMale:   isMale,
		})
	}

	updateAccentOptions()
	filterAndRenderVoices()
}

func detectVoiceGender(name string) (bool, bool) {
	lower := strings.ToLower(name)

	femaleKeywords := []string{
		"female", "woman", "girl", "samantha", "victoria", "zira", "sonia",
		"monica", "karen", "tessa", "moira", "fiona", "luciana", "paulina",
		"helena", "alice", "veena", "leela", "ananya", "kalpana", "salma",
		"nour", "fatima", "laila", "elsa", "federica", "cosimo",
	}

	maleKeywords := []string{
		"male", "man", "boy", "david", "george", "james", "mark", "daniel",
		"oliver", "rishi", "neel", "amit", "tariq", "maged", "shakir",
		"diego", "jorge", "alberto", "carlos", "matteo", "luca",
	}

	for _, kw := range femaleKeywords {
		if strings.Contains(lower, kw) {
			return true, false
		}
	}
	for _, kw := range maleKeywords {
		if strings.Contains(lower, kw) {
			return false, true
		}
	}

	// If no explicit keyword, classify none as exclusive
	return false, false
}

func bindSpeechControls() {
	doc := js.Global().Get("document")

	// Language change -> update regional accents
	langSelect := doc.Call("getElementById", "languageSelect")
	if !langSelect.IsNull() {
		langSelect.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			updateAccentOptions()
			filterAndRenderVoices()
			return nil
		}))
	}

	accentSelect := doc.Call("getElementById", "accentSelect")
	if !accentSelect.IsNull() {
		accentSelect.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			filterAndRenderVoices()
			return nil
		}))
	}

	personaSelect := doc.Call("getElementById", "voicePersonaSelect")
	if !personaSelect.IsNull() {
		personaSelect.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			filterAndRenderVoices()
			return nil
		}))
	}

	// Sliders: Speed and Pitch
	speedRange := doc.Call("getElementById", "speedRange")
	speedValue := doc.Call("getElementById", "speedValue")
	if !speedRange.IsNull() && !speedValue.IsNull() {
		speedRange.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			val := speedRange.Get("value").String()
			speedValue.Set("textContent", val+"x")
			return nil
		}))
	}

	pitchRange := doc.Call("getElementById", "pitchRange")
	pitchValue := doc.Call("getElementById", "pitchValue")
	if !pitchRange.IsNull() && !pitchValue.IsNull() {
		pitchRange.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			val := pitchRange.Get("value").String()
			pitchValue.Set("textContent", val)
			return nil
		}))
	}

	// Playback Buttons
	startBtn := doc.Call("getElementById", "startBtn")
	if !startBtn.IsNull() {
		startBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			StartReading(0)
			return nil
		}))
	}

	pauseBtn := doc.Call("getElementById", "pauseBtn")
	if !pauseBtn.IsNull() {
		pauseBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			PauseReading()
			return nil
		}))
	}

	resumeBtn := doc.Call("getElementById", "resumeBtn")
	if !resumeBtn.IsNull() {
		resumeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ResumeReading()
			return nil
		}))
	}

	stopBtn := doc.Call("getElementById", "stopBtn")
	if !stopBtn.IsNull() {
		stopBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			StopReading()
			return nil
		}))
	}
}

func updateAccentOptions() {
	doc := js.Global().Get("document")
	langSelect := doc.Call("getElementById", "languageSelect")
	accentSelect := doc.Call("getElementById", "accentSelect")

	if langSelect.IsNull() || accentSelect.IsNull() {
		return
	}

	lang := langSelect.Get("value").String()
	if lang == "auto" {
		lang = "en"
	}

	accents, ok := regionalAccents[lang]
	if !ok {
		accents = regionalAccents["en"]
	}

	// Rebuild options HTML
	var html strings.Builder
	for _, acc := range accents {
		html.WriteString(fmt.Sprintf("<option value=\"%s\">%s</option>", acc.Code, acc.Label))
	}
	accentSelect.Set("innerHTML", html.String())
}

func filterAndRenderVoices() {
	doc := js.Global().Get("document")
	langSelect := doc.Call("getElementById", "languageSelect")
	accentSelect := doc.Call("getElementById", "accentSelect")
	personaSelect := doc.Call("getElementById", "voicePersonaSelect")
	voiceSelect := doc.Call("getElementById", "voiceSelect")
	voiceCoverage := doc.Call("getElementById", "voiceCoverage")

	if voiceSelect.IsNull() {
		return
	}

	selectedLang := "en"
	if !langSelect.IsNull() && langSelect.Get("value").String() != "auto" {
		selectedLang = langSelect.Get("value").String()
	}

	selectedAccent := "all"
	if !accentSelect.IsNull() {
		selectedAccent = accentSelect.Get("value").String()
	}

	selectedGender := "any"
	if !personaSelect.IsNull() {
		selectedGender = personaSelect.Get("value").String()
	}

	var matched []BrowserVoice

	for _, v := range availableVoices {
		// Language check
		vLangNorm := strings.ReplaceAll(strings.ToLower(v.Lang), "_", "-")
		if !strings.HasPrefix(vLangNorm, selectedLang) {
			continue
		}

		// Regional accent check
		if selectedAccent != "all" {
			targetRegion := "-" + strings.ToLower(selectedAccent)
			if !strings.HasSuffix(vLangNorm, targetRegion) && !strings.Contains(vLangNorm, targetRegion) {
				continue
			}
		}

		// Gender check: only female, male, or any (NO neutral)
		if selectedGender == "female" && !v.IsFemale {
			// If not strictly identified as female, allow if not strictly identified as male
			if v.IsMale {
				continue
			}
		} else if selectedGender == "male" && !v.IsMale {
			if v.IsFemale {
				continue
			}
		}

		matched = append(matched, v)
	}

	var html strings.Builder
	if len(matched) == 0 {
		html.WriteString("<option value=\"-1\">Default system voice</option>")
	} else {
		for _, v := range matched {
			genderTag := ""
			if v.IsFemale {
				genderTag = " (Female)"
			} else if v.IsMale {
				genderTag = " (Male)"
			}
			html.WriteString(fmt.Sprintf("<option value=\"%d\">%s [%s]%s</option>", v.Index, v.Name, v.Lang, genderTag))
		}
	}
	voiceSelect.Set("innerHTML", html.String())

	if !voiceCoverage.IsNull() {
		coverageText := fmt.Sprintf("%d voices ready", len(matched))
		if len(matched) == 0 {
			coverageText = "System fallback"
		}
		voiceCoverage.Set("textContent", coverageText)
	}
}

func StartReading(startWordIndex int) {
	synth := js.Global().Get("speechSynthesis")
	if synth.IsNull() || synth.IsUndefined() {
		return
	}

	if isReading && isPaused {
		ResumeReading()
		return
	}

	// Cancel any active utterance
	synth.Call("cancel")

	words := GetCurrentDocumentWords()
	if len(words) == 0 {
		ShowToast("No document text to read. Please enter or upload text.", "error")
		return
	}

	if startWordIndex < 0 || startWordIndex >= len(words) {
		startWordIndex = 0
	}

	// Prepare remaining text to speak with pronunciation lexicon substitutions
	textToSpeak := strings.Join(words[startWordIndex:], " ")
	if strings.TrimSpace(textToSpeak) == "" {
		return
	}

	phoneticText := ApplyPronunciationLexicon(textToSpeak)
	utterance := js.Global().Get("SpeechSynthesisUtterance").New(phoneticText)

	// Configure speed & pitch
	doc := js.Global().Get("document")
	speedRange := doc.Call("getElementById", "speedRange")
	pitchRange := doc.Call("getElementById", "pitchRange")
	voiceSelect := doc.Call("getElementById", "voiceSelect")

	rate := 1.0
	if !speedRange.IsNull() {
		fmt.Sscanf(speedRange.Get("value").String(), "%f", &rate)
	}
	pitch := 1.0
	if !pitchRange.IsNull() {
		fmt.Sscanf(pitchRange.Get("value").String(), "%f", &pitch)
	}

	utterance.Set("rate", rate)
	utterance.Set("pitch", pitch)

	// Assign selected voice
	if !voiceSelect.IsNull() {
		selectedVal := voiceSelect.Get("value").String()
		var voiceIdx int
		if n, _ := fmt.Sscanf(selectedVal, "%d", &voiceIdx); n == 1 && voiceIdx >= 0 {
			rawVoices := synth.Call("getVoices")
			if voiceIdx < rawVoices.Get("length").Int() {
				utterance.Set("voice", rawVoices.Index(voiceIdx))
			}
		}
	}

	runningWordOffset := startWordIndex

	// Boundary tracking for synchronized highlight & live teleprompter
	utterance.Set("onboundary", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			charIndex := args[0].Get("charIndex").Int()
			// Calculate word index from char offset
			prefix := textToSpeak
			if charIndex < len(prefix) {
				prefix = prefix[:charIndex]
			}
			wordOffset := len(strings.Fields(prefix))
			activeIdx := runningWordOffset + wordOffset
			if activeIdx < len(words) {
				currentWordIdx = activeIdx
				SetActiveWord(activeIdx)
				UpdateTeleprompter(activeIdx)
				updateReadingProgress(activeIdx, len(words))
			}
		}
		return nil
	}))

	utterance.Set("onend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isReading = false
		isPaused = false
		ClearActiveWord()
		updatePlaybackUI(false, false)
		updateReadingProgress(len(words), len(words))
		return nil
	}))

	utterance.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isReading = false
		isPaused = false
		ClearActiveWord()
		updatePlaybackUI(false, false)
		return nil
	}))

	isReading = true
	isPaused = false
	updatePlaybackUI(true, false)

	synth.Call("speak", utterance)
}

func PauseReading() {
	synth := js.Global().Get("speechSynthesis")
	if !synth.IsNull() && !synth.IsUndefined() && isReading && !isPaused {
		synth.Call("pause")
		isPaused = true
		updatePlaybackUI(true, true)
	}
}

func ResumeReading() {
	synth := js.Global().Get("speechSynthesis")
	if !synth.IsNull() && !synth.IsUndefined() && isReading && isPaused {
		synth.Call("resume")
		isPaused = false
		updatePlaybackUI(true, false)
	}
}

func StopReading() {
	synth := js.Global().Get("speechSynthesis")
	if !synth.IsNull() && !synth.IsUndefined() {
		synth.Call("cancel")
	}
	isReading = false
	isPaused = false
	ClearActiveWord()
	updatePlaybackUI(false, false)
}

func updatePlaybackUI(reading, paused bool) {
	doc := js.Global().Get("document")
	startBtn := doc.Call("getElementById", "startBtn")
	pauseBtn := doc.Call("getElementById", "pauseBtn")
	resumeBtn := doc.Call("getElementById", "resumeBtn")
	stopBtn := doc.Call("getElementById", "stopBtn")
	playbackLabel := doc.Call("getElementById", "playbackLabel")

	if !startBtn.IsNull() {
		startBtn.Set("disabled", reading && !paused)
	}
	if !pauseBtn.IsNull() {
		pauseBtn.Set("disabled", !reading || paused)
	}
	if !resumeBtn.IsNull() {
		resumeBtn.Set("disabled", !reading || !paused)
	}
	if !stopBtn.IsNull() {
		stopBtn.Set("disabled", !reading)
	}

	if !playbackLabel.IsNull() {
		if !reading {
			playbackLabel.Set("textContent", "Idle")
		} else if paused {
			playbackLabel.Set("textContent", "Paused")
		} else {
			playbackLabel.Set("textContent", "Reading")
		}
	}
}

func updateReadingProgress(current, total int) {
	if total <= 0 {
		return
	}
	pct := (current * 100) / total
	if pct > 100 {
		pct = 100
	}

	doc := js.Global().Get("document")
	progressLabel := doc.Call("getElementById", "progressLabel")
	progressBar := doc.Call("getElementById", "progressBar")

	if !progressLabel.IsNull() {
		progressLabel.Set("textContent", fmt.Sprintf("%d%%", pct))
	}
	if !progressBar.IsNull() {
		progressBar.Get("style").Set("width", fmt.Sprintf("%d%%", pct))
	}
}

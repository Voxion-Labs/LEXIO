//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
	"time"
)

var (
	activeViewID = "dashboardView"
)

func initUI() {
	doc := js.Global().Get("document")

	// Tab switching for views
	tabs := doc.Call("querySelectorAll", ".workspace-tab")
	tabCount := tabs.Get("length").Int()
	for i := 0; i < tabCount; i++ {
		tab := tabs.Index(i)
		tab.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			viewAttr := this.Call("getAttribute", "data-view").String()
			if viewAttr != "" {
				SwitchView(viewAttr)
			}
			return nil
		}))
	}

	// Focus mode button
	focusBtn := doc.Call("getElementById", "toggleFocusModeBtn")
	if !focusBtn.IsNull() {
		focusBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			body := doc.Get("body")
			isFocus := body.Get("classList").Call("toggle", "focus-mode").Bool()
			if isFocus {
				focusBtn.Set("innerHTML", "<i class=\"fas fa-compress\"></i> Exit focus")
			} else {
				focusBtn.Set("innerHTML", "<i class=\"fas fa-expand\"></i> Focus mode")
			}
			return nil
		}))
	}

	// Side Panels: Manual and Notes
	openManualBtn := doc.Call("getElementById", "openManualBtn")
	manualPanel := doc.Call("getElementById", "manualPanel")
	if !openManualBtn.IsNull() && !manualPanel.IsNull() {
		openManualBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			openSidePanel("manualPanel")
			return nil
		}))
	}

	openNotesBtn := doc.Call("getElementById", "openNotesBtn")
	notesPanel := doc.Call("getElementById", "notesPanel")
	if !openNotesBtn.IsNull() && !notesPanel.IsNull() {
		openNotesBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			loadSavedNotes()
			openSidePanel("notesPanel")
			return nil
		}))
	}

	// Close panel buttons
	closeBtns := doc.Call("querySelectorAll", "[data-close-panel]")
	for i := 0; i < closeBtns.Get("length").Int(); i++ {
		btn := closeBtns.Index(i)
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			panelID := this.Call("getAttribute", "data-close-panel").String()
			closeSidePanel(panelID)
			return nil
		}))
	}

	// Backdrop click closes all panels
	backdrop := doc.Call("getElementById", "panelBackdrop")
	if !backdrop.IsNull() {
		backdrop.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			closeAllSidePanels()
			return nil
		}))
	}

	// Notes save button
	saveNotesBtn := doc.Call("getElementById", "saveNotesBtn")
	if !saveNotesBtn.IsNull() {
		saveNotesBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			saveNotes()
			return nil
		}))
	}

	// Dismiss loading screen
	dismissLoadingScreen()
}

func SwitchView(viewID string) {
	activeViewID = viewID
	doc := js.Global().Get("document")

	// Update all tabs matching data-view
	tabs := doc.Call("querySelectorAll", ".workspace-tab")
	for i := 0; i < tabs.Get("length").Int(); i++ {
		tab := tabs.Index(i)
		matches := tab.Call("getAttribute", "data-view").String() == viewID
		if matches {
			tab.Get("classList").Call("add", "active")
		} else {
			tab.Get("classList").Call("remove", "active")
		}
	}

	// Update workspace views
	views := doc.Call("querySelectorAll", ".workspace-view")
	for i := 0; i < views.Get("length").Int(); i++ {
		view := views.Index(i)
		if view.Get("id").String() == viewID {
			view.Get("classList").Call("add", "active")
		} else {
			view.Get("classList").Call("remove", "active")
		}
	}
}

func openSidePanel(panelID string) {
	doc := js.Global().Get("document")
	panel := doc.Call("getElementById", panelID)
	backdrop := doc.Call("getElementById", "panelBackdrop")

	if !panel.IsNull() {
		panel.Get("classList").Call("add", "open")
	}
	if !backdrop.IsNull() {
		backdrop.Get("classList").Call("add", "active")
	}
}

func closeSidePanel(panelID string) {
	doc := js.Global().Get("document")
	panel := doc.Call("getElementById", panelID)
	backdrop := doc.Call("getElementById", "panelBackdrop")

	if !panel.IsNull() {
		panel.Get("classList").Call("remove", "open")
	}
	if !backdrop.IsNull() {
		backdrop.Get("classList").Call("remove", "active")
	}
}

func closeAllSidePanels() {
	doc := js.Global().Get("document")
	panels := doc.Call("querySelectorAll", ".side-panel")
	for i := 0; i < panels.Get("length").Int(); i++ {
		panels.Index(i).Get("classList").Call("remove", "open")
	}
	backdrop := doc.Call("getElementById", "panelBackdrop")
	if !backdrop.IsNull() {
		backdrop.Get("classList").Call("remove", "active")
	}
}

func loadSavedNotes() {
	storage := js.Global().Get("localStorage")
	if storage.IsNull() || storage.IsUndefined() {
		return
	}
	saved := storage.Call("getItem", "lexio_workspace_notes")
	if !saved.IsNull() && !saved.IsUndefined() {
		doc := js.Global().Get("document")
		input := doc.Call("getElementById", "notesInput")
		if !input.IsNull() {
			input.Set("value", saved.String())
		}
	}
}

func saveNotes() {
	doc := js.Global().Get("document")
	input := doc.Call("getElementById", "notesInput")
	status := doc.Call("getElementById", "notesStatus")

	if !input.IsNull() {
		val := input.Get("value").String()
		storage := js.Global().Get("localStorage")
		if !storage.IsNull() && !storage.IsUndefined() {
			storage.Call("setItem", "lexio_workspace_notes", val)
		}
		if !status.IsNull() {
			status.Set("textContent", "Saved locally on this device")
		}
		ShowToast("Notes saved locally", "success")
	}
}

func ShowToast(message, tone string) {
	doc := js.Global().Get("document")
	toast := doc.Call("getElementById", "globalToast")
	if toast.IsNull() {
		toast = doc.Call("createElement", "div")
		toast.Set("id", "globalToast")
		toast.Set("className", "status-indicator status-floating")
		doc.Get("body").Call("appendChild", toast)
	}

	toast.Set("textContent", message)
	toast.Get("classList").Call("add", "visible")

	go func() {
		time.Sleep(3 * time.Second)
		js.Global().Get("document").Call("getElementById", "globalToast").Get("classList").Call("remove", "visible")
	}()
}

func dismissLoadingScreen() {
	go func() {
		time.Sleep(350 * time.Millisecond)
		doc := js.Global().Get("document")
		loading := doc.Call("getElementById", "loadingScreen")
		if !loading.IsNull() {
			loading.Get("classList").Call("add", "fade-out")
			time.Sleep(400 * time.Millisecond)
			loading.Get("style").Set("display", "none")
		}
	}()
}

//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"
)

type UserAccount struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}

var (
	currentAuthMode = "create"
	currentUser     *UserAccount
	avatarDataURL   = ""
)

func initAuth() {
	doc := js.Global().Get("document")

	// Mode tabs
	tabCreateAcc := doc.Call("getElementById", "tabCreateAcc")
	tabSignIn := doc.Call("getElementById", "tabSignIn")
	if !tabCreateAcc.IsNull() && !tabCreateAcc.IsUndefined() {
		tabCreateAcc.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			setAuthMode("create")
			return nil
		}))
	}
	if !tabSignIn.IsNull() && !tabSignIn.IsUndefined() {
		tabSignIn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			setAuthMode("signin")
			return nil
		}))
	}

	// Avatar picker
	avatarPickerBtn := doc.Call("getElementById", "avatarPickerBtn")
	avatarInput := doc.Call("getElementById", "avatarInput")
	if !avatarPickerBtn.IsNull() && !avatarInput.IsNull() {
		avatarPickerBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				args[0].Call("preventDefault")
			}
			avatarInput.Call("click")
			return nil
		}))

		avatarInput.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			event := args[0]
			files := event.Get("target").Get("files")
			if files.Get("length").Int() == 0 {
				avatarDataURL = ""
				syncLoginAvatarPreview()
				return nil
			}

			file := files.Index(0)
			fileType := file.Get("type").String()
			if !strings.HasPrefix(fileType, "image/") {
				ShowToast("Please upload an image file (PNG, JPG, SVG, WebP)", "error")
				avatarInput.Set("value", "")
				return nil
			}

			reader := js.Global().Get("FileReader").New()
			reader.Set("onload", js.FuncOf(func(rThis js.Value, rArgs []js.Value) interface{} {
				res := rArgs[0].Get("target").Get("result")
				avatarDataURL = res.String()
				syncLoginAvatarPreview()
				ShowToast("Display photo updated", "success")
				return nil
			}))
			reader.Call("readAsDataURL", file)
			return nil
		}))
	}

	// Name input listener for live avatar preview
	nameInput := doc.Call("getElementById", "name")
	if !nameInput.IsNull() {
		nameInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if avatarDataURL == "" {
				syncLoginAvatarPreview()
			}
			return nil
		}))
	}

	// Auth form submit
	loginForm := doc.Call("getElementById", "loginForm")
	if !loginForm.IsNull() {
		loginForm.Call("addEventListener", "submit", js.FuncOf(handleAuthSubmit))
	}

	// Quick Demo buttons
	quickDemoBtn := doc.Call("getElementById", "quickDemoBtn")
	if !quickDemoBtn.IsNull() {
		quickDemoBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			runQuickDemo()
			return nil
		}))
	}

	modalQuickDemoBtn := doc.Call("getElementById", "modalQuickDemoBtn")
	if !modalQuickDemoBtn.IsNull() {
		modalQuickDemoBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			closeOptionsModal()
			runQuickDemo()
			return nil
		}))
	}

	// Options Modal handlers
	openOptionsBtn := doc.Call("getElementById", "openOptionsBtn")
	if !openOptionsBtn.IsNull() {
		openOptionsBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			openOptionsModal()
			return nil
		}))
	}

	closeOptionsBtn := doc.Call("getElementById", "closeOptionsBtn")
	if !closeOptionsBtn.IsNull() {
		closeOptionsBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			closeOptionsModal()
			return nil
		}))
	}

	saveOptionsBtn := doc.Call("getElementById", "saveOptionsBtn")
	if !saveOptionsBtn.IsNull() {
		saveOptionsBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			saveOptions()
			return nil
		}))
	}

	// Get Started scroll button
	navGetStartedBtn := doc.Call("getElementById", "navGetStartedBtn")
	if !navGetStartedBtn.IsNull() {
		navGetStartedBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			card := doc.Call("querySelector", ".auth-card")
			if !card.IsNull() {
				opts := js.Global().Get("Object").New()
				opts.Set("behavior", "smooth")
				opts.Set("block", "center")
				card.Call("scrollIntoView", opts)
			}
			return nil
		}))
	}

	// Logout button
	logoutBtn := doc.Call("getElementById", "logoutBtn")
	if !logoutBtn.IsNull() {
		logoutBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			handleLogout()
			return nil
		}))
	}

	// Dev bypass
	urlParams := js.Global().Get("URLSearchParams").New(js.Global().Get("window").Get("location").Get("search"))
	if urlParams.Call("has", "dev").Bool() {
		launchWorkspace(&UserAccount{
			Name:     "Developer",
			Username: "developer",
			Email:    "dev@local.lexio",
			Phone:    "+1 415 555 0100",
		})
	}

	syncLoginAvatarPreview()
}

func setAuthMode(mode string) {
	currentAuthMode = mode
	doc := js.Global().Get("document")

	tabCreateAcc := doc.Call("getElementById", "tabCreateAcc")
	tabSignIn := doc.Call("getElementById", "tabSignIn")
	nameContainer := doc.Call("getElementById", "nameContainer")
	avatarRow := doc.Call("getElementById", "avatarRow")
	authSubmitBtn := doc.Call("getElementById", "authSubmitBtn")

	if mode == "create" {
		if !tabCreateAcc.IsNull() {
			tabCreateAcc.Get("classList").Call("add", "active")
		}
		if !tabSignIn.IsNull() {
			tabSignIn.Get("classList").Call("remove", "active")
		}
		if !nameContainer.IsNull() {
			nameContainer.Get("style").Set("display", "flex")
		}
		if !avatarRow.IsNull() {
			avatarRow.Get("style").Set("display", "flex")
		}
		if !authSubmitBtn.IsNull() {
			authSubmitBtn.Set("innerHTML", "<span>Create Account &amp; Enter</span> <i class=\"fas fa-arrow-right\"></i>")
		}
	} else {
		if !tabSignIn.IsNull() {
			tabSignIn.Get("classList").Call("add", "active")
		}
		if !tabCreateAcc.IsNull() {
			tabCreateAcc.Get("classList").Call("remove", "active")
		}
		if !nameContainer.IsNull() {
			nameContainer.Get("style").Set("display", "none")
		}
		if !avatarRow.IsNull() {
			avatarRow.Get("style").Set("display", "none")
		}
		if !authSubmitBtn.IsNull() {
			authSubmitBtn.Set("innerHTML", "<span>Sign In to Workspace</span> <i class=\"fas fa-arrow-right\"></i>")
		}
	}
}

func handleAuthSubmit(this js.Value, args []js.Value) interface{} {
	if len(args) > 0 {
		args[0].Call("preventDefault")
	}

	doc := js.Global().Get("document")
	nameInput := doc.Call("getElementById", "name")
	usernameInput := doc.Call("getElementById", "username")
	emailInput := doc.Call("getElementById", "email")
	passwordInput := doc.Call("getElementById", "password")
	phoneInput := doc.Call("getElementById", "phone")

	name := ""
	if !nameInput.IsNull() {
		name = strings.TrimSpace(nameInput.Get("value").String())
	}
	username := ""
	if !usernameInput.IsNull() {
		username = strings.TrimSpace(usernameInput.Get("value").String())
	}
	email := ""
	if !emailInput.IsNull() {
		email = strings.TrimSpace(emailInput.Get("value").String())
	}
	password := ""
	if !passwordInput.IsNull() {
		password = strings.TrimSpace(passwordInput.Get("value").String())
	}
	phone := "+1 415 555 0188"
	if !phoneInput.IsNull() && phoneInput.Get("value").String() != "" {
		phone = strings.TrimSpace(phoneInput.Get("value").String())
	}

	if email == "" || password == "" {
		ShowToast("Please enter your email and password", "error")
		return nil
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		ShowToast("Please provide a valid email address", "error")
		return nil
	}
	if len(password) < 4 {
		ShowToast("Password must be at least 4 characters long", "error")
		return nil
	}

	if currentAuthMode == "create" {
		if name == "" {
			ShowToast("Please enter your full name", "error")
			return nil
		}
		if username == "" {
			ShowToast("Please enter a username", "error")
			return nil
		}

		user := &UserAccount{
			Name:     name,
			Username: username,
			Email:    email,
			Phone:    phone,
			Avatar:   avatarDataURL,
		}

		// Store account locally
		saveAccountLocally(user)
		launchWorkspace(user)
	} else {
		// Sign in mode
		user := &UserAccount{
			Name:     username,
			Username: username,
			Email:    email,
			Phone:    phone,
			Avatar:   avatarDataURL,
		}
		if user.Name == "" {
			user.Name = strings.Split(email, "@")[0]
		}
		launchWorkspace(user)
	}

	return nil
}

func saveAccountLocally(user *UserAccount) {
	storage := js.Global().Get("localStorage")
	if storage.IsNull() || storage.IsUndefined() {
		return
	}
	raw := storage.Call("getItem", "lexio_accounts")
	var accounts []UserAccount
	if !raw.IsNull() && !raw.IsUndefined() {
		json.Unmarshal([]byte(raw.String()), &accounts)
	}
	accounts = append(accounts, *user)
	data, _ := json.Marshal(accounts)
	storage.Call("setItem", "lexio_accounts", string(data))
}

func runQuickDemo() {
	demoUser := &UserAccount{
		Name:     "Liam Brooks",
		Username: "liam_brooks",
		Email:    "liam.brooks@voxionlabs.io",
		Phone:    "+1 415 555 0188",
		Avatar:   "",
	}
	launchWorkspace(demoUser)
}

func launchWorkspace(user *UserAccount) {
	currentUser = user
	doc := js.Global().Get("document")

	userNameDisplay := doc.Call("getElementById", "userName")
	if !userNameDisplay.IsNull() {
		userNameDisplay.Set("textContent", user.Name)
	}

	userAvatar := doc.Call("getElementById", "userAvatar")
	userInitials := doc.Call("getElementById", "userInitials")
	paintAvatar(userAvatar, userInitials, user.Name, user.Avatar, getUserInitials(user.Name))

	// Switch pages
	loginPage := doc.Call("getElementById", "loginPage")
	mainApp := doc.Call("getElementById", "mainApp")
	if !loginPage.IsNull() {
		loginPage.Get("style").Set("display", "none")
	}
	if !mainApp.IsNull() {
		mainApp.Get("style").Set("display", "flex")
	}

	SwitchView("dashboardView")

	ShowToast("Welcome, "+user.Name+"!", "success")
}

func handleLogout() {
	StopReading()

	currentUser = nil
	avatarDataURL = ""

	doc := js.Global().Get("document")
	loginForm := doc.Call("getElementById", "loginForm")
	if !loginForm.IsNull() {
		loginForm.Call("reset")
	}

	syncLoginAvatarPreview()

	mainApp := doc.Call("getElementById", "mainApp")
	loginPage := doc.Call("getElementById", "loginPage")
	if !mainApp.IsNull() {
		mainApp.Get("style").Set("display", "none")
	}
	if !loginPage.IsNull() {
		loginPage.Get("style").Set("display", "flex")
		js.Global().Get("window").Call("scrollTo", 0, 0)
	}

	ShowToast("Logged out successfully", "info")
}

func getUserInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "U"
	}
	if len(parts) == 1 {
		return strings.ToUpper(parts[0][:1])
	}
	return strings.ToUpper(parts[0][:1] + parts[1][:1])
}

func paintAvatar(el, initialsEl js.Value, name, imgData, fallbackInitials string) {
	if el.IsNull() || el.IsUndefined() {
		return
	}
	initials := fallbackInitials
	if initials == "" {
		initials = getUserInitials(name)
	}

	style := el.Get("style")
	style.Set("backgroundImage", "")
	style.Set("backgroundSize", "")
	style.Set("backgroundPosition", "")
	style.Set("color", "#FFFFFF")

	if imgData != "" {
		style.Set("background", "#0468D7")
		style.Set("backgroundImage", fmt.Sprintf("url(%s)", imgData))
		style.Set("backgroundSize", "cover")
		style.Set("backgroundPosition", "center")
		if !initialsEl.IsNull() && !initialsEl.IsUndefined() {
			initialsEl.Get("style").Set("opacity", "0")
		} else {
			el.Set("textContent", "")
		}
		return
	}

	// Calculate deterministic gradient hue
	hue := 205
	for _, c := range name {
		hue = (hue + int(c)) % 360
	}
	grad := fmt.Sprintf("linear-gradient(135deg, hsl(%d, 82%%, 48%%), hsl(%d, 88%%, 56%%))", hue, (hue+45)%360)
	style.Set("background", grad)

	if !initialsEl.IsNull() && !initialsEl.IsUndefined() {
		initialsEl.Set("textContent", initials)
		initialsEl.Get("style").Set("opacity", "1")
	} else {
		el.Set("textContent", initials)
	}
}

func syncLoginAvatarPreview() {
	doc := js.Global().Get("document")
	preview := doc.Call("getElementById", "loginAvatarPreview")
	nameInput := doc.Call("getElementById", "name")
	name := "Liam Brooks"
	if !nameInput.IsNull() && strings.TrimSpace(nameInput.Get("value").String()) != "" {
		name = strings.TrimSpace(nameInput.Get("value").String())
	}
	paintAvatar(preview, js.Null(), name, avatarDataURL, getUserInitials(name))
}

func openOptionsModal() {
	doc := js.Global().Get("document")
	modal := doc.Call("getElementById", "optionsModal")
	if !modal.IsNull() {
		modal.Get("classList").Call("add", "active")
	}
}

func closeOptionsModal() {
	doc := js.Global().Get("document")
	modal := doc.Call("getElementById", "optionsModal")
	if !modal.IsNull() {
		modal.Get("classList").Call("remove", "active")
	}
}

func saveOptions() {
	doc := js.Global().Get("document")
	optLang := doc.Call("getElementById", "optDefaultLang")
	optAccent := doc.Call("getElementById", "optDefaultAccent")
	optPace := doc.Call("getElementById", "optDefaultPace")

	storage := js.Global().Get("localStorage")
	if !storage.IsNull() && !storage.IsUndefined() {
		if !optLang.IsNull() {
			storage.Call("setItem", "lexio_opt_lang", optLang.Get("value").String())
		}
		if !optAccent.IsNull() {
			storage.Call("setItem", "lexio_opt_accent", optAccent.Get("value").String())
		}
		if !optPace.IsNull() {
			storage.Call("setItem", "lexio_opt_pace", optPace.Get("value").String())
		}
	}
	ShowToast("Workspace preferences saved", "success")
	closeOptionsModal()
}

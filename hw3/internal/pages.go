package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"hw3/assets"
	"net/http"
)

var (
	LoginTpl        *template.Template
	RegistrationTpl *template.Template
	ResetTpl        *template.Template
	ConfirmResetTpl *template.Template
	//ServiceTpl      *template.Template
)

func InitTemplates() {
	loginHtml := assets.MustAssetString("templates/login.html")
	LoginTpl = template.Must(template.New("login_view").Parse(loginHtml))

	regHtml := assets.MustAssetString("templates/registration.html")
	RegistrationTpl = template.Must(template.New("registration_view").Parse(regHtml))

	resetHtml := assets.MustAssetString("templates/reset.html")
	ResetTpl = template.Must(template.New("reset_view").Parse(resetHtml))

	confirmResetHtml := assets.MustAssetString("templates/confirmReset.html")
	ConfirmResetTpl = template.Must(template.New("confirm_reset_view").Parse(confirmResetHtml))
}

// Render a template, or server error.
func render(w http.ResponseWriter, r *http.Request, tpl *template.Template, name string, data interface{}) {
	buf := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(buf, name, data); err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
	w.Write(buf.Bytes())
}

// Push the given resource to the client.
func push(w http.ResponseWriter, resource string) {
	pusher, ok := w.(http.Pusher)
	if ok {
		if err := pusher.Push(resource, nil); err == nil {
			return
		}
	}
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	push(w, "/static/registration.css")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	render(w, r, RegistrationTpl, "registration_view", nil)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	push(w, "/static/login.css")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	render(w, r, LoginTpl, "login_view", nil)
}

func ResetHandler(w http.ResponseWriter, r *http.Request) {
	push(w, "/static/reset.css")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	render(w, r, ResetTpl, "reset_view", nil)
}

func ConfirmResetHandler(w http.ResponseWriter, r *http.Request) {
	push(w, "/static/reset.css")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	render(w, r, ConfirmResetTpl, "confirm_reset_view", nil)
}

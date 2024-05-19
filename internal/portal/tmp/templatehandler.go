// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package templates

import (
	"embed"
	"text/template"
)

type TemplateHandler struct {
	TmplHome     *template.Template
	TmplLogin    *template.Template
	TmplRegister *template.Template
}

//go:embed templates/*.html
var templates embed.FS

func NewTemplateHandler() *TemplateHandler {
	mainTemplate := []string{"templates/main.html", "templates/main.style.html", "templates/header.html", "templates/nav.html", "templates/footer.html"}
	homeTemplate := "templates/home.html"
	loginTemplate := "templates/login.html"
	registerTemplate := "templates/register.html"

	return &TemplateHandler{
		TmplHome:     template.Must(template.ParseFS(templates, append(mainTemplate, homeTemplate)...)),
		TmplLogin:    template.Must(template.ParseFS(templates, append(mainTemplate, loginTemplate)...)),
		TmplRegister: template.Must(template.ParseFS(templates, append(mainTemplate, registerTemplate)...)),
	}
}

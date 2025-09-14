package util

import (
	"embed"
	"html/template"
	"net/url"
)

//go:embed templates/*.html
var tmplFS embed.FS

func NewTemplate(name string) *template.Template {
	funcMap := template.FuncMap {
		"pathescape": url.PathEscape,
	}
	return template.Must(template.New(name).Funcs(funcMap).ParseFS(tmplFS, "templates/" + name))
}

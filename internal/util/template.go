package util

import (
	"embed"
	"html/template"
	"net/url"

	"github.com/akikareha/himewiki/internal/format"
)

//go:embed templates/*.html
var tmplFS embed.FS

func formatDiff(text string) template.HTML {
	return template.HTML(format.Diff(text))
}

func NewTemplate(name string) *template.Template {
	funcMap := template.FuncMap {
		"pathescape": url.PathEscape,
		"fmtdiff": formatDiff,
	}
	return template.Must(template.New(name).Funcs(funcMap).ParseFS(tmplFS, "templates/" + name))
}

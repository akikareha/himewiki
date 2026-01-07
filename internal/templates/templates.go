package templates

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/akikareha/himewiki/internal/format"
)

//go:embed templates/*.html
var tmplFS embed.FS

func formatDiff(text string) template.HTML {
	return template.HTML(format.Diff(text))
}

func Render(w http.ResponseWriter, name string, data any) error {
	funcMap := template.FuncMap{
		"pathescape": url.PathEscape,
		"fmtdiff":    formatDiff,
	}
	t := template.Must(template.New(name+".html").Funcs(funcMap).ParseFS(tmplFS, "templates/"+name+".html"))
	err := t.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to process template", http.StatusInternalServerError)
		log.Println("failed to process template:", err)
	}
	return err
}

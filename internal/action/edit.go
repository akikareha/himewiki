package action 

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
	"github.com/akikareha/himewiki/internal/format"
)

func render(text string) string {
	if strings.HasPrefix(text, "=") {
		return format.Creole(text)
	} else if strings.HasPrefix(text, "#") {
		return format.Markdown(text)
	} else {
		return format.Nomark(text)
	}
}

func View(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	content, err := data.Load(params.Name)
	if err != nil {
		http.Redirect(w, r, "/"+url.PathEscape(params.Name)+"?a=edit", http.StatusFound)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/view.html"))
	escaped := url.PathEscape(params.Name)
	rendered := render(content)
	tmpl.Execute(w, struct {
		SiteName string
		Name string
		Escaped string
		Rendered template.HTML
	}{
		SiteName: cfg.Site.Name,
		Name: params.Name,
		Escaped: escaped,
		Rendered: template.HTML(rendered),
	})
}

func Edit(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	var previewed bool
	var content string
	if r.Method != http.MethodPost {
		previewed = false
		content, _ = data.Load(params.Name)
	} else {
		previewed = r.FormValue("previewed") == "true"
		content = r.FormValue("content")
		preview := r.FormValue("preview")
		save := r.FormValue("save")
		if previewed && save != "" {
			if err := data.Save(params.Name, content); err != nil {
				http.Error(w, "Failed to save", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/"+url.PathEscape(params.Name), http.StatusFound)
			return
		} else if preview != "" {
			previewed = true
		} else {
			http.Error(w, "Invalid state", http.StatusInternalServerError)
			return
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/edit.html"))
	escaped := url.PathEscape(params.Name)
	rendered := render(content)
	tmpl.Execute(w, struct {
		SiteName string
		Name string
		Escaped string
		Text string
		Rendered template.HTML 
		Previewed bool
	}{
		SiteName: cfg.Site.Name,
		Name: params.Name,
		Escaped: escaped,
		Text: content,
		Rendered: template.HTML(rendered),
		Previewed: previewed,
	})
}

func All(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	pages, err := data.LoadAll()
	if err != nil {
		http.Error(w, "Failed to load pages", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/all.html"))
	tmpl.Execute(w, struct {
		SiteName string
		Pages []data.Name
	}{
		SiteName: cfg.Site.Name,
		Pages: pages,
	})
}

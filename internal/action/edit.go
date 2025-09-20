package action 

import (
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
	"github.com/akikareha/himewiki/internal/filter"
	"github.com/akikareha/himewiki/internal/format"
	"github.com/akikareha/himewiki/internal/util"
)

func render(cfg *config.Config, title string, text string) (string, string, string, string) {
	if strings.HasPrefix(text, "=") {
		return format.Creole(cfg, title, text)
	} else if strings.HasPrefix(text, "#") {
		return format.Markdown(cfg, title, text)
	} else {
		return format.Nomark(cfg, title, text)
	}
}

func summarize(s string, n int) string {
	if n < 2 {
		panic("n is too small")
	}

	var b strings.Builder
	long := false
	i := 0
	for _, r := range(s) {
		if i >= n - 2 {
			long = true
			break
		}
		if r == '\r' || r == '\n' || r == '\t' {
			r = ' '
		}
		b.WriteRune(r)
		i += 1
	}

	if long {
		return b.String() + ".."
	} else {
		return b.String()
	}
}

func View(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	_, content, err := data.Load(params.Name)
	if err != nil {
		http.Redirect(w, r, "/"+url.PathEscape(params.Name)+"?a=edit", http.StatusFound)
		return
	}

	tmpl := util.NewTemplate("view.html")
	title, _, plain, rendered := render(cfg, params.Name, content)
	summary := summarize(plain, 144)
	tmpl.Execute(w, struct {
		Base string
		SiteName string
		Card string
		Name string
		Summary string
		Title string
		Rendered template.HTML
	}{
		Base: cfg.Site.Base,
		SiteName: cfg.Site.Name,
		Card: cfg.Site.Card,
		Name: params.Name,
		Summary: summary,
		Title: title,
		Rendered: template.HTML(rendered),
	})
}

func Edit(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	var previewed bool
	var revisionID int
	var content string
	var preview string
	var save string
	if r.Method != http.MethodPost {
		previewed = false
		revisionID, content, _ = data.Load(params.Name)
		preview = ""
		save = ""
	} else {
		previewed = r.FormValue("previewed") == "true"
		var err error
		revisionID, err = strconv.Atoi(r.FormValue("revision_id"))
		if err != nil || revisionID < 0 {
			http.Error(w, "Invalid revision ID", http.StatusInternalServerError)
			return
		}
		content = r.FormValue("content")
		preview = r.FormValue("preview")
		save = r.FormValue("save")
	}

	var filtered string
	var err error
	if previewed && save != "" {
		filtered, err = filter.Apply(cfg, params.Name, content)
	} else {
		filtered, err = content, nil
	}
	if err != nil {
		http.Error(w, "Failed to filter content", http.StatusInternalServerError)
		return
	}
	title, normalized, _, rendered := render(cfg, params.Name, filtered)

	if previewed && save != "" {
		if err := data.Save(params.Name, normalized, revisionID); err != nil {
			http.Error(w, "Failed to save", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/"+url.PathEscape(params.Name), http.StatusFound)
		return
	} else if preview != "" {
		previewed = true
	}

	tmpl := util.NewTemplate("edit.html")
	tmpl.Execute(w, struct {
		SiteName string
		Name string
		Previewed bool
		RevisionID int
		Text string
		Title string
		Rendered template.HTML 
	}{
		SiteName: cfg.Site.Name,
		Name: params.Name,
		Previewed: previewed,
		RevisionID: revisionID,
		Text: normalized,
		Title: title,
		Rendered: template.HTML(rendered),
	})
}

func All(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	pages, err := data.LoadAll()
	if err != nil {
		http.Error(w, "Failed to load pages", http.StatusInternalServerError)
		return
	}

	tmpl := util.NewTemplate("all.html")
	tmpl.Execute(w, struct {
		SiteName string
		Pages []string
	}{
		SiteName: cfg.Site.Name,
		Pages: pages,
	})
}

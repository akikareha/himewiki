package action

import (
	"html/template"
	"net/http"
	"net/url"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
)

func Revisions(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	revs, err := data.LoadRevisions(params.Name)
	if err != nil {
		http.Error(w, "Failed to load revisions", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/revisions.html"))
	escaped := url.PathEscape(params.Name)
tmpl.Execute(w, struct {
		SiteName string
		Name string
		Escaped string
		Revisions []data.Revision
	}{
		SiteName: cfg.Site.Name,
		Name: params.Name,
		Escaped: escaped,
		Revisions: revs,
	})
}

func Revert(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	if params.ID == nil {
		http.Error(w, "Bad revision id", 400);
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusInternalServerError)
		return
	}

	err := data.Revert(params.Name, *params.ID)
	if err != nil {
		http.Error(w, "Failed to revert", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/" + url.PathEscape(params.Name), http.StatusFound)
}

func ViewRevision(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	if params.ID == nil {
		http.Error(w, "Bad revision id", 400);
		return
	}

	content, err := data.LoadRevision(params.Name, *params.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}


	tmpl := template.Must(template.ParseFiles("templates/revision.html"))
	escaped := url.PathEscape(params.Name)
	rendered := render(content)
	tmpl.Execute(w, struct {
		SiteName string
		Name string
		Escaped string
		Rendered template.HTML
		ID int
	}{
		SiteName: cfg.Site.Name,
		Name: params.Name,
		Escaped: escaped,
		Rendered: template.HTML(rendered),
		ID: *params.ID,
	})
}

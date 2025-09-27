package action

import (
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
	"github.com/akikareha/himewiki/internal/format"
	"github.com/akikareha/himewiki/internal/util"
)

func Revisions(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	pageStr := r.URL.Query().Get("p")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	revs, err := data.LoadRevisions(params.DbName, page, perPage)
	if err != nil {
		http.Error(w, "Failed to load revisions", http.StatusInternalServerError)
		return
	}

	tmpl := util.NewTemplate("revisions.html")
tmpl.Execute(w, struct {
		SiteName string
		Name string
		Title string
		Revisions []data.Revision
		NextPage int
	}{
		SiteName: cfg.Site.Name,
		Name: params.Name,
		Title: params.DbName,
		Revisions: revs,
		NextPage: page + 1,
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

	err := data.Revert(params.DbName, *params.ID)
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

	content, err := data.LoadRevision(params.DbName, *params.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}


	tmpl := util.NewTemplate("revision.html")
	title, _, _, rendered := format.Apply(cfg, params.DbName, content)
	tmpl.Execute(w, struct {
		SiteName string
		Name string
		Title string
		Rendered template.HTML
		ID int
	}{
		SiteName: cfg.Site.Name,
		Name: params.Name,
		Title: title,
		Rendered: template.HTML(rendered),
		ID: *params.ID,
	})
}

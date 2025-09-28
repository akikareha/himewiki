package action 

import (
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
	"github.com/akikareha/himewiki/internal/filter"
	"github.com/akikareha/himewiki/internal/format"
	"github.com/akikareha/himewiki/internal/util"
)

func View(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	_, content, err := data.Load(params.DbName)
	if err != nil {
		http.Redirect(w, r, "/"+url.PathEscape(params.Name)+"?a=edit", http.StatusFound)
		return
	}

	tmpl := util.NewTemplate("view.html")
	title, _, plain, rendered := format.Apply(cfg, params.DbName, content)
	summary := format.Summarize(plain, 144)
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
		revisionID, content, _ = data.Load(params.DbName)
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
		filtered, err = filter.Apply(cfg, params.DbName, content)
	} else {
		filtered, err = content, nil
	}
	if err != nil {
		http.Error(w, "Failed to filter content", http.StatusInternalServerError)
		return
	}
	title, normalized, _, rendered := format.Apply(cfg, params.DbName, filtered)

	if previewed && save != "" {
		if err := data.Save(cfg, params.DbName, normalized, revisionID); err != nil {
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

const perBigPage = 500

func All(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	pageStr := r.URL.Query().Get("p")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	pages, err := data.LoadAll(page, perBigPage)
	if err != nil {
		http.Error(w, "Failed to load pages", http.StatusInternalServerError)
		return
	}

	tmpl := util.NewTemplate("all.html")
	tmpl.Execute(w, struct {
		SiteName string
		Pages []string
		NextPage int
	}{
		SiteName: cfg.Site.Name,
		Pages: pages,
		NextPage: page + 1,
	})
}

const perPage = 50

func Recent(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	pageStr := r.URL.Query().Get("p")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	records, err := data.Recent(page, perPage)
	if err != nil {
		http.Error(w, "Failed to load pages", http.StatusInternalServerError)
		return
	}

	tmpl := util.NewTemplate("recent.html")
	tmpl.Execute(w, struct {
		SiteName string
		Records []data.RecentRecord
		NextPage int
	}{
		SiteName: cfg.Site.Name,
		Records: records,
		NextPage: page + 1,
	})
}

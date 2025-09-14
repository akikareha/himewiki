package action

import (
	"net/http"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
	"github.com/akikareha/himewiki/internal/util"
)

func Search(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	word := r.URL.Query().Get("w")
	var results []string
	if word != "" {
		searchType := r.URL.Query().Get("t")
		if searchType == "name" {
			results, _ = data.SearchNames(word)
		} else if searchType == "content" {
			results, _ = data.SearchContents(word)
		} else {
			http.NotFound(w, r)
			return
		}
	}

	tmpl := util.NewTemplate("search.html")
	tmpl.Execute(w, struct {
		SiteName string
		Word    string
		Results []string
	}{
		SiteName: cfg.Site.Name,
		Word:    word,
		Results: results,
	})
}

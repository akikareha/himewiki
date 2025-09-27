package action

import (
	"net/http"
	"strconv"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
	"github.com/akikareha/himewiki/internal/util"
)

func Search(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	pageStr := r.URL.Query().Get("p")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	word := r.URL.Query().Get("w")
	searchType := r.URL.Query().Get("t")
	var results []string
	if word != "" {
		if searchType == "name" {
			results, _ = data.SearchNames(word, page, perBigPage)
		} else if searchType == "content" {
			results, _ = data.SearchContents(word, page, perBigPage)
		} else {
			http.NotFound(w, r)
			return
		}
	}

	if searchType == "" {
		searchType = "name"
	}

	tmpl := util.NewTemplate("search.html")
	tmpl.Execute(w, struct {
		SiteName string
		Type string
		Word string
		Results []string
		NextPage int
	}{
		SiteName: cfg.Site.Name,
		Type: searchType,
		Word:    word,
		Results: results,
		NextPage: page + 1,
	})
}

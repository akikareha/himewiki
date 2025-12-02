package action

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
	"github.com/akikareha/himewiki/internal/templates"
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
	for i := 0; i < len(results); i++ {
		r := results[i]
		if strings.IndexByte(r, '.') != -1 {
			results[i] = r + ".wiki"
		}
	}

	if searchType == "" {
		searchType = "name"
	}

	tmpl := templates.New("search.html")
	tmpl.Execute(w, struct {
		SiteName string
		Type     string
		Word     string
		Results  []string
		NextPage int
	}{
		SiteName: cfg.Site.Name,
		Type:     searchType,
		Word:     word,
		Results:  results,
		NextPage: page + 1,
	})
}

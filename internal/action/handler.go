package action

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/akikareha/himewiki/internal/config"
)

type Params struct {
	Name   string
	Action string
	ID     *int
}

func parse(cfg *config.Config, r *http.Request) Params {
	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" {
	  name = cfg.Site.Front
	}

	action := r.URL.Query().Get("a")

	idStr := r.URL.Query().Get("i")
	id, err := strconv.Atoi(idStr)
	var idRef = &id
	if err != nil || id <= 0 {
		idRef = nil
	}

	return Params{Name: name, Action: action, ID: idRef}
}

func handle(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	params := parse(cfg, r)

	switch params.Action {
	case "static":
		Static(cfg, w, r, &params)
	case "", "view":
		View(cfg, w, r, &params)
	case "edit":
		Edit(cfg, w, r, &params)
	case "all":
		All(cfg, w, r, &params)
	case "revs":
		Revisions(cfg, w, r, &params)
	case "revert":
		Revert(cfg, w, r, &params)
	case "rev":
		ViewRevision(cfg, w, r, &params)
	case "search":
		Search(cfg, w, r, &params)
	default:
		http.NotFound(w, r)
	}
}

func Handler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handle(cfg, w, r)
	}
}

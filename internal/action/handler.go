package action

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/akikareha/himewiki/internal/config"
)

type Params struct {
	Name   string
	DbName string
	Ext    string
	Action string
	ID     *int
}

func parse(cfg *config.Config, r *http.Request) Params {
	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" {
		name = cfg.Wiki.Front
	}
	ext := filepath.Ext(name)
	if ext == "" {
		if name[0] == '.' {
			ext = name
		}
	}
	dbName := name
	if ext == ".wiki" {
		dbName = name[:len(name)-len(ext)]
	}
	if ext == "" {
		ext = ".wiki"
	}
	ext = ext[1:]

	action := r.URL.Query().Get("a")

	idStr := r.URL.Query().Get("i")
	id, err := strconv.Atoi(idStr)
	var idRef = &id
	if err != nil || id <= 0 {
		idRef = nil
	}

	return Params{
		Name:   name,
		DbName: dbName,
		Ext:    ext,
		Action: action,
		ID:     idRef,
	}
}

func handle(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	params := parse(cfg, r)

	if params.Ext == "wiki" {
		switch params.Action {
		case "info":
			Info(cfg, w, r, &params)
		case "", "view":
			View(cfg, w, r, &params)
		case "edit":
			Edit(cfg, w, r, &params)
		case "all":
			All(cfg, w, r, &params)
		case "recent":
			Recent(cfg, w, r, &params)
		case "revs":
			Revisions(cfg, w, r, &params)
		case "revert":
			Revert(cfg, w, r, &params)
		case "rev":
			ViewRevision(cfg, w, r, &params)
		case "search":
			Search(cfg, w, r, &params)
		case "upload":
			Upload(cfg, w, r, &params)
		case "allimgs":
			AllImages(cfg, w, r, &params)
		default:
			http.NotFound(w, r)
		}
	} else {
		switch params.Action {
		case "", "view":
			ViewImage(cfg, w, r, &params)
		default:
			http.NotFound(w, r)
		}
	}
}

func Handler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handleStatic(cfg, w, r) {
			return
		}
		handle(cfg, w, r)
	}
}

package action

import (
	"embed"
	"mime"
	"net/http"

	"github.com/akikareha/himewiki/internal/config"
)

//go:embed static/*
var static embed.FS

func Static(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	data, err := static.ReadFile("static/" + params.DbName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	mimeType := mime.TypeByExtension("." + params.Ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)

	w.Write(data)
}

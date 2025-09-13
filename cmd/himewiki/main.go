package main

import (
	"net/http"

	"github.com/akikareha/himewiki/internal/action"
	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
)

func main() {
	cfg := config.Load("config.yaml")

	db := data.Connect(cfg)
	defer db.Close()

	http.HandleFunc("/", action.Handler(cfg))
	http.ListenAndServe(cfg.App.Addr, nil)
}

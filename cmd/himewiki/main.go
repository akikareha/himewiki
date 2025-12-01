package main

import (
	"net/http"
	"os"

	"github.com/akikareha/himewiki/internal/action"
	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
)

func main() {
	if len(os.Args) < 2 {
		print("Usage: " + os.Args[0] + " himewiki.yaml\n")
		return
	}
	cfg := config.Load(os.Args[1])

	db := data.Connect(cfg)
	defer db.Close()

	http.HandleFunc("/", action.Handler(cfg))
	http.ListenAndServe(cfg.App.Addr, nil)
}

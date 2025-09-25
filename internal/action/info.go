package action

import (
	"net/http"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/util"
)


func Info(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	public := config.Publish(cfg)

	tmpl := util.NewTemplate("info.html")
	tmpl.Execute(w, struct {
		SiteName string
		Public config.Public
	}{
		SiteName: cfg.Site.Name,
		Public: public,
	})
}

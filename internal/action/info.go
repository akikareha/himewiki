package action

import (
	"net/http"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/data"
	"github.com/akikareha/himewiki/internal/util"
)


func Info(cfg *config.Config, w http.ResponseWriter, r *http.Request, params *Params) {
	stat := data.Stat()
	public := config.Publish(cfg)

	tmpl := util.NewTemplate("info.html")
	tmpl.Execute(w, struct {
		SiteName string
		Stat data.Info
		Public config.Public
	}{
		SiteName: cfg.Site.Name,
		Stat: stat,
		Public: public,
	})
}

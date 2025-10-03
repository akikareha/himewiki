package action

import (
	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/filter"
	"github.com/akikareha/himewiki/internal/format"
	"github.com/akikareha/himewiki/internal/data"
)

func runGnome(cfg *config.Config, targetRecent int) {
	recentNames, err := data.RecentNames(cfg.Gnome.Recent)
	if err != nil {
		return
	}
	if len(recentNames) < 1 {
		return
	}
	targetName := recentNames[targetRecent % len(recentNames)]

	revisionID, content, err := data.Load(targetName)
	if err != nil {
		return
	}

	filtered, err := filter.GnomeApply(cfg, targetName, content)
	_, normalized, _, _ := format.Apply(cfg, targetName, filtered)

	_, err = data.Save(cfg, targetName, normalized, revisionID)
	if err != nil {
		return
	}
}

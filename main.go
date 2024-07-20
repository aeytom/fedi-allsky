package main

import (
	"github.com/aeytom/fedi-allsky/app"
	"github.com/aeytom/fedi-allsky/toot"
)

func main() {
	cfg := app.LoadConfig()
	cfg.Allsky.Init(cfg.Logger())
	defer cfg.Allsky.DbClose()

	m := toot.Init(&cfg.Mastodon, cfg.Logger())
	m.RegisterAllsky(&cfg.Allsky)
	go cfg.Allsky.ListenAllskyHttp(m)
	m.WatchNotifications()
	m.ProcessNotifications()
}

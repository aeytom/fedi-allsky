package main

import (
	"github.com/aeytom/fedi-allsky/app"
	"github.com/aeytom/fedi-allsky/toot"
)

func main() {
	cfg := app.LoadConfig()
	cfg.Allsky.Init(cfg.Logger())

	m := toot.Init(&cfg.Mastodon, cfg.Logger())
	m.RegisterAllsky(&cfg.Allsky)
	m.ProcessNotifications()

	go cfg.Allsky.ListenMotionWebhook(m)
	m.WatchNotifications()

}

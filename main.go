package main

import (
	"time"

	"github.com/aeytom/fedi-allsky/app"
	"github.com/aeytom/fedi-allsky/dump1090"
	"github.com/aeytom/fedi-allsky/toot"
)

func main() {
	cfg := app.LoadConfig()

	dump1090 := dump1090.Init(&cfg.Dump1090, cfg.Logger())
	defer dump1090.Close()
	dump1090.Tick(*time.NewTicker(5 * time.Second))

	cfg.Allsky.Init(cfg.Logger(), dump1090)
	defer cfg.Allsky.DbClose()

	m := toot.Init(&cfg.Mastodon, cfg.Logger())
	m.RegisterAllsky(&cfg.Allsky)
	go cfg.Allsky.ListenAllskyHttp(m)
	m.WatchNotifications()
	m.ProcessNotifications()
}

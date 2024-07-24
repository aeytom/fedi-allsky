package allsky

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aeytom/fedilib"
	"github.com/mattn/go-mastodon"
)

type Config struct {
	LocalUrl        string  `yaml:"local_url,omitempty" json:"local_url,omitempty"`
	PublicUrl       string  `yaml:"public_url,omitempty" json:"public_url,omitempty"`
	ListenPort      int16   `yaml:"listen_port,omitempty" json:"listen_port,omitempty"`
	ListenHost      string  `yaml:"listen_host,omitempty" json:"listen_host,omitempty"`
	MinStarCount    int64   `yaml:"min_star_count,omitempty" json:"min_star_count,omitempty"`
	MinIssAlititude float64 `yaml:"min_iss_alititude,omitempty" json:"min_iss_alititude,omitempty"`
	MinMeteorCount  int64   `yaml:"min_meteor_count,omitempty" json:"min_meteor_count,omitempty"`
	//
	SqliteDb string `yaml:"sqlite_db,omitempty" json:"sqlite_db,omitempty"`
	//
	log  *log.Logger
	db   *sql.DB
	toot fedilib.Toot
}

func (s *Config) Init(log *log.Logger) {
	if s.LocalUrl == "" {
		s.LocalUrl = "http://allsky.fritz.box"
	}
	if s.PublicUrl == "" {
		s.PublicUrl = "https://allsky.tay-tec.de"
	}
	if s.ListenHost == "" {
		s.ListenHost = "127.0.0.1"
	}
	if s.ListenPort == 0 {
		s.ListenPort = 18888
	}
	if s.MinStarCount == 0 {
		s.MinStarCount = 900
	}
	if s.MinIssAlititude == 0.0 {
		s.MinIssAlititude = 30.0
	}
	if s.MinMeteorCount == 0 {
		s.MinMeteorCount = 1
	}

	if s.SqliteDb == "" {
		s.SqliteDb = "allsky-post.db?mode=rwc&_journal=wal"
	}

	s.log = log
	s.DbOpen(s.SqliteDb)
}

func (s *Config) Current() (io.ReadCloser, error) {
	url := s.LocalUrl + "/current/tmp/image.jpg?_ts=" + fmt.Sprint(time.Now().Unix())
	return s.ImageHttp(url)
}

// Image gets a named image
// http://allsky.fritz.box/images/20240713/image-20240714015605.jpg
func (s *Config) Image(date string, file string) (io.ReadCloser, error) {
	url := s.LocalUrl + "/images/" + date + "/" + file
	return s.ImageHttp(url)
}

// Image gets a image with HTTP
func (s *Config) ImageHttp(url string) (io.ReadCloser, error) {
	log.Print(url)
	if resp, err := http.Get(url); err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, errors.New("invalid status code " + fmt.Sprint(resp.StatusCode))
	} else if resp.Header.Get("content-type") != "image/jpeg" {
		return nil, errors.New("invalid content type " + resp.Header.Get("content-type"))
	} else {
		// defer resp.Body.Close()
		return resp.Body, nil
	}
}

func (s *Config) tootAllskyParams(p *AllskyParams, status *mastodon.Status) error {
	text := "New #allsky image from " + p.as_date + "-" + p.as_time + "\n\n" +
		"Sunset 🌇:  " + p.as_sun_sunset.Format(time.DateTime) + "\n" +
		"Sunrise 🌅: " + p.as_sun_sunrise.Format(time.DateTime) + "\n" +
		"\nSky\n" +
		fmt.Sprintf("Star count:   %d\n", p.as_starcount) +
		fmt.Sprintf("Meteor count: %d\n", p.as_meteorcount) +
		fmt.Sprintf("ISS altitude: %.1f°\n", p.as_25544alt) +
		"\nCamera\n" +
		fmt.Sprintf("Gain:         %d\n", p.as_gain) +
		fmt.Sprintf("Exposure:     %v\n", p.as_exposure_us) +
		fmt.Sprintf("Temperature:  %d°C (sensor)\n", p.as_temperature_c) +
		"\n" +
		"use /report to get more info about the last night.\n" +
		"\n" +
		s.PublicUrl + "\n"

	toot := mastodon.Toot{
		Status:     text,
		Visibility: mastodon.VisibilityPublic,
		Language:   "en",
	}
	if status != nil {
		toot.Status = fmt.Sprintf("Hello %s (@%s),\n\n", status.Account.Username, status.Account.Acct) + toot.Status
		toot.Visibility = mastodon.VisibilityDirectMessage
		toot.InReplyToID = status.ID
	}

	if ifile, err := s.Image(p.date_name, filepath.Base(p.current_image)); err != nil {
		return err
	} else if err := s.toot.TootWithImageReader(toot, ifile, "Allsky Image"); err != nil {
		return err
	}
	return nil
}

func (s *Config) TootBest(status *mastodon.Status) error {

	date_name := s.dbGetDateName()
	if date_name == "" {
		return errors.New("data not found")
	}

	p, err := s.dbGetBestStarcount(date_name)
	if err != nil {
		return errors.Join(errors.New("data not found for date_name "+date_name), err)
	}

	return s.tootAllskyParams(p, status)
}

func (s *Config) TootMeteorCount(status *mastodon.Status, n string) error {

	date_name := s.dbGetDateName()
	if date_name == "" {
		return errors.New("data not found")
	}

	if n == "" {
		p, err := s.dbGetBestMeteors(date_name)
		if err != nil {
			return errors.Join(errors.New("data not found for date_name "+date_name), err)
		}
		return s.tootAllskyParams(p, status)
	} else {
		if lm, err := s.dbListMeteors(date_name); err != nil {
			return err
		} else if idx, err := strconv.ParseInt(n, 10, 64); err != nil {
			return errors.Join(errors.New(n), err)
		} else if int64(len(lm)) >= idx {
			return s.tootAllskyParams(lm[idx-1], status)
		}
	}
	return errors.New("meteor data not found")
}

func (s *Config) TootIssVisible(status *mastodon.Status) error {

	date_name := s.dbGetDateName()
	if date_name == "" {
		return errors.New("data not found")
	}

	p, err := s.dbGetBestIss(date_name)
	if err != nil {
		return errors.Join(errors.New("data not found for date_name "+date_name), err)
	}

	return s.tootAllskyParams(p, status)
}

func (s *Config) TootMeteorList(status *mastodon.Status) error {

	date_name := s.dbGetDateName()
	if date_name == "" {
		return errors.New("data not found")
	}

	dn := mustParseDateTimeSplit(date_name, "")

	text := "#allsky »/report« for " + dn.Format(time.DateOnly) + "\n"

	if cs, err := s.dbGetBestStarcount(date_name); err != nil {
		s.log.Println(err)
	} else {
		csd := mustParseDateTimeSplit(cs.as_date, cs.as_time)
		text += fmt.Sprintf("\nMost visible stars »%d« counted at %s\n", cs.as_starcount, csd.Format(time.DateTime))
	}

	if lm, err := s.dbListMeteors(date_name); err != nil {
		s.log.Println(err)
	} else {
		text += "\nMeteor detected at time (visible stars):\n"
		for idx, m := range lm {
			md := mustParseDateTimeSplit(m.as_date, m.as_time)
			text += fmt.Sprintf("- /meteor%d – %d at %s (%d)\n", (idx + 1), m.as_meteorcount, md.Format(time.TimeOnly), m.as_starcount)
		}
	}

	toot := mastodon.Toot{
		Status:     text,
		Visibility: mastodon.VisibilityPublic,
		Language:   "en",
	}
	if status != nil {
		toot.Status = fmt.Sprintf("Hello %s (@%s),\n\n", status.Account.Username, status.Account.Acct) + toot.Status
		toot.Visibility = mastodon.VisibilityDirectMessage
		toot.InReplyToID = status.ID
	}

	if len(toot.Status) >= 500 {
		toot.Status = toot.Status[0:497] + "\n…"
	}

	return s.toot.TootWithImage(toot, "")
}

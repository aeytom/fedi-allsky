package toot

import (
	"fmt"

	"github.com/mattn/go-mastodon"
)

func (s *Config) RegisterAllsky(cfg Allsky) {
	s.allsky = cfg
}

func (s *Config) cmdAllsky(status *mastodon.Status) error {

	s.allsky.Current()

	ir, err := s.allsky.Current()
	if err != nil {
		return err
	}

	var m []mastodon.ID
	if a, err := s.Client().UploadMediaFromReader(s.Ctx(), ir); err != nil {
		return err
	} else {
		a.Description = "Allsky Bild"
		m = []mastodon.ID{a.ID}
	}

	toot := mastodon.Toot{
		Status:      fmt.Sprintf("@%s\n\nThe current #allsky image.", status.Account.Acct),
		InReplyToID: status.ID,
		MediaIDs:    m,
		Visibility:  mastodon.VisibilityDirectMessage,
		Language:    "en",
	}
	if _, err := s.Client().PostStatus(s.Ctx(), &toot); err != nil {
		return err
	}
	return nil
}

func (s *Config) cmdBestStarcount(status *mastodon.Status) error {
	err := s.allsky.TootBest(status)
	return err
}

func (s *Config) cmdMeteorCount(status *mastodon.Status) error {
	err := s.allsky.TootMeteorCount(status)
	return err
}

func (s *Config) cmdMeteorList(status *mastodon.Status) error {
	err := s.allsky.TootMeteorList(status)
	return err
}

func (s *Config) cmdIssVisible(status *mastodon.Status) error {
	err := s.allsky.TootIssVisible(status)
	return err
}

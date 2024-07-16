package toot

import (
	"fmt"

	"github.com/mattn/go-mastodon"
)

func (s *Config) RegisterAllsky(cfg Allsky) {
	s.allsky = cfg
}

func (s *Config) cmdAllsky(status *mastodon.Status) error {
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

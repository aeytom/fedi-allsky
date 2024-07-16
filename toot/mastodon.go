package toot

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"

	"github.com/aeytom/fedilib"
	"github.com/mattn/go-mastodon"
)

type MotionEye interface {
	Action(camera string, action string) (string, error)
	LastPhoto(camera string) (string, error)
}

type Allsky interface {
	Current() (io.ReadCloser, error)
}

type Config struct {
	fedilib.Fedi
	allsky Allsky
}

func Init(cfg *fedilib.Config, log *log.Logger) *Config {
	m := &Config{}
	m.Fedi.Init(cfg, m, log)
	return m
}

func (s *Config) HandleNotification(n *mastodon.Notification) {
	s.Log().Print("notification ", n.Type, " from ", n.Account.Acct)
	if n.Account.Bot {
		s.Log().Print("ignore bot")
		return
	}
	switch n.Type {
	case "mention":
		s.Log().Print("… mention ", n.Account.Acct, " :: ", n.Status.Content)
		s.handleMention(n)
	case "follow":
		s.Log().Print("… followed ", n.Account.Acct)
		s.sendHelp(&n.Account, "")
	case "follow_request":
		s.Log().Print("… follow_request from ", n.Account.Acct)
		s.answerFollowRequest(&n.Account)
	default:
		s.Log().Print("unbekannter notification type ", n.Type)
	}
	s.Client().DismissNotification(s.Ctx(), n.ID)
}

func (s *Config) handleMention(n *mastodon.Notification) {

	var err error
	var text string

	if text, err = fedilib.StripHtmlFromString(n.Status.Content); err != nil {
		s.Log().Fatal(err)
	}

	s.Log().Printf("handleMention from: %#v :: %#v", n.Status.Account.Acct, text)
	cmd := regexp.MustCompile(`/(help|allsky)\b`).FindString(text)
	s.Log().Print("command " + cmd)

	switch cmd {
	case "/help":
		s.sendHelp(&n.Status.Account, "")
	case "/allsky":
		err = s.cmdAllsky(n.Status)
	default:
		err = errors.New("does not understand your message")
	}

	if err != nil {
		log.Println(err)
		s.sendHelp(&n.Status.Account, err.Error())
	}
}

func (s *Config) sendHelp(account *mastodon.Account, msg string) error {

	if err := s.IsFollower(account); err != nil {
		return s.pleaseFollowToUseCommands(account)
	}

	if msg == "" {
		msg = `I will post #allsky images on
		- clear sky
		- ISS transits
		- detected meteors
		`
	}

	s.MarkAccount(account, "Pending")
	help := `
Hello %s (@%s)!

%s

I understand these commands:

- "/help"     – diese Meldung
- "/allsky"   – das aktuelle Bild der Allsky Kamera 

Your %s
@%s
`
	t := &mastodon.Toot{
		Status:     fmt.Sprintf(help, account.DisplayName, account.Acct, msg, s.CurrentAccount().DisplayName, s.CurrentAccount().Acct),
		Visibility: mastodon.VisibilityDirectMessage,
		Language:   "en",
	}
	_, err := s.Client().PostStatus(s.Ctx(), t)
	return err
}

func (s *Config) pleaseFollowToUseCommands(account *mastodon.Account) error {

	tpl := `
Hello %s (@%s)!

Please follow me to receive #allsky images and use commands.

%s
@%s
`
	t := &mastodon.Toot{
		Status: fmt.Sprintf(
			tpl,
			account.DisplayName,
			account.Acct,
			s.CurrentAccount().DisplayName,
			s.CurrentAccount().Acct),
		Visibility: mastodon.VisibilityDirectMessage,
		Language:   "en",
	}
	st, err := s.Client().PostStatus(s.Ctx(), t)
	s.Log().Printf("sent follow message %s to @%s", st.ID, account.Acct)
	return err
}

func (s *Config) answerFollowRequest(account *mastodon.Account) error {

	pg := &mastodon.Pagination{
		Limit: 40,
	}
	fr := false
	if la, err := s.Client().GetFollowRequests(s.Ctx(), pg); err != nil {
		s.Log().Fatal(err)
	} else {
		for _, l := range la {
			if l.ID == account.ID {
				fr = true
				break
			}
		}
	}
	if !fr {
		s.Log().Print("ignore follow request from ", account.Acct)
		return s.sendHelp(account, "Please follow me")
	}

	s.MarkAccount(account, "Pending")
	help := `
Willkommen %s (@%s)!

Ich bin der MotionEye und Allsky Bot.

Bitte antworte einfach auf diesen Toot und stelle dich vor bevor ich dir antworte.

Dein %s
@%s
`
	t := &mastodon.Toot{
		Status:     fmt.Sprintf(help, account.DisplayName, account.Acct, s.CurrentAccount().DisplayName, s.CurrentAccount().Acct),
		Visibility: mastodon.VisibilityDirectMessage,
		Language:   "en",
	}
	_, err := s.Client().PostStatus(s.Ctx(), t)
	return err
}

package allsky

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aeytom/fedilib"
)

type Config struct {
	LocalUrl        string  `yaml:"local_url,omitempty" json:"local_url,omitempty"`
	PublicUrl       string  `yaml:"public_url,omitempty" json:"public_url,omitempty"`
	ListenPort      int16   `yaml:"listen_port,omitempty" json:"listen_port,omitempty"`
	ListenHost      string  `yaml:"listen_host,omitempty" json:"listen_host,omitempty"`
	MinStarCount    int64   `yaml:"min_star_count,omitempty"`
	MinIssAlititude float64 `yaml:"min_iss_alititude,omitempty"`
	MinMeteorCount  int64   `yaml:"min_meteor_count,omitempty"`

	//
	log  *log.Logger
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
	s.log = log
}

func (s *Config) Current() (io.ReadCloser, error) {
	url := s.LocalUrl + "/current/tmp/image.jpg?_ts=" + fmt.Sprint(time.Now().Unix())
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

// Image gets a named image
// http://allsky.fritz.box/images/20240713/image-20240714015605.jpg
func (s *Config) Image(date string, file string) (io.ReadCloser, error) {
	if resp, err := http.Get(s.LocalUrl + "/images/" + date + "/" + file); err != nil {
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

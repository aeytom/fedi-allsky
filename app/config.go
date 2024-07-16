package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aeytom/fedi-allsky/allsky"
	"github.com/aeytom/fedilib"
	"github.com/go-yaml/yaml"
)

type Env struct {
	AppDir   string         `yaml:"dir,omitempty" json:"app_dir,omitempty"`
	Mastodon fedilib.Config `yaml:"mastodon,omitempty" json:"mastodon,omitempty"`
	Allsky   allsky.Config  `yaml:"allsky,omitempty" json:"allsky,omitempty"`
}

type Settings struct {
	Env
	log *log.Logger
}

var (
	Config Settings
)

func LoadConfig() *Settings {

	Config.log = log.Default()

	help := flag.Bool("help", false, "show command line usage")
	envPath := getEnvArg("DOT_ENV", "dotEnv", "env.yaml", "dot env path (YAML)")
	showCfg := flag.Bool("showCfg", false, "show config content")
	flag.Parse()

	if ep, err := filepath.Abs(*envPath); err != nil {
		log.Fatalln(*envPath, err)
	} else {
		*envPath = ep
	}

	ed, err := os.ReadFile(*envPath)
	if err != nil {
		log.Fatalln(*envPath, err)
	} else {
		err = yaml.Unmarshal([]byte(ed), &Config.Env)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if Config.AppDir == "" {
		Config.AppDir = "."
	}

	Config.AppDir, err = filepath.Abs(Config.AppDir)
	if err != nil {
		log.Fatalln(err)
	}

	if *showCfg {
		Config.Show()
		os.Exit(0)
	}

	if *help {
		Config.Usage()
		os.Exit(0)
	}

	return &Config
}

func getEnvArg(env string, arg string, dflt string, usage string) *string {
	ev, avail := os.LookupEnv(env)
	if avail {
		dflt = ev
	}
	v := flag.String(arg, dflt, usage)
	return v
}

func (s *Settings) Logger() *log.Logger {
	return s.log
}

func (s *Settings) Show() {
	cb, _ := yaml.Marshal(s)
	fmt.Println(string(cb))
}

func (s *Settings) Usage() {
	fmt.Println("")
	fmt.Printf("== Usage %s ==\n", os.Args[0])
	fmt.Println("")
	s.Show()
	fmt.Println("")
	fmt.Printf("Run: %s -dotEnv .env.yaml\n", os.Args[0])
	fmt.Println("")
}

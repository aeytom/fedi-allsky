package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

// Command line args
var (
	ArgWorkdir  string
	ArgPort     string
	ArgBotToken string
	ArgImage    string
	ArgMUrl     string
	ArgS3Bucket string
)

// ParseArgs parses command line flags
func ParseArgs() {
	pdir := getEnvArg("TNOTIFY_DIR", "dir", ".", "base directory")
	ArgPort = *getEnvArg("TNOTIFY_PORT", "port", "18358", "command port")
	ArgBotToken = *getEnvArg("TNOTIFY_BOT_TOKEN", "token", "1557531115:AAGC6dsxMyZhX9ULqBwc4fYJSuXmoRxuRBI", "Telegram bot token")
	ArgImage = *getEnvArg("TNOTIFY_IMAGE", "image", "/data/output/Camera1/lastsnap.jpg", "last camera image")
	ArgMUrl = *getEnvArg("MOTION_CONTROL_URL", "motion-url", "http://localhost:7999", "Motion Webcontrol URL")
	ArgS3Bucket = *getEnvArg("BUCKET", "bucket", "tay-birdfeed", "AWS S3 bucket name")
	flag.Parse()

	dir, err := filepath.Abs(*pdir)
	if err != nil {
		panic(err)
	}
	ArgWorkdir = dir

	log.Printf("arg dir := '%s'", dir)
	log.Printf("arg port := %s", ArgPort)
	log.Printf("arg telegram token := '%s'", ArgBotToken)
	log.Printf("arg motion control url := '%s'", ArgMUrl)
}

func getEnvArg(env string, arg string, dflt string, usage string) *string {
	ev, avail := os.LookupEnv(env)
	if avail {
		dflt = ev
	}
	v := flag.String(arg, dflt, usage)
	return v
}

package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	cameras map[string]string
)

func motionCtrlRequest(path string, contentType string) ([]byte, error) {
	log.Print(ArgMUrl + path)
	res, err := http.Get(ArgMUrl + path)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("Unexpected response from motion webcontrol: " + res.Status)
	}

	if contentType != "" && !strings.HasPrefix(res.Header.Get("Content-Type"), contentType) {
		return nil, errors.New("Unexpected response: " + res.Header.Get("Content-Type"))
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func motionWebcontrolHtmlOutput(value bool) error {
	onoff := "1"
	contentType := "text/plain"
	if value {
		onoff = "2"
		contentType = "text/html"
	}
	_, err := motionCtrlRequest("/0/config/set?webcontrol_interface="+onoff, contentType)
	return err
}

func motionAction(camera string, action string) (string, error) {
	if err := motionWebcontrolHtmlOutput(false); err != nil {
		return "", err
	}

	action = regexp.MustCompile(`[^a-z]+`).ReplaceAllString(action, "_")

	body, err := motionCtrlRequest("/"+camera+"/action/"+action, "text/plain")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func motionConfigGet(camera string, setting string) (string, error) {
	if err := motionWebcontrolHtmlOutput(false); err != nil {
		return "", err
	}

	setting = regexp.MustCompile(`[^a-z_-]`).ReplaceAllString(setting, "_")

	body, err := motionCtrlRequest("/"+camera+"/config/get?query="+setting, "text/plain")
	if err != nil {
		return "", err
	}

	if matches := regexp.MustCompile(`(?m)^` + setting + `\s*=\s*(.*?)\s*$`).FindSubmatch(body); matches != nil {
		return string(matches[1]), nil
	}
	return "", errors.New("Motion get setting failed: " + setting)
}

func motionGetCameras() error {
	if err := motionWebcontrolHtmlOutput(true); err != nil {
		return err
	}

	body, err := motionCtrlRequest("/", "text/html")
	if err != nil {
		return err
	}

	rea := regexp.MustCompile(`(?m)^<a href='/(\d+)/'>(.+?)</a>`)

	cameras = make(map[string]string)
	for i, m := range rea.FindAllStringSubmatch(string(body), -1) {
		if i > 0 {
			cameras[m[1]] = m[2]
		}
	}

	return nil
}

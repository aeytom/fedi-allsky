package main

import (
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"

	"github.com/oliamb/cutter"
)

type CropParam struct {
	file   string
	cx     int
	cy     int
	width  int
	height int
}

func crop(p CropParam) (string, error) {
	motImage, err := os.Open(p.file)
	if err != nil {
		return "", err
	}
	defer motImage.Close()

	imgCfg, _, err := image.DecodeConfig(motImage)
	if err != nil {
		return "", err
	}

	if 0 >= p.cx || p.cx >= imgCfg.Width {
		return p.file, nil
	}

	if 0 >= p.cy || p.cy >= imgCfg.Height {
		return p.file, nil
	}

	if imgCfg.Width < p.width || imgCfg.Width < p.cx+p.width/2 || 0 > p.cx-p.width/2 {
		return p.file, nil
	}

	if imgCfg.Height < p.height || imgCfg.Height < p.cy+p.height/2 || 0 > p.cy-p.height/2 {
		return p.file, nil
	}

	motImage.Seek(0, 0)
	imgData, _, err := image.Decode(motImage)
	if err != nil {
		return "", err
	}

	croppedImg, err := cutter.Crop(imgData, cutter.Config{
		Width:  p.width,
		Height: p.height,
		Anchor: image.Point{p.cx, p.cy},
		Mode:   cutter.Centered, // optional, default value
	})
	if err != nil {
		return "", err
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	jpeg.Encode(tmpFile, croppedImg, &jpeg.Options{Quality: 80})
	return tmpFile.Name(), nil
}

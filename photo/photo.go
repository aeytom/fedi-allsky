package photo

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"

	"github.com/oliamb/cutter"
)

type CropParam struct {
	File             string
	CenterHorizontal int
	CenterVertical   int
	Width            int
	Height           int
	BorderPercent    int
}

func (p *CropParam) Crop() (string, error) {
	motImage, err := os.Open(p.File)
	if err != nil {
		return "", err
	}
	defer motImage.Close()

	imgCfg, _, err := image.DecodeConfig(motImage)
	if err != nil {
		return "", err
	}

	if p.BorderPercent > 0 {
		// extend crop size
		p.Width += p.Width * p.BorderPercent / 100
		p.Height += p.Height * p.BorderPercent / 100
	}

	if 0 >= p.CenterHorizontal || p.CenterHorizontal >= imgCfg.Width {
		// horizontal center out of bounds
		p.CenterHorizontal = imgCfg.Width / 2
	}

	if 0 >= p.CenterVertical || p.CenterVertical >= imgCfg.Height {
		// vertical center out of bounds
		p.CenterVertical = imgCfg.Height / 2
	}

	if 0 >= p.Width || p.Width > imgCfg.Width {
		// crop width out of bounds
		p.Width = imgCfg.Width
	}

	if 0 >= p.Height || p.Height > imgCfg.Height {
		// crop height out of bounds
		p.Height = imgCfg.Height
	}

	p.Width = max(p.Width, 300)
	p.Height = max(p.Height, 300)

	motImage.Seek(0, 0)
	imgData, _, err := image.Decode(motImage)
	if err != nil {
		return "", err
	}

	// grayImg, _ := Grayscale(imgData)

	cc := cutter.Config{
		Width:  p.Width,
		Height: p.Height,
		Anchor: image.Point{
			max(0, p.CenterHorizontal-p.Width/2),
			max(0, p.CenterVertical-p.Height/2),
		},
		Mode: cutter.TopLeft,
	}
	croppedImg, err := cutter.Crop(imgData, cc)
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "prefix-")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	jpeg.Encode(tmpFile, croppedImg, &jpeg.Options{Quality: 80})
	return tmpFile.Name(), nil
}

func Grayscale(img image.Image) (image.Image, error) {
	size := img.Bounds().Size()
	rect := image.Rect(0, 0, size.X, size.Y)
	wImg := image.NewRGBA(rect)

	for x := 0; x < size.X; x++ {
		// and now loop thorough all of this x's y
		for y := 0; y < size.Y; y++ {
			pixel := img.At(x, y)
			originalColor := color.RGBAModel.Convert(pixel).(color.RGBA)
			// Offset colors a little, adjust it to your taste
			r := float64(originalColor.R) * 0.92126
			g := float64(originalColor.G) * 0.97152
			b := float64(originalColor.B) * 0.90722
			// average
			grey := uint8((r + g + b) / 3)
			c := color.RGBA{
				R: grey, G: grey, B: grey, A: originalColor.A,
			}
			wImg.Set(x, y, c)
		}
	}
	return wImg, nil
}

package photo

import (
	"log"
	"testing"
)

func TestCrop(t *testing.T) {
	p := CropParam{
		File:             "buntspecht.jpeg",
		CenterHorizontal: 290,
		CenterVertical:   470,
		Width:            364,
		Height:           792,
	}

	f, err := p.Crop()
	if err != nil {
		t.Error(err)
	}

	log.Println(f)
}

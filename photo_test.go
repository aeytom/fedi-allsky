package main

import (
	"log"
	"testing"
)

func TestCrop(t *testing.T) {
	p := CropParam{
		file:   "buntspecht.jpeg",
		cx:     290,
		cy:     470,
		width:  364,
		height: 792,
	}

	f, err := crop(p)
	if err != nil {
		t.Error(err)
	}

	log.Println(f)
}

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// RunServer …
func RunServer() {

	mux := http.NewServeMux()
	mux.Handle("/photo", logHandler(http.HandlerFunc(htPhoto)))
	mux.Handle("/", logHandler(http.HandlerFunc(htNotify)))
	if err := http.ListenAndServe("127.0.0.1:"+ArgPort, mux); err != nil {
		panic(err)
	}
}

//
func htPhoto(w http.ResponseWriter, req *http.Request) {

	var cropInfo CropParam

	cacheHeader(w)

	camera := "1"
	if val, err := strconv.ParseInt(req.FormValue("c"), 10, 16); err == nil {
		camera = strconv.Itoa(int(val))
	}
	// log.Println("Camera: " + camera)

	if target_dir, err := motionConfigGet(camera, "target_dir"); err == nil {
		// log.Println("- target dir: " + target_dir)
		if file, err := filepath.Abs(filepath.Clean(req.FormValue("file"))); err == nil {
			// log.Println("- image file: " + file)
			if strings.HasPrefix(file, target_dir) {
				// log.Println("- crop file: " + file)
				cropInfo.file = file
			}
		} else {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// http://127.0.0.1:18358/photo?file=%f&w=%i&h=%J&x=%K&y=%L

	if val, err := strconv.ParseInt(req.FormValue("x"), 10, 16); err == nil {
		cropInfo.cx = int(val)
	}
	if val, err := strconv.ParseInt(req.FormValue("y"), 10, 16); err == nil {
		cropInfo.cy = int(val)
	}
	if val, err := strconv.ParseInt(req.FormValue("w"), 10, 16); err == nil {
		cropInfo.width = int(val)
	}
	if val, err := strconv.ParseInt(req.FormValue("h"), 10, 16); err == nil {
		cropInfo.height = int(val)
	}

	log.Println(cropInfo)
	rtext := fmt.Sprintf("Crop and sending image: '%s' center (%d,%d) size (%d,%d)\n",
		cropInfo.file, cropInfo.cx, cropInfo.cy, cropInfo.width, cropInfo.height)
	_, _ = io.WriteString(w, rtext)

	if imgFile, err := crop(cropInfo); err == nil {
		rtext += fmt.Sprintf("result image: '%s'\n", imgFile)
		log.Println(imgFile)
		defer os.Remove(imgFile)
		rtext = telegramSendCroppedImage(imgFile)
	} else {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err := io.WriteString(w, rtext)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//
func htNotify(w http.ResponseWriter, req *http.Request) {

	var notice []string

	var photoPath = req.FormValue("photo")
	if stat, err := os.Stat(photoPath); err == nil {
		ArgImage = photoPath
		notice = append(notice, "New photo from "+stat.ModTime().Format(time.UnixDate))
		notice = append(notice, "Use /photo to get this photo.")
	}

	if msg := req.FormValue("msg"); msg != "" {
		notice = append(notice, msg)
	}

	rtext := notifyTelegram(strings.Join(notice, "\n"))

	cacheHeader(w)
	_, err := io.WriteString(w, rtext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// logHandler …
func logHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.String()
		log.Printf("%s \"%s\" %s \"%s\"", r.Method, u, r.Proto, r.UserAgent())
		h.ServeHTTP(w, r)
	}
}

// cacheHeader …
func cacheHeader(w http.ResponseWriter) {
	w.Header().Add("Cache-Control", "must-revalidate, private, max-age=20")
}

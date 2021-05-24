package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// RunServer …
func RunServer() {

	mux := http.NewServeMux()
	mux.Handle("/", logHandler(http.HandlerFunc(htNotify)))
	if err := http.ListenAndServe("127.0.0.1:"+ArgPort, mux); err != nil {
		panic(err)
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

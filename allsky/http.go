package allsky

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aeytom/fedilib"
	"github.com/mattn/go-mastodon"
)

// ListenMotionWebhook …
func (s *Config) ListenMotionWebhook(mcfg fedilib.Toot) {

	s.toot = mcfg

	mux := http.NewServeMux()
	mux.Handle("/notify", logHandler(http.HandlerFunc(s.htNotify)))
	if err := http.ListenAndServe(s.ListenHost+":"+fmt.Sprint(s.ListenPort), mux); err != nil {
		panic(err)
	}
}

// htNotify handles http://127.0.0.1:18888/?msg=Bewegung+Event+%v+center(%K,%L)
// defined in motioneye "Motion Notifications" field "Web Hook URL"
func (s *Config) htNotify(w http.ResponseWriter, req *http.Request) {

	cacheHeader(w)

	p := ParseRequest(req)

	if p.as_starcount < (s.MinStarCount / 4) {
		writeResponseError("too less stars: "+fmt.Sprint(p.as_starcount), nil, http.StatusOK, w)
		return
	}

	if p.as_starcount > s.MinStarCount {
		// clear sky
		s.tootImage(p, w)
		return
	}

	if p.as_25544visible && p.as_25544alt > s.MinIssAlititude {
		// ISS above 30°
		s.tootImage(p, w)
		return
	}

	if p.as_meteorcount > s.MinMeteorCount {
		// meteors ?
		s.tootImage(p, w)
		return
	}

	writeResponseError("no interesting objects and not enough stars: "+fmt.Sprint(p.as_starcount), nil, http.StatusOK, w)
}

func (s *Config) tootImage(p *AllskyParams, w http.ResponseWriter) bool {

	if !touchLockFile(86400 / 2) {
		writeResponseError("no post to often", nil, http.StatusTooEarly, w)
		return true
	}

	status := "New #allsky image from " + p.as_date + "-" + p.as_time + "\n\n" +
		"Sunset 🌇:  " + p.as_sun_sunset.Format(time.DateTime) + "\n" +
		"Sunrise 🌅: " + p.as_sun_sunrise.Format(time.DateTime) + "\n" +
		"\nSky\n" +
		fmt.Sprintf("Star count:   %d\n", p.as_starcount) +
		fmt.Sprintf("Meteor count: %d\n", p.as_meteorcount) +
		fmt.Sprintf("ISS altitude: %.1f°\n", p.as_25544alt) +
		"\nCamera\n" +
		fmt.Sprintf("Gain:         %d\n", p.as_gain) +
		fmt.Sprintf("Exposure:     %v\n", p.as_exposure_us) +
		fmt.Sprintf("Temperature:  %d°C (sensor)\n", p.as_temperature_c) +
		"\n" +
		s.PublicUrl + "\n"

	toot := mastodon.Toot{
		Status:     status,
		Visibility: mastodon.VisibilityPublic,
		Language:   "en",
	}

	if ifile, err := s.Image(p.date_name, filepath.Base(p.current_image)); err != nil {
		writeResponseError("read allsky image", err, http.StatusInternalServerError, w)
		return true
	} else if err := s.toot.TootWithImageReader(toot, ifile, "Allsky Image"); err != nil {
		writeResponseError("post toot", err, http.StatusInternalServerError, w)
		return true
	}
	return false
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

func writeResponseError(msg string, err error, code int, w http.ResponseWriter) {
	if err != nil {
		msg += " :: " + err.Error()
	}
	log.Println(msg)
	http.Error(w, msg, code)
}

func touchLockFile(minage int) bool {
	fileName := os.TempDir() + "/temp.txt"
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		return true
	}
	if info.ModTime().Add(time.Second * time.Duration(minage)).Before(time.Now()) {
		currentTime := time.Now().Local()
		err = os.Chtimes(fileName, currentTime, currentTime)
		if err != nil {
			fmt.Println(err)
		}
		return true
	}
	return false
}

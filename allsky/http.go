package allsky

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aeytom/fedi-allsky/dump1090"
	"github.com/aeytom/fedilib"
)

type ExtraValue struct {
	Value  string `json:"value"`
	Expire int16  `json:"expire,omitempty"`
}

type ExtraVisibleFlights struct {
	Count ExtraValue `json:"FLIGHT_COUNT,omitempty"`
	List  ExtraValue `json:"FLIGHT_LIST,omitempty"`
}

// ListenAllskyHttp …
func (s *Config) ListenAllskyHttp(mcfg fedilib.Toot) {

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

	vf := s.dump1090.FlightsVisible(time.Now().Add(-40*time.Second), time.Now(), 25)
	s.writeExtraVisibleFlights(vf)

	var p *AllskyParams
	if pr, err := ParseRequest(req); err == nil {
		p = pr
	} else {
		writeResponseError("parameter parse error", err, http.StatusBadRequest, w)
		return
	}
	s.dbStore(p)

	if p.as_starcount < (s.MinStarCount / 4) {
		writeResponseError("too less stars: "+fmt.Sprint(p.as_starcount), nil, http.StatusOK, w)
		return
	}

	if p.as_starcount > s.MinStarCount {
		// clear sky
		s.autoTootImage("as_starcount", p, w)
		return
	}

	// if p.as_25544visible && p.as_25544alt > s.MinIssAlititude {
	// 	// ISS above 30°
	// 	s.autoTootImage("as_25544alt", p, w)
	// 	return
	// }

	if p.as_meteorcount > s.MinMeteorCount && len(vf) == 0 {
		// meteors ?
		s.autoTootImage("as_meteorcount", p, w)
		return
	}

	writeResponseError("no interesting objects and not enough stars: "+fmt.Sprint(p.as_starcount), nil, http.StatusOK, w)
}

func (s *Config) writeExtraVisibleFlights(vf []dump1090.VisibleAircraft) {
	if ex, err := os.OpenFile(filepath.Join(s.ExtraPath, "flights.json"), os.O_WRONLY+os.O_CREATE+os.O_TRUNC, 0644); err != nil {
		s.log.Println(err)
	} else {
		defer ex.Close()
		cf := make([]string, len(vf))
		for i, f := range vf {
			cf[i] = fmt.Sprintf("%s (%.1f km, %.1f°)", f.Flight, f.Distance, f.Elevation)
		}
		c := ExtraVisibleFlights{
			Count: ExtraValue{Value: fmt.Sprint(len(vf)), Expire: 60},
			List:  ExtraValue{Value: strings.Join(cf, ", "), Expire: 60},
		}
		if b, err := json.Marshal(c); err != nil {
			s.log.Println(err)
		} else if _, err := ex.Write(b); err != nil {
			s.log.Println(err)
		}
	}
}

func (s *Config) autoTootImage(key string, p *AllskyParams, w http.ResponseWriter) {

	if !s.lastActionBefore(key, 16*time.Hour) {
		writeResponseError("do not post to often for "+key, nil, http.StatusOK, w)
		return
	}

	err := s.tootAllskyParams(p, nil)
	if err != nil {
		writeResponseError("toot allsky params", err, http.StatusInternalServerError, w)
	} else {
		s.markActionTime(key)
		s.dbExpire(p.date_name)
		writeResponseError("allsky image posted", nil, http.StatusOK, w)
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

func writeResponseError(msg string, err error, code int, w http.ResponseWriter) {
	if err != nil {
		msg += " :: " + err.Error()
	}
	log.Println(msg)
	http.Error(w, msg, code)
}

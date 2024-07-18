package allsky

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type AllskyParams struct {
	as_25544alt      float64       // =-45deg 38' 33.7"
	as_25544visible  bool          // =No
	as_date          string        // =20240714
	as_exposure_us   time.Duration // =30000000
	as_gain          int64         // =255
	as_meteorcount   int64         // =0
	as_starcount     int64         // =958
	as_sun_sunrise   time.Time     // =20240714
	as_sun_sunset    time.Time     // =20240713
	as_temperature_c int64         // =32
	as_time          string        // =015605
	current_image    string        // =/home/tay/allsky/tmp/image-20240714015605.jpg
	date_name        string        // =20240714
}

func ParseRequest(req *http.Request) (p *AllskyParams, err error) {
	as_25544alt := req.FormValue("AS_25544ALT")           // =-45deg 38' 33.7"
	as_25544visible := req.FormValue("AS_25544VISIBLE")   // =No
	as_date := req.FormValue("AS_DATE")                   // =20240714
	as_exposure_us := req.FormValue("AS_EXPOSURE_US")     // =30000000
	as_gain := req.FormValue("AS_GAIN")                   // =255
	as_meteorcount := req.FormValue("AS_METEORCOUNT")     // =0
	as_starcount := req.FormValue("AS_STARCOUNT")         // =958
	as_sun_sunrise := req.FormValue("AS_SUN_SUNRISE")     // =20240714
	as_sun_sunset := req.FormValue("AS_SUN_SUNSET")       // =20240713
	as_temperature_c := req.FormValue("AS_TEMPERATURE_C") // =32
	as_time := req.FormValue("AS_TIME")                   // =015605
	current_image := req.FormValue("CURRENT_IMAGE")       // =/home/tay/allsky/tmp/image-20240714015605.jpg
	date_name := req.FormValue("DATE_NAME")               // =20240714

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("catch param parse error: ", r)
			err = errors.New(fmt.Sprint(r))
		}
	}()

	p = &AllskyParams{
		as_25544alt:      mustParseAngle(as_25544alt),
		as_25544visible:  mustParseBool(as_25544visible),
		as_date:          as_date,
		as_exposure_us:   mustParseDuration(as_exposure_us + "us"),
		as_gain:          mustParseInt(as_gain),
		as_meteorcount:   mustParseInt(as_meteorcount),
		as_starcount:     mustParseInt(as_starcount),
		as_sun_sunrise:   mustParseDateTime(as_sun_sunrise),
		as_sun_sunset:    mustParseDateTime(as_sun_sunset),
		as_temperature_c: mustParseInt(as_temperature_c),
		as_time:          as_time,
		current_image:    current_image,
		date_name:        date_name,
	}

	return p, err
}

func mustParseAngle(val string) float64 {
	var angle float64
	re := regexp.MustCompile("^(-)?(\\d{1,3}(?:\\.\\d+)?)(?:deg|°)(?: (\\d{1,2})'(?: (\\d{1,2}(?:\\.\\d+)?)\")?)?$")
	if !re.MatchString(val) {
		return angle
	}
	matches := re.FindStringSubmatch(val)
	if deg, err := strconv.ParseFloat(matches[2], 64); err != nil {
		log.Panic(err, val)
		return angle
	} else {
		angle += deg
	}
	if matches[3] != "" {
		if min, err := strconv.ParseInt(matches[3], 10, 16); err != nil {
			log.Panic(err, val)
			return angle
		} else {
			angle += float64(min) / 60.0
		}
	}
	if matches[4] != "" {
		if sec, err := strconv.ParseFloat(matches[4], 64); err != nil {
			log.Panic(err, val)
			return angle
		} else {
			angle += sec / 3600.0
		}
	}
	if matches[1] == "-" {
		return 0 - angle
	} else {
		return angle
	}
}

func mustParseInt(val string) int64 {
	if i, err := strconv.ParseInt(val, 10, 64); err != nil {
		log.Panic("not an integer: "+val, err)
		return 0
	} else {
		return i
	}
}

func mustParseBool(val string) bool {
	if val == "Yes" {
		return true
	} else {
		return false
	}
}

func mustParseDateTime(val string) time.Time {
	if loc, err := time.LoadLocation("Europe/Berlin"); err != nil {
		log.Panic(val, err)
	} else if t, err := time.ParseInLocation("20060102 15:04:05", val, loc); err != nil {
		log.Panic(val, err)
	} else {
		return t
	}
	return time.Time{}
}

func mustParseDuration(val string) time.Duration {
	if t, err := time.ParseDuration(val); err != nil {
		log.Panic(val, err)
	} else {
		return t
	}
	return 0
}

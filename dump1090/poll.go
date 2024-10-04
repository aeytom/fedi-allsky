package dump1090

import (
	"github.com/jftuga/geodist"

	"database/sql"
	"log"
	"math"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const FEED_KM = 3280.8399
const AIRCRAFT_JSON = "http://adsb.fritz.box/dump1090/data/aircraft.json"
const STATION_LAT = 52.284623
const STATION_LON = 13.207523

type Config struct {
	Url     string        `json:"url,omitempty"`
	Station geodist.Coord `json:"station,omitempty"`
	Dsn     string        `json:"dsn,omitempty"`

	log *log.Logger
	db  *sql.DB
}

func Init(d *Config, log *log.Logger) *Config {
	d.log = log

	if d.Url == "" {
		d.Url = AIRCRAFT_JSON
	}

	if d.Dsn == "" {
		d.Dsn = "aircraft.db?mode=rwc&_journal=wal"
	}

	if d.Station.Lon == 0 {
		d.Station = geodist.Coord{Lat: STATION_LAT, Lon: STATION_LON}
	}

	if handle, err := sql.Open("sqlite3", d.Dsn); err != nil {
		d.log.Fatal(err)
	} else {
		d.db = handle
		d.dbCreate()
	}

	return d
}

func (d *Config) Close() {
	d.db.Close()
}

func (d *Config) Tick(ticker time.Ticker) {
	go func() {
		for {
			t := <-ticker.C
			d.log.Println("call dump1090 api at ", t.Format(time.RFC3339))
			d.poll()
			d.dbExpire(t.Add(-24 * time.Hour))
		}
	}()
}

func (d *Config) dbCreate() error {
	sqlStmt := `CREATE TABLE IF NOT EXISTS "dump1090" (
		"Now" TEXT NOT NULL,
		"Hex" TEXT NOT NULL,
		"Lat" NUMERIC NOT NULL,
		"Lon" NUMERIC NOT NULL,
		"Flight" TEXT NOT NULL,
		"Track" NUMERIC NOT NULL,
		"Speed" NUMERIC NOT NULL,
		"Altitude" NUMERIC NOT NULL,
		"Distance" NUMERIC NOT NULL,
		"Angle" NUMERIC NOT NULL,
		PRIMARY KEY("Now","Hex")
	);`
	if _, err := d.db.Exec(sqlStmt); err != nil {
		return err
	}

	sqlStmt = `CREATE INDEX "Hex" ON "dump1090" (
		"Hex" ASC
	)`
	if _, err := d.db.Exec(sqlStmt); err != nil {
		return err
	}

	return nil
}

func (s *Config) dbExpire(before time.Time) error {
	sdel := "DELETE FROM `dump1090` WHERE `Now` < ?"
	_, err := s.db.Exec(sdel, before.Add(-4*time.Hour))
	return err
}

func (d *Config) poll() {
	if r, err := GetReport(d.Url); err != nil {
		log.Print(err)
	} else {
		now := time.Unix(int64(r.Now), int64(time.Second)*int64(r.Now-math.Floor(r.Now))).UTC().Format(time.RFC3339Nano)
		for _, a := range r.Aircraft {
			if a.Flight == nil {
				continue
			}
			if a.Lat == nil || a.Lon == nil {
				continue
			}
			ac := geodist.Coord{Lat: *a.Lat, Lon: *a.Lon}
			_, distance := geodist.HaversineDistance(d.Station, ac)
			elevation := math.Pi / 2
			if distance > 0 {
				elevation = math.Atan(*a.Altitude / FEED_KM / distance)
			}
			if elevation > degToRad(15) {
				d.log.Printf("… see flight %s in %.1f km elevation angle %.1f°\n", *a.Flight, distance, radToDeg(elevation))
				sql := "INSERT OR IGNORE INTO `dump1090` (`Now`,`Hex`,`Lat`,`Lon`,`Flight`,`Track`,`Speed`,`Altitude`,`Distance`,`Angle`) VALUES (?,?,?,?,?,?,?,?,?,?)"
				if _, err := d.db.Exec(
					sql,
					now,
					a.Hex,
					a.Lat,
					a.Lon,
					a.Flight,
					a.Track,
					a.Speed,
					a.Altitude,
					distance,
					elevation,
				); err != nil {
					d.log.Println(err)
				}
			}
		}
	}
}

func radToDeg(rad float64) float64 {
	return rad * 180 / math.Pi
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

type VisibleAircraft struct {
	Flight    string
	Distance  float64
	Elevation float64
}

func (d *Config) FlightsVisible(from time.Time, to time.Time, minelevation float64) []VisibleAircraft {
	out := make([]VisibleAircraft, 0)
	if r, err := d.db.Query(
		"SELECT Flight,Distance,Angle FROM `dump1090` WHERE `Now` > ? AND `Now` < ? AND `Angle` > ? GROUP BY `Flight` ORDER BY `Angle` DESC LIMIT 10",
		from.UTC().Format(time.RFC3339Nano),
		to.UTC().Format(time.RFC3339Nano),
		degToRad(minelevation),
	); err != nil {
		d.log.Println(err)
	} else {
		for r.Next() {
			f := VisibleAircraft{}
			var rad float64
			if err = r.Scan(&f.Flight, &f.Distance, &rad); err != nil {
				d.log.Println(err)
				continue
			} else {
				f.Elevation = radToDeg(rad)
				out = append(out, f)
			}
		}
	}
	return out
}

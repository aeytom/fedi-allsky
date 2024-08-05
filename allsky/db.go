package allsky

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const columns = "`as_25544alt`,`as_25544visible`,`as_date`,`as_exposure_us`,`as_gain`,`as_meteorcount`,`as_starcount`,`as_sun_sunrise`,`as_sun_sunset`,`as_temperature_c`,`as_time`,`current_image`,`date_name`"

type Entry struct {
	Id             int64
	Lfdnr          string
	Start          time.Time
	Thema          string
	Von            string
	Bis            string
	Plz            string
	StrasseNr      string
	Aufzugsstrecke string
}

func (s *Config) DbOpen(dsn string) error {
	s.log.Print("sqllite3 DSN ", dsn)
	if handle, err := sql.Open("sqlite3", dsn); err != nil {
		return err
	} else {
		s.db = handle
		s.dbCreate()
	}
	return nil
}

func (s *Config) DbClose() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Config) Db() *sql.DB {
	return s.db
}

func (s *Config) dbCreate() error {
	sqlStmt := `CREATE TABLE IF NOT EXISTS "allsky" (
		"as_25544alt"	TEXT NOT NULL,
		"as_25544visible"	TEXT NOT NULL,
		"as_date"	TEXT NOT NULL,
		"as_exposure_us"	NUMERIC NOT NULL,
		"as_gain"	NUMERIC NOT NULL,
		"as_meteorcount"	NUMERIC NOT NULL,
		"as_starcount"	NUMERIC NOT NULL,
		"as_sun_sunrise"	TEXT NOT NULL,
		"as_sun_sunset"	TEXT NOT NULL,
		"as_temperature_c"	NUMERIC NOT NULL,
		"as_time"	TEXT NOT NULL,
		"current_image"	TEXT NOT NULL,
		"date_name"	TEXT NOT NULL,
		PRIMARY KEY("current_image")
	);`
	if _, err := s.db.Exec(sqlStmt); err != nil {
		return err
	}

	sqlStmt = `CREATE INDEX "date_name" ON "allsky" (
		"date_name"	DESC
	)`
	if _, err := s.db.Exec(sqlStmt); err != nil {
		return err
	}

	sqlStmt = `CREATE TABLE IF NOT EXISTS "sent" (
		"key"	TEXT NOT NULL,
		"ts"	TEXT NOT NULL,
		PRIMARY KEY("key")
	)`
	if _, err := s.db.Exec(sqlStmt); err != nil {
		return err
	}

	return nil
}

func (s *Config) dbExpire(date_name string) error {
	sdel := "DELETE FROM `allsky` WHERE `date_name` < ?"
	_, err := s.db.Exec(sdel, date_name)
	return err
}

func (s *Config) lastActionBefore(key string, age time.Duration) bool {
	ts := time.Now().Add(0 - age)
	sql := "SELECT `ts` FROM `sent` WHERE `ts` > ? LIMIT 1"
	row := s.db.QueryRow(sql, ts.UTC().Format(time.RFC3339Nano))
	var tsdb string
	if err := row.Scan(&tsdb); err != nil {
		s.log.Println(err)
		return true
	}
	if ts, err := time.Parse(time.RFC3339Nano, tsdb); err != nil {
		s.log.Println(err)
		return true
	} else {
		s.log.Println("last action ", key, ts.Format(time.RFC3339))
		return false
	}
}

func (s *Config) markActionTime(key string) error {
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	sql := "INSERT INTO `sent`(`key`,`ts`) VALUES(?,?) ON CONFLICT(`key`) DO UPDATE SET `ts` = ?"
	_, err := s.db.Exec(sql, key, ts, ts)
	return err
}

func (s *Config) dbStore(p *AllskyParams) bool {
	sql := "INSERT OR IGNORE INTO `allsky` (" + columns + ") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)"
	if rslt, err := s.db.Exec(
		sql,
		p.as_25544alt,
		p.as_25544visible,
		p.as_date,
		p.as_exposure_us,
		p.as_gain,
		p.as_meteorcount,
		p.as_starcount,
		p.as_sun_sunrise.Format(time.RFC3339),
		p.as_sun_sunset.Format(time.RFC3339),
		p.as_temperature_c,
		p.as_time,
		p.current_image,
		p.date_name,
	); err != nil {
		s.log.Println(p.current_image, err)
	} else if ra, err := rslt.RowsAffected(); err != nil {
		s.log.Println(p.current_image, err)
	} else {
		return ra > 0
	}
	return false
}

func (s *Config) dbGetDateName() string {
	sql := "SELECT " + columns + " FROM `allsky` ORDER BY `date_name` DESC LIMIT 1"
	row := s.db.QueryRow(sql)
	if p, err := s.dbRow(row); err != nil {
		return ""
	} else {
		return p.date_name
	}
}

func (s *Config) dbGetBestStarcount(date_name string) (*AllskyParams, error) {
	sql := "SELECT " + columns + " FROM `allsky` WHERE `date_name` = ? ORDER BY `as_starcount` DESC LIMIT 1"
	row := s.db.QueryRow(sql, date_name)
	return s.dbRow(row)
}

func (s *Config) dbGetBestMeteors(date_name string) (*AllskyParams, error) {
	sql := "SELECT " + columns + " FROM `allsky` WHERE `date_name` = ? AND `as_starcount` > ? AND `as_meteorcount` > 0 ORDER BY `as_meteorcount` DESC LIMIT 1"
	row := s.db.QueryRow(sql, date_name, s.MinStarCount/3)
	return s.dbRow(row)
}

func (s *Config) dbListMeteors(date_name string) ([]*AllskyParams, error) {
	sql := "SELECT " + columns + " FROM `allsky` WHERE `date_name` = ? AND `as_starcount` > ? AND `as_meteorcount` > 0 AND `as_meteorcount` < 8 ORDER BY (`as_starcount` *  `as_starcount` / `as_meteorcount`) DESC LIMIT 9"
	if rows, err := s.db.Query(sql, date_name, s.MinStarCount/4); err != nil {
		return nil, err
	} else {
		pp := make([]*AllskyParams, 0)
		defer rows.Close()
		for rows.Next() {
			p := AllskyParams{}
			var as_sun_sunrise, as_sun_sunset string
			if err := rows.Scan(&p.as_25544alt, &p.as_25544visible, &p.as_date, &p.as_exposure_us, &p.as_gain, &p.as_meteorcount, &p.as_starcount, &as_sun_sunrise, &as_sun_sunset, &p.as_temperature_c, &p.as_time, &p.current_image, &p.date_name); err != nil {
				return nil, err
			} else {
				var err error
				if p.as_sun_sunrise, err = time.Parse(time.RFC3339, as_sun_sunrise); err != nil {
					s.log.Println(as_sun_sunrise, err)
					continue
				}
				if p.as_sun_sunset, err = time.Parse(time.RFC3339, as_sun_sunset); err != nil {
					s.log.Println(as_sun_sunset, err)
					continue
				}
				pp = append(pp, &p)
			}
		}
		return pp, nil
	}
}

func (s *Config) dbGetBestIss(date_name string) (*AllskyParams, error) {
	sql := "SELECT " + columns + " FROM `allsky` WHERE `date_name` = ? AND `as_starcount` > ? AND `as_25544visible` > 0 ORDER BY `as_25544alt` DESC LIMIT 1"
	row := s.db.QueryRow(sql, date_name, s.MinStarCount/3)
	return s.dbRow(row)
}

func (s *Config) dbRow(row *sql.Row) (*AllskyParams, error) {
	p := AllskyParams{}
	var as_sun_sunrise, as_sun_sunset string
	if err := row.Scan(&p.as_25544alt, &p.as_25544visible, &p.as_date, &p.as_exposure_us, &p.as_gain, &p.as_meteorcount, &p.as_starcount, &as_sun_sunrise, &as_sun_sunset, &p.as_temperature_c, &p.as_time, &p.current_image, &p.date_name); err != nil {
		s.log.Println(err)
		return nil, err
	}
	var err error
	if p.as_sun_sunrise, err = time.Parse(time.RFC3339, as_sun_sunrise); err != nil {
		s.log.Println(as_sun_sunrise, err)
		return nil, err
	}
	if p.as_sun_sunset, err = time.Parse(time.RFC3339, as_sun_sunset); err != nil {
		s.log.Println(as_sun_sunset, err)
		return nil, err
	}
	return &p, err
}

package dump1090

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

func GetReport(url string) (Report, error) {
	var report Report

	client := http.DefaultClient
	client.Timeout = 5 * time.Second
	resp, err := client.Get(url)
	if err != nil {
		return report, err
		// log.Fatal("Unable to get data ", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return report, err
		// log.Fatal("Unable to read body ", err)
	}

	err = json.Unmarshal(body, &report)
	if err != nil {
		return report, err
		// log.Fatal("Unable to parse body ", err)
	}

	/*
	 * Clean up the flight ID by removing leading and trailing spaces
	 */
	for i, a := range report.Aircraft {
		if a.Flight != nil {
			trimmed := strings.TrimSpace(*a.Flight)
			report.Aircraft[i].Flight = &trimmed
		}

		//fmt.Println(reflect.TypeOf(a.alt_baro))
	}

	return report, nil
}

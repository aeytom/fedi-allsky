package dump1090

import "encoding/json"

type Report struct {
	Now      float64    `json:"now"`
	Messages uint64     `json:"messages"`
	Aircraft []Aircraft `json:"aircraft"`
}

type Aircraft struct {
	Hex      string   `json:"hex"`
	Squawk   *string  `json:"squawk,omitempty"`
	Lat      *float64 `json:"lat,omitempty"`
	Lon      *float64 `json:"lon,omitempty"`
	Flight   *string  `json:"flight,omitempty"`
	Speed    *float64 `json:"speed,omitempty"`
	Track    *float64 `json:"track,omitempty"`
	Rssi     *float32 `json:"rssi,omitempty"`
	Altitude *float64 `json:"altitude,omitempty"`

	/*
	 * This field might be a number, a string (usually "ground"), or nil
	 */
	BarometerAltitude json.Token `json:"alt_baro,omitempty"`

	/*
	 * These fields are NOT part of the dump1090 format, but they may
	 * be added by external decorators
	 */
	Registration *string
	Description  *string
	AircraftType *string
}

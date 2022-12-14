package issDataCollector

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TimezoneLocation struct {
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	TimezoneID  string `json:"timezone_id"`
	Offset      int    `json:"offset"`
	CountryCode string `json:"country_code"`
	MapURL      string `json:"map_url"`
}

type ISSLocation struct {
	Name       string  `json:"name"`
	ID         int     `json:"id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Altitude   float64 `json:"altitude"`
	Velocity   float64 `json:"velocity"`
	Visibility string  `json:"visibility"`
	Footprint  float64 `json:"footprint"`
	Timestamp  int     `json:"timestamp"`
	Daynum     float64 `json:"daynum"`
	SolarLat   float64 `json:"solar_lat"`
	SolarLon   float64 `json:"solar_lon"`
	Units      string  `json:"units"`
}

type ISSData struct {
	ID          int `gorm:"type:BigInt;primaryKey;autoIncrement"`
	Name        string
	Latitude    float64
	Longitude   float64
	Altitude    float64
	Velocity    float64
	Visibility  string
	Footprint   float64
	Timestamp   int
	Daynum      float64
	SolarLat    float64
	SolarLon    float64
	Units       string
	TimezoneID  string
	Offset      int
	CountryCode string
}

func ISSHandler(w http.ResponseWriter, r *http.Request) {
	data, err := getISSLocation()
	if err != nil {
		log.Printf("error while getting iss data: {}\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "error"}`))
		return
	}
	tzLocation, err := getTimezoneLocation(data.Latitude, data.Longitude)
	if err != nil {
		log.Printf("error while getting timezone from location: {}\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "error"}`))
		return
	}
	err = saveData(data, tzLocation)
	if err != nil {
		log.Printf("error while saving data: {}\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "error"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}

func getISSLocation() (ISSLocation, error) {
	res, err := http.Get("https://api.wheretheiss.at/v1/satellites/25544")
	if err != nil {
		return ISSLocation{}, err
	}
	if res.StatusCode != 200 {
		return ISSLocation{}, fmt.Errorf("error: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ISSLocation{}, err
	}
	data := ISSLocation{}
	json.Unmarshal(body, &data)
	return data, nil
}

func getTimezoneLocation(latitude float64, longitude float64) (TimezoneLocation, error) {
	url := fmt.Sprintf("https://api.wheretheiss.at/v1/coordinates/%f,%f", latitude, longitude)

	res, err := http.Get(url)
	if err != nil {
		return TimezoneLocation{}, err
	}
	if res.StatusCode != 200 {
		return TimezoneLocation{}, fmt.Errorf("error: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return TimezoneLocation{}, err
	}
	data := TimezoneLocation{}
	json.Unmarshal(body, &data)
	return data, nil
}

func saveData(issData ISSLocation, tzData TimezoneLocation) error {
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		return err
	}
	data := ISSData{
		Name:        issData.Name,
		Latitude:    issData.Latitude,
		Longitude:   issData.Longitude,
		Altitude:    issData.Altitude,
		Velocity:    issData.Velocity,
		Visibility:  issData.Visibility,
		Footprint:   issData.Footprint,
		Timestamp:   issData.Timestamp,
		Daynum:      issData.Daynum,
		SolarLat:    issData.SolarLat,
		SolarLon:    issData.SolarLon,
		Units:       issData.Units,
		TimezoneID:  tzData.TimezoneID,
		Offset:      tzData.Offset,
		CountryCode: tzData.CountryCode,
	}
	if err = db.Create(&data).Error; err != nil {
		return err
	}
	return nil
}

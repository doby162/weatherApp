package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type PointData struct {
	Properties struct {
		Forecast string `json:"forecast"`
	} `json:"properties"`
}

type PeriodData struct {
	Name                       string `json:"name"`
	Temperature                int    `json:"temperature"`
	TemperatureUnit            string `json:"temperatureUnit"`
	ProbabilityOfPrecipitation struct {
		Value int `json:"value"`
	} `json:"probabilityOfPrecipitation"`
	WindSpeed        string `json:"windSpeed"`
	DetailedForecast string `json:"detailedForecast"`
}

type ForecastData struct {
	Properties struct {
		Periods []PeriodData `json:"periods"`
	} `json:"properties"`
}

func main() {
	mux := http.NewServeMux()
	client := &http.Client{Timeout: 10 * time.Second}
	email := os.Getenv("EMAIL")
	log.Println("using " + email + " as contact email for weather.gov")
	userAgent := fmt.Sprintf("(Michael's Weather App, %s)", email)

	mux.HandleFunc("GET /forecast/{lat}/{lon}", func(w http.ResponseWriter, r *http.Request) {
		lat := r.PathValue("lat")
		lon := r.PathValue("lon")
		log.Println("request for ", lat, lon)
		valid := validateLatLon(lat, lon)
		if !valid {
			log.Println("invalid lat/lon")
			http.Error(w, "invalid lat/lon", http.StatusBadRequest)
			return
		}

		pointRequest, err := http.NewRequest("GET", fmt.Sprintf("https://api.weather.gov/points/%s,%s", lat, lon), nil)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to construct point request", http.StatusBadGateway)
			return
		}

		pointRequest.Header.Set("User-Agent", userAgent)

		pointResponse, err := client.Do(pointRequest)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to fetch point data", http.StatusBadGateway)
			return
		}
		defer pointResponse.Body.Close()

		if pointResponse.StatusCode != 200 {
			log.Println("Error: ", pointResponse.Status)
			http.Error(w, pointResponse.Status, http.StatusInternalServerError)
			return
		}

		var pointData PointData
		err = json.NewDecoder(pointResponse.Body).Decode(&pointData)
		if err != nil {
			log.Println("parsing error: ", err)
			http.Error(w, "failed to parse point data", http.StatusInternalServerError)
			return
		}

		if !strings.HasPrefix(pointData.Properties.Forecast, "https://api.weather.gov/") {
			log.Println("invalid forecast URL")
			http.Error(w, "invalid forecast URL", http.StatusInternalServerError)
			return
		}
		forecastRequest, err := http.NewRequest("GET", pointData.Properties.Forecast, nil)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to construct forecast request", http.StatusInternalServerError)
			return
		}
		forecastRequest.Header.Set("User-Agent", userAgent)

		forecastResponse, err := client.Do(forecastRequest)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to fetch forecast data", http.StatusInternalServerError)
			return
		}
		defer forecastResponse.Body.Close()

		if forecastResponse.StatusCode != 200 {
			log.Println("Error: ", forecastResponse.Status)
			http.Error(w, forecastResponse.Status, http.StatusInternalServerError)
			return
		}

		var forecastData ForecastData
		err = json.NewDecoder(forecastResponse.Body).Decode(&forecastData)
		if err != nil {
			log.Println("parsing error: ", err)
			http.Error(w, "failed to parse forecast data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		for _, period := range forecastData.Properties.Periods {
			_, err := fmt.Fprintf(w, "%s\n\n The temperature is %d%s, which in my opinion is %s\n\n %s\n\n----------------------------------\n\n",
				period.Name,
				period.Temperature,
				period.TemperatureUnit,
				temperatureOpinion(period.Temperature),
				period.DetailedForecast)
			if err != nil {
				return
			}
		}
	})

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Println("Server error:", err)
		os.Exit(1)
	}
}

func temperatureOpinion(temp int) string {
	// TODO parse wind data and apply some kind of modifier
	if temp < 50 {
		return "cold"
	} else if temp <= 80 {
		return "moderate"
	}
	return "hot"
}

func validateLatLon(lat, lon string) bool {
	latFloat, err := strconv.ParseFloat(lat, 64)
	if err != nil || latFloat < -90 || latFloat > 90 {
		return false
	}

	lonFloat, err := strconv.ParseFloat(lon, 64)
	if err != nil || lonFloat < -180 || lonFloat > 180 {
		return false
	}
	return true
}

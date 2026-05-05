package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	client := &http.Client{}
	email := os.Getenv("EMAIL")
	fmt.Println("using " + email + " as contact email for weather.gov")
	userAgent := fmt.Sprintf("(Michael's Weather App, %s)", email)

	mux.HandleFunc("GET /forecast/{lat}/{lon}", func(w http.ResponseWriter, r *http.Request) {
		lat := r.PathValue("lat")
		lon := r.PathValue("lon")

		pointRequest, err := http.NewRequest("GET", fmt.Sprintf("https://api.weather.gov/points/%s,%s", lat, lon), nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		pointRequest.Header.Set("User-Agent", userAgent)

		pointResponse, err := client.Do(pointRequest)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(pointResponse.Body)

		if pointResponse.StatusCode != 200 {
			fmt.Println("Error: ", pointResponse.Status)
			return
		}

		var pointData PointData
		err = json.NewDecoder(pointResponse.Body).Decode(&pointData)
		if err != nil {
			fmt.Println("parsing error: ", err)
			return
		}

		forecastRequest, err := http.NewRequest("GET", pointData.Properties.Forecast, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		forecastRequest.Header.Set("User-Agent", userAgent)

		forecastResponse, err := client.Do(forecastRequest)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(forecastResponse.Body)

		if forecastResponse.StatusCode != 200 {
			fmt.Println("Error: ", forecastResponse.Status)
			return
		}

		var forecastData ForecastData
		err = json.NewDecoder(forecastResponse.Body).Decode(&forecastData)

		for _, period := range forecastData.Properties.Periods {
			_, err := fmt.Fprintf(w, "%s\n\n The temperature is %d%s, which is my opinion is %s\n\n %s\n\n----------------------------------\n\n",
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

	fmt.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
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

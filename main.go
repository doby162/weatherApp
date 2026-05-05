package main

import (
	"encoding/json"
	"fmt"
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
		}

		pointRequest.Header.Set("User-Agent", userAgent)

		pointResponse, err := client.Do(pointRequest)
		if err != nil {
			fmt.Println(err)
		}
		defer pointResponse.Body.Close()

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

		forecastRequest, _ := http.NewRequest("GET", pointData.Properties.Forecast, nil)
		forecastRequest.Header.Set("User-Agent", userAgent)

		forecastResponse, err := client.Do(forecastRequest)
		if err != nil {
			fmt.Println(err)
		}
		defer forecastResponse.Body.Close()

		if forecastResponse.StatusCode != 200 {
			fmt.Println("Error: ", forecastResponse.Status)
			return
		}

		var forecastData ForecastData
		err = json.NewDecoder(forecastResponse.Body).Decode(&forecastData)

		fmt.Fprintf(w, "Viewing forecast with coords: %s, %s", lat, lon)
		fmt.Println(pointData.Properties.Forecast)

		for _, period := range forecastData.Properties.Periods {
			fmt.Fprintf(w, "\n\n%s", period.Name)
			fmt.Fprintf(w, "\n\n The temperature is %d%s, which is my opinion is %s", period.Temperature, period.TemperatureUnit, tempOpinion(period.Temperature))
			fmt.Fprintf(w, "\n\n %s", period.DetailedForecast)
			fmt.Fprintf(w, "\n\n----------------------------------")

			//fmt.Fprintf(w, "\n\n%s: %d %s, %d percent, %s, %s", period.Name, period.Temperature, period.TemperatureUnit, period.ProbabilityOfPrecipitation.Value, period.WindSpeed, period.DetailedForecast)
		}

	})

	fmt.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
	}
}

func tempOpinion(temp int) string {
	// TODO parse wind data and apply some kind of modifier. 60 degrees is warm but 60 and windy is just annoying
	if temp > 100 {
		return "too hot"
	} else if temp > 80 {
		return "hot"
	} else if temp > 60 {
		return "warm"
	} else if temp < 0 {
		return "Chicago chilly"
	} else if temp < 40 {
		return "chilly"
	} else if temp < 61 {
		return "sweater weather"
	} else {
		return "a temperature"
	}
}

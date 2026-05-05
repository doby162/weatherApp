package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	client := &http.Client{}
	email := os.Getenv("EMAIL")
	fmt.Println("using " + email + " as contact email for weather.gov")

	mux.HandleFunc("GET /forecast/{lat}/{lon}", func(w http.ResponseWriter, r *http.Request) {
		lat := r.PathValue("lat")
		lon := r.PathValue("lon")

		req, _ := http.NewRequest("GET", "https://example.com", nil)

		req.Header.Set("User-Agent", fmt.Sprintf("(Michael's Weather App, %s)", email))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Body)

		fmt.Fprintf(w, "Viewing forecast with coords: %s, %s", lat, lon)
	})

	fmt.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
	}
}

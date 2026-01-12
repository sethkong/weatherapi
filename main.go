package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const baseUrl = "https://api.weather.gov"

type WeatherHandler struct{}
type Forecast struct {
	Number                     int    `json:"number"`
	Name                       string `json:"name"`
	StartTime                  string `json:"startTime"`
	EndTime                    string `json:"endTime"`
	IsDaytime                  bool   `json:"isDaytime"`
	Temperature                int    `json:"temperature"`
	TemperatureUnit            string `json:"temperatureUnit"`
	TemperatureTrend           any    `json:"temperatureTrend"`
	WindSpeed                  string `json:"windSpeed"`
	WindDirection              string `json:"windDirection"`
	Icon                       string `json:"icon"`
	ShortForecast              string `json:"shortForecast"`
	DetailedForecast           string `json:"detailedForecast"`
	ProbabilityOfPrecipitation any    `json:"probabilityOfPrecipitation"`
}

func weatherRoutes() chi.Router {
	router := chi.NewRouter()
	weatherHandler := WeatherHandler{}
	router.Get("/{latitude}/{longitude}", weatherHandler.GetWeatherForecast)
	return router
}

func callWeatherService(endpoint string) ([]byte, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func MeaureTemperature(temp int) string {
	if temp >= 80 {
		return "hot"
	} else if temp <= 60 {
		return "cold"
	} else {
		return "moderate"
	}
}

func (wh *WeatherHandler) GetWeatherForecast(w http.ResponseWriter, r *http.Request) {
	latitude := chi.URLParam(r, "latitude")
	longitude := chi.URLParam(r, "longitude")
	endpoint := fmt.Sprintf("%s/points/%s,%s", baseUrl, latitude, longitude)
	body, err := callWeatherService(endpoint)
	if err != nil {
		handleHttpError(w, err)
		return
	}
	var jsonBody map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &jsonBody); err != nil {
		handleHttpError(w, err)
		return
	}
	var properties map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonBody["properties"]), &properties); err != nil {
		handleHttpError(w, err)
		return
	}
	var forecastEndpoint string
	if err := json.Unmarshal([]byte(properties["forecast"]), &forecastEndpoint); err != nil {
		handleHttpError(w, err)
		return
	}
	forecastBody, err := callWeatherService(forecastEndpoint)
	if err != nil {
		handleHttpError(w, err)
		return
	}
	var jsonForcastBody map[string]json.RawMessage
	if err := json.Unmarshal([]byte(forecastBody), &jsonForcastBody); err != nil {
		handleHttpError(w, err)
		return
	}
	var forecastProperties map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonForcastBody["properties"]), &forecastProperties); err != nil {
		handleHttpError(w, err)
		return
	}
	var forecasts []Forecast
	if err := json.Unmarshal([]byte(forecastProperties["periods"]), &forecasts); err != nil {
		handleHttpError(w, err)
		return
	}
	if len(forecasts) == 0 {
		handleHttpError(w, fmt.Errorf("no forecasts found"))
		return
	}
	temperature := MeaureTemperature(forecasts[0].Temperature)
	var response = map[string]string{
		"shortForecast": forecasts[0].ShortForecast,
		"temperature":   temperature,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleHttpError(w, err)
	}
}

func handleHttpError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})
	router.Mount("/weather", weatherRoutes())
	http.ListenAndServe(":4000", router)
}

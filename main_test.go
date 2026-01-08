package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockRoundTripper struct {
	fn func(req *http.Request) *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req), nil
}

func TestMeaureTemperature(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{input: 85, expected: "hot"},
		{input: 55, expected: "cold"},
		{input: 70, expected: "moderate"},
	}
	for _, test := range tests {
		result := MeaureTemperature(test.input)
		if result != test.expected {
			t.Errorf("MeaureTemperature(%d) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestGetWeatherForecast_Success(t *testing.T) {
	// Save real transport and restore after test
	realTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = realTransport
	}()

	// Mock HTTP responses
	http.DefaultTransport = &mockRoundTripper{
		fn: func(req *http.Request) *http.Response {

			switch req.URL.String() {

			case "https://api.weather.gov/points/40,70":
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`
					{
					  "properties": {
						"forecast": "https://api.weather.gov/forecast"
					  }
					}`)),
					Header: make(http.Header),
				}

			case "https://api.weather.gov/forecast":
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`
					{
					  "properties": {
						"periods": [
						  {
							"temperature": 85,
							"shortForecast": "Partly Cloudy"
						  }
						]
					  }
					}`)),
					Header: make(http.Header),
				}
			}

			t.Fatalf("unexpected request to %s", req.URL.String())
			return nil
		},
	}

	// Setup router
	r := chi.NewRouter()
	r.Mount("/", weatherRoutes())

	req := httptest.NewRequest(http.MethodGet, "/40/70", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Assertions
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["shortForecast"] != "Partly Cloudy" {
		t.Errorf("expected 'Partly Cloudy', got %s", resp["shortForecast"])
	}

	if resp["temperature"] != "hot" {
		t.Errorf("expected 'hot', got %s", resp["temperature"])
	}
}

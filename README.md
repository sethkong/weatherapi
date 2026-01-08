# weatherapi

## Requirements

1. Accepts latitude and longitude coordinates
2. Returns the short forecast for that area for Today (“Partly Cloudy” etc)
3. Returns a characterization of whether the temperature is “hot”, “cold”, or “moderate”
	  (use your discretion on mapping temperatures to each type)
4. Use the National Weather Service API Web Service as a data source.

## Usage

Starts the weather REST api

```cmd
cd weatherapp
go run main.go
```

In another terminal invoke the `weather` endpoint by providing the `latitude` and `longitude` parameters.

```cmd
curl localhost:4000/weather/35.2271/-80.8431
```
Or using a web browser: `http://localhost:4000/weather/35.2271/-80.8431`
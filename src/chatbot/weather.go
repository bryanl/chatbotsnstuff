package chatbot

import (
	"encoding/json"
	"net/http"
	"net/url"
)

var (
	// WeatherAPIKey is the weather api key.
	WeatherAPIKey string
)

type weatherResp struct {
	WeatherFields []weatherWeatherResp `json:"weather"`
	Main          weatherMainResp      `json:"main"`
}

type weatherWeatherResp struct {
	Description string `json:"description"`
}

type weatherMainResp struct {
	Temp float64 `json:"temp"`
}

func weatherByZip(zip string) (*weatherResp, error) {
	u := url.URL{
		Scheme: "http",
		Host:   "api.openweathermap.org",
		Path:   "/data/2.5/weather",
	}

	v := u.Query()
	v.Set("zip", zip+",us")
	v.Set("APPID", WeatherAPIKey)
	v.Set("units", "imperial")

	u.RawQuery = v.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var wr weatherResp
	err = json.NewDecoder(resp.Body).Decode(&wr)
	if err != nil {
		return nil, err
	}

	return &wr, nil
}

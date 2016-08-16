package chatbot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var (
	// WeatherAPIKey is the weather api key.
	WeatherAPIKey string
)

func weatherState(fields []string) state {
	return func(e Event) state {
		if len(fields) != 2 {
			e.Gateway.Tell(Destination(e.Creator), "usage: *!weather <zip>*")
			return nil
		}

		wr, err := weatherByZip(fields[1])
		if err != nil {
			return errorState(err)
		}

		msg := fmt.Sprintf("It is currently %02.fF in %s: %s\n",
			wr.Main.Temp, fields[1], wr.WeatherFields[0].Description)
		e.Gateway.Tell(Destination(e.Creator), msg)

		return nil
	}
}

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

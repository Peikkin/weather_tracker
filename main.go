package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ApiConfigData struct {
	OpenWeatherMapApiKey string `json:"OpenWeatherMapApiKey"`
}

type WeatherData struct {
	Name    string `json:"name"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Celsius  float64 `json:"temp"`
		Feels    float64 `json:"feels_like"`
		Pressure float64 `json:"pressure"`
		Humidity float64 `json:"humidity"`
		TempMin  float64 `json:"temp_min"`
		TempMax  float64 `json:"temp_max"`
		SeaLevel float64 `json:"sea_level"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All float64 `json:"all"`
	} `json:"clouds"`
	Sys struct {
		Sunrise int64 `json:"sunrise"`
		Sunset  int64 `json:"sunset"`
	} `json:"sys"`
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	http.HandleFunc("/hello", Hello)
	http.HandleFunc("/weather/",
		func(w http.ResponseWriter, r *http.Request) {
			city := strings.SplitN(r.URL.Path, "/", 3)[2]
			data, err := query(city)
			if err != nil {
				log.Error().Err(err).Msg("Ошибка получения данных")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			// json.NewEncoder(w).Encode(data)

			jsonData, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				log.Error().Err(err).Msg("Ошибка кодирования JSON ответа")
			}

			w.Write(jsonData)
		})

	log.Info().Msg("Запуск сервера на порту :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("Ошибка запуска сервера")
	}
}

func LoadApiConfig(filename string) (ApiConfigData, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка чтения конфига")
		return ApiConfigData{}, err
	}

	var config ApiConfigData

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка кодирования конфига")
		return ApiConfigData{}, err
	}
	return config, nil
}

func Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!"))
}

func query(city string) (WeatherData, error) {
	apiConfig, err := LoadApiConfig(".apiConfig")
	if err != nil {
		log.Fatal().Err(err).Msg("Ошибка загрузки ApiKey")
	}
	resp, err := http.Get("https://api.openweathermap.org/data/2.5/weather?units=metric&q=" + city + "&appid=" + apiConfig.OpenWeatherMapApiKey)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка получения ответа")
		return WeatherData{}, err
	}
	defer resp.Body.Close()

	var data WeatherData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка декодирования ответа")
		return WeatherData{}, err
	}

	return data, nil
}

package main

import (
	"chatbot"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

const (
	botName = "BOT"
)

type specification struct {
	WeatherAPIKey string `envconfig:"weather_api_key" required:"true"`
}

func main() {
	var s specification
	err := envconfig.Process("chatbot", &s)
	if err != nil {
		logrus.WithError(err).Fatal("unable to parse configuration")
	}

	chatbot.WeatherAPIKey = s.WeatherAPIKey

	gw := chatbot.NewLocalGateway(botName)

	errChan := make(chan error)

	go func() {
		for err := range errChan {
			logrus.WithError(err).Error("chatbot failure")
		}
	}()

	cb := chatbot.New(gw)
	go cb.Start(errChan)

	done := make(chan bool)
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		gw.Stop()

		done <- true
	}()

	logrus.Info("bot booted")
	<-done
}

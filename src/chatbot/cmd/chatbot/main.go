package main

import (
	"chatbot"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

type specification struct {
	WeatherAPIKey string `envconfig:"weather_api_key" required:"true"`
	BotName       string `envconfig:"bot_name" default:"TwikiTheBot"`
	IRCChan       string `envconfig:"irc_chan" required:"true"`
}

func main() {
	var s specification
	err := envconfig.Process("chatbot", &s)
	if err != nil {
		logrus.WithError(err).Fatal("unable to parse configuration")
	}

	chatbot.WeatherAPIKey = s.WeatherAPIKey

	errChan := make(chan error)

	go func() {
		for err := range errChan {
			logrus.WithError(err).Error("chatbot failure")
		}
	}()

	localGw := initLocalGW(&s, errChan)
	ircGw := initIRCGW(&s, errChan)

	cb := chatbot.New(localGw, ircGw)
	go cb.Start(errChan)

	done := make(chan bool)
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cb.Stop()

		done <- true
	}()

	logrus.Info("bot booted")
	<-done
}

func initLocalGW(s *specification, errChan chan error) chatbot.Gateway {
	return chatbot.NewLocalGateway(s.BotName)
}

func initIRCGW(s *specification, errChan chan error) chatbot.Gateway {
	return chatbot.NewIRCGateway(s.BotName, s.IRCChan)
}

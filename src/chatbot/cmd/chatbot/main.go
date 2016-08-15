package main

import (
	"chatbot"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
)

func main() {
	gw := chatbot.NewLocalGateway()

	errChan := make(chan error)

	go func() {
		for err := range errChan {
			logrus.WithError(err).Error("local chat gateway failure")
		}
	}()

	go gw.Start(errChan)

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

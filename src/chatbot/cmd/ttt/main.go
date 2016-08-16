package main

import (
	"chatbot"

	"github.com/Sirupsen/logrus"
)

func main() {
	ttt := chatbot.NewTicTacToe()

	board, status, err := ttt.Play(1)
	if err != nil {
		logrus.WithError(err).Error("play failed")
	}

	logrus.WithFields(logrus.Fields{
		"board":  board,
		"status": status,
	}).Info("play results")
}

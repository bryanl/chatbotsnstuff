package chatbot

import (
	"strings"

	"github.com/Sirupsen/logrus"
)

type state func(e Event) state

func unknownState(fields []string) state {
	return func(e Event) state {
		msg := strings.Join(fields, " ")
		logrus.WithField("command", msg).Info("unknown state")
		e.Gateway.Tell(Destination(e.Creator), "unknown command: "+msg)
		return nil
	}
}

func errorState(err error) state {
	logrus.WithError(err).Error("error state")
	return nil
}

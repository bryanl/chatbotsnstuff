package chatbot

import (
	"fmt"
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

func weatherSate(fields []string) state {
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

func errorState(err error) state {
	logrus.WithError(err).Error("error state")
	return nil
}

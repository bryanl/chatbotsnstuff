package chatbot

import (
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
)

// SlackGateway is a gateway for chatting over slack.
type SlackGateway struct {
	api     *slack.Client
	botName string
	botChan string
	logger  *logrus.Entry
	events  chan Event
}

var _ Gateway = (*SlackGateway)(nil)

// NewSlackGateway creates an instance of SlackGateway.
func NewSlackGateway(slackToken, botName, botChan string) *SlackGateway {
	logger := logrus.WithFields(logrus.Fields{
		"gateway": "slack",
		"botName": botName,
	})

	return &SlackGateway{
		api:     slack.New(slackToken),
		botName: botName,
		botChan: botChan,
		logger:  logger,
		events:  make(chan Event),
	}
}

// Events are events from the SlackGateway.
func (g *SlackGateway) Events() <-chan Event {
	return g.events
}

// Start starts the slack gateway.
func (g *SlackGateway) Start(errChan chan error) {
	rtm := g.api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			g.logger.Info("connected to slack")

		case *slack.MessageEvent:
			g.events <- Event{
				Type:    MessageEvent,
				Creator: ev.Channel,
				Payload: ev.Text,
				Gateway: g,
			}
		}
	}
}

// Stop the slack gateway.
func (g *SlackGateway) Stop() {

}

// Tell sends a message to a destination.
func (g *SlackGateway) Tell(dest Destination, msg string) error {
	params := slack.PostMessageParameters{
		Username: g.botName,
	}
	_, _, err := g.api.PostMessage(string(dest), msg, params)
	return err
}

// Display displays an image.
func (g *SlackGateway) Display(dest Destination, imageData io.Reader) error {
	return nil
}

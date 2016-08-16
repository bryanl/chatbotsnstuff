//go:generate stringer -type=EventType

package chatbot

import (
	"io"

	"github.com/Sirupsen/logrus"
)

// Destination is where a message will be displayed.
type Destination string

// Event is a bot event.
type Event struct {
	Type    EventType
	Gateway Gateway
	Creator string
	Payload interface{}
}

// EventType is an event type.
type EventType int

const (
	// AddEvent denotes that a client has been added.
	AddEvent EventType = iota
	// MessageEvent denotes a client has sent a message.
	MessageEvent
)

// Gateway is a Chatbot's interface to the world.
type Gateway interface {
	Start(errChan chan error)
	Stop()
	Tell(dest Destination, msg string) error
	Display(dest Destination, imageData io.Reader) error
	Events() <-chan Event
}

// Chatbot is a chatbot.
type Chatbot struct {
	gateways  []Gateway
	brain     *brain
	eventChan chan Event
	logger    *logrus.Entry
}

// New creates an instance of Chatbot.
func New(gateways ...Gateway) *Chatbot {
	return &Chatbot{
		brain:    newBrain(),
		gateways: gateways,
		logger:   logrus.WithField("chatbot", "main"),
	}
}

// Start starts the chatbot.
func (c *Chatbot) Start(errChan chan error) {
	c.eventChan = make(chan Event, 10)

	go func() {
		for event := range c.eventChan {
			c.logger.WithField("event", event).Info("received event")

			switch event.Type {
			case MessageEvent:
				s := c.brain.Parse(event.Payload.(string))

				for s != nil {
					s = s(event)
				}
			}
		}
	}()

	for _, gw := range c.gateways {
		go gw.Start(errChan)
		go func(currentGw Gateway) {
			for event := range currentGw.Events() {
				c.eventChan <- event
			}
		}(gw)

	}
}

// Stop stops the chatbot.
func (c *Chatbot) Stop() {
	for _, gw := range c.gateways {
		gw.Stop()
	}

	close(c.eventChan)
}

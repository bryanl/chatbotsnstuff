package chatbot

import (
	"crypto/tls"
	"io"
	"strings"

	"github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
)

// IRCGateway is a gateway for chatting via IRC.
type IRCGateway struct {
	botName string
	botChan string
	logger  *logrus.Entry
	events  chan Event
	conn    *irc.Conn
}

var _ Gateway = (*IRCGateway)(nil)

// NewIRCGateway creates an instance of IRCGateway.
func NewIRCGateway(botName, botChan string) *IRCGateway {
	logger := logrus.StandardLogger().WithFields(logrus.Fields{
		"gateway": "irc",
		"botName": botName,
		"botChan": botChan,
	})

	return &IRCGateway{
		botName: botName,
		botChan: botChan,
		logger:  logger,
		events:  make(chan Event),
	}
}

// Events are events from the IRCGateway.
func (g *IRCGateway) Events() <-chan Event {
	return g.events
}

// Start starts the irc gateway.
func (g *IRCGateway) Start(errChan chan error) {
	cfg := irc.NewConfig(g.botName)
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: "irc.freenode.net"}
	cfg.Server = "irc.freenode.net:7000"
	cfg.NewNick = func(n string) string { return n + "^" }

	g.conn = irc.Client(cfg)

	g.conn.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			conn.Join(g.botChan)
			g.logger.Info("joined channel")
		})

	g.conn.HandleFunc(irc.PRIVMSG,
		func(conn *irc.Conn, line *irc.Line) {
			if line.Args[0] == g.botChan {
				g.logger.Info("sending event")
				g.events <- Event{
					Gateway: g,
					Type:    MessageEvent,
					Creator: line.Args[0],
					Payload: strings.Join(line.Args[1:], " "),
				}
				g.logger.Info("sent event")
			}
		})

	// Tell client to connect.
	g.logger.Info("connecting to irc")
	if err := g.conn.Connect(); err != nil {
		g.logger.WithError(err).Error("connection failure")
	}
}

// Stop the irc gateway.
func (g *IRCGateway) Stop() {
	g.logger.Info("shutting down")

	if g.conn != nil {
		g.conn.Quit("dying")
	}
}

// Tell sends a message to a destination.
func (g *IRCGateway) Tell(dest Destination, msg string) error {
	g.conn.Privmsg(string(dest), msg)
	return nil
}

// Display displays an image.
func (g *IRCGateway) Display(dest Destination, imageData io.Reader) error {
	g.conn.Privmsg(string(dest), "one day i'll upload an image")
	return nil
}

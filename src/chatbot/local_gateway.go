package chatbot

import (
	"bufio"
	"io"
	"net"
	"strings"

	"github.com/Sirupsen/logrus"
)

const (
	connBufSize = 4096
)

type localMessage struct {
	sender   net.Conn
	userName string
	msg      string
}

type chatChans struct {
	msg  chan localMessage
	add  chan localClient
	rm   chan localClient
	stop chan bool
}

func newChatChans() *chatChans {
	return &chatChans{
		msg:  make(chan localMessage),
		add:  make(chan localClient),
		rm:   make(chan localClient),
		stop: make(chan bool),
	}
}

type localClient struct {
	conn net.Conn
	ch   chan<- string
}

func (l *localClient) Write(msg string) error {
	_, err := l.conn.Write([]byte(msg))
	return err
}

// LocalGateway is a gateway for chatting locally. It takes input from
// stdin, and talks to stdout.
type LocalGateway struct {
	botName  string
	logger   *logrus.Entry
	cc       *chatChans
	doneChan chan struct{}
}

// NewLocalGateway creates an instance of LocalGateway.
func NewLocalGateway(botName string) *LocalGateway {
	logger := logrus.StandardLogger().WithFields(logrus.Fields{
		"gateway": "local",
		"botName": botName,
	})

	return &LocalGateway{
		botName: botName,
		logger:  logger,
	}
}

// Start starts the local gateway.
func (g *LocalGateway) Start(errChan chan error) {
	listener, err := net.Listen("tcp", ":8889")
	if err != nil {
		errChan <- err
		return
	}

	g.logger.WithField("addr", listener.Addr()).Info("starting listener")

	defer func() {
		g.logger.Info("shutting down listener")
	}()

	g.cc = newChatChans()

	go g.handleMessages(g.cc)

	g.doneChan = make(chan struct{})

	go func() {
		<-g.doneChan
		if err := listener.Close(); err != nil {
			g.logger.WithError(err).Error("listener close failure")
		}

	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			errChan <- err
		}

		go g.handleConnection(conn, g.cc)
	}
}

// Stop the local gateway.
func (g *LocalGateway) Stop() {
	g.logger.Info("shutting down")
	if g.cc != nil {
		g.cc.stop <- true
	}

	if g.doneChan != nil {
		g.doneChan <- struct{}{}
	}
}

func (g *LocalGateway) handleMessages(cc *chatChans) {
	clients := make(map[net.Conn]chan<- string)
	for {
		select {
		case msg := <-cc.msg:
			g.logger.WithFields(logrus.Fields{
				"out":      strings.TrimSpace(msg.msg),
				"userName": msg.userName,
			}).Info("sending message")

			for conn, ch := range clients {
				if conn != msg.sender {
					go func(userName, out string, mch chan<- string) {
						mch <- (userName + ": " + out)
					}(msg.userName, msg.msg, ch)
				}
			}
		case client := <-cc.add:
			clients[client.conn] = client.ch
		case client := <-cc.rm:
			delete(clients, client.conn)
		case <-cc.stop:
			for k := range clients {
				if err := k.Close(); err != nil {
					g.logger.WithError(err).Error("could not close connection")
				}
			}
		}
	}
}

func (g *LocalGateway) handleConnection(conn net.Conn, cc *chatChans) {
	ch := make(chan string)

	lc := localClient{
		conn: conn,
		ch:   ch,
	}

	cc.add <- lc

	defer func() {
		conn.Close()
		cc.rm <- lc
	}()

	go func() {
		for msg := range ch {
			lc.Write(msg)
		}
	}()

	buf := make([]byte, connBufSize)

	lc.Write("Username?: ")
	r := bufio.NewReader(conn)
	userName, err := r.ReadString('\n')
	if err != nil {
		lc.Write("invalid username\n")
		return
	}
	userName = strings.TrimSpace(userName)

	g.logger.WithField("userName", userName).Info("new connection")

	g.Tell(Destination(userName), "hello "+userName+"\n")

	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			g.logger.WithError(err).Error("could not read input")
			break
		}

		cc.msg <- localMessage{
			userName: userName,
			sender:   conn,
			msg:      string(buf[0:n]),
		}
	}
}

// Tell sends a message to a destination.
func (g *LocalGateway) Tell(dest Destination, msg string) error {
	g.cc.msg <- localMessage{
		userName: g.botName,
		msg:      msg,
	}

	return nil
}

// Display displays an image.
func (g *LocalGateway) Display(dest Destination, imageData io.Reader) error {
	g.cc.msg <- localMessage{
		userName: "BOT",
		msg:      "copy file to image server\n",
	}

	return nil
}

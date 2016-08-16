package chatbot

import "strings"

const (
	botCommandPrefix = "!"
)

// brain is the chatbot brain.
type brain struct {
}

// newBrain creates a new instance of Brain.
func newBrain() *brain {
	return &brain{}
}

// Parse parses a potential bot command.
func (b *brain) Parse(msg string) state {
	fields := strings.Fields(msg)
	if isBotCommand(fields) {
		command := strings.TrimPrefix(fields[0], botCommandPrefix)
		switch command {
		case "weather":
			return weatherSate(fields)
		default:
			return unknownState(fields)
		}
	}

	return nil
}

func isBotCommand(fields []string) bool {
	return strings.HasPrefix(fields[0], botCommandPrefix)
}

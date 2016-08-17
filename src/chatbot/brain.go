package chatbot

import "strings"

const (
	botCommandPrefix = "!"
)

// brain is the chatbot brain.
type brain struct {
	greyMatter map[string]interface{}
}

// newBrain creates a new instance of Brain.
func newBrain() *brain {
	return &brain{
		greyMatter: map[string]interface{}{},
	}
}

// Parse parses a potential bot command.
func (b *brain) Parse(who, msg string) state {
	if msg == "" {
		return unknownState([]string{})
	}
	fields := strings.Fields(msg)
	if isBotCommand(fields) {
		command := strings.TrimPrefix(fields[0], botCommandPrefix)
		switch command {
		case "weather":
			return weatherState(fields)
		case "tictactoe":
			// !tictactoe
			// !tictactoe play tl <- places an x on the top left spot

			if len(fields) == 1 {
				return startTicTacToeState(who, b.greyMatter["tictactoe"])
			} else if len(fields) == 3 {
				return playTicTacToeState(fields, who, b.greyMatter["tictactoe"])
			}

			return unknownState(fields)
		default:
			return unknownState(fields)
		}
	}

	return nil
}

func isBotCommand(fields []string) bool {
	return strings.HasPrefix(fields[0], botCommandPrefix)
}

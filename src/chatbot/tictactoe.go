//go:generate stringer -type=TicTacToeStatus

package chatbot

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

const (
	playerPiece = "o"
)

var (
	ticTacToeRows = []string{"top", "middle", "bottom"}
	ticTacToeCols = []string{"left", "center", "right"}
)

type ticTacToeScoreboard map[string]*TicTacToe

func startTicTacToeState(who string, greyMatter interface{}) state {
	if greyMatter == nil {
		greyMatter = ticTacToeScoreboard{}
	}

	m := greyMatter.(ticTacToeScoreboard)
	if m[who] == nil {
		m[who] = NewTicTacToe()
	}

	logrus.WithFields(logrus.Fields{
		"gm":  greyMatter,
		"who": who,
	}).Info("start tic tac toe")

	return nil
}

func playTicTacToeState(fields []string, who string, greyMatter interface{}) state {
	logrus.WithFields(logrus.Fields{
		"who": who,
		"gm":  greyMatter,
	}).Info("play tic tac toe")
	return nil
}

// TicTacToeBoard is a tic tac toe board.
type TicTacToeBoard struct {
	cells [9]string
}

// NewTicTacToeBoard creates an instance of TicTacToeBoard.
func NewTicTacToeBoard() TicTacToeBoard {
	return TicTacToeBoard{
		cells: [9]string{},
	}
}

func (t *TicTacToeBoard) String() string {
	rows := []string{}

	for i := 0; i < 3; i++ {
		row := []string{}
		row = append(row, t.cells[i*3+0], t.cells[i*3+1], t.cells[i*3+2])
		rows = append(rows, strings.Join(row, ","))
	}

	return strings.Join(rows, " | ")
}

func convertTicTacRespToBoard(resp ticTacToeResp) TicTacToeBoard {
	posMap := map[string]int{}

	i := 0
	for _, row := range ticTacToeRows {
		for _, col := range ticTacToeCols {
			posMap[row+"-"+col] = i
			i++
		}
	}

	board := NewTicTacToeBoard()

	for _, cell := range resp.Data.Board {
		pos := posMap[cell.ID]
		board.cells[pos] = cell.Value
	}

	return board
}

type ticTacToeCell struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type ticTacToeReq struct {
	PlayerPiece   string          `json:"player_piece"`
	OpponentPiece string          `json:"opponent_piece"`
	Board         []ticTacToeCell `json:"board"`
}

type ticTacToeResp struct {
	Status string       `json:"status"`
	Data   ticTacToeReq `json:"data"`
}

// TicTacToeStatus is the status of the TicTacToe game.
type TicTacToeStatus int

const (
	// TicTacToeStatusGo denotes the game is ongoing
	TicTacToeStatusGo TicTacToeStatus = iota
	// TicTacToeStatusDraw denotes the game is over and a draw
	TicTacToeStatusDraw
	// TicTacToeStatusLose denotes the game is over and a loss
	TicTacToeStatusLose
	// TicTacToeStatusError denotes the game is an an error state
	TicTacToeStatusError
)

func (t *ticTacToeResp) GameStatus() TicTacToeStatus {
	switch t.Status {
	case "draw":
		return TicTacToeStatusDraw
	case "win":
		return TicTacToeStatusLose
	default:
		return TicTacToeStatusGo
	}
}

// TicTacToe is a tic tac toe game
type TicTacToe struct {
	board      TicTacToeBoard
	isGameOver bool
}

// NewTicTacToe creates a new instance of tic tac toe.
func NewTicTacToe() *TicTacToe {
	return &TicTacToe{board: NewTicTacToeBoard()}
}

// Play plays a position and keeps track of status
func (t *TicTacToe) Play(pos int) (*TicTacToeBoard, TicTacToeStatus, error) {
	var err error

	defer func() {
		if err != nil {
			t.isGameOver = true
		}
	}()

	if t.isGameOver {
		return nil, TicTacToeStatusError, errors.New("game is over")
	}

	if t.board.cells[pos] != "" {
		return nil, TicTacToeStatusError, errors.New("position has already been filled")
	}

	t.board.cells[pos] = playerPiece

	req := ticTacToeReq{
		PlayerPiece:   "x",
		OpponentPiece: playerPiece,
		Board:         []ticTacToeCell{},
	}

	i := 0
	for _, row := range ticTacToeRows {
		for _, col := range ticTacToeCols {
			cell := ticTacToeCell{
				ID:    row + "-" + col,
				Value: t.board.cells[i],
			}
			req.Board = append(req.Board, cell)
			i++
		}
	}

	resp, err := t.makeReq(req)
	if err != nil {
		return nil, TicTacToeStatusError, err
	}

	board := convertTicTacRespToBoard(*resp)
	return &board, resp.GameStatus(), err
}

func (t *TicTacToe) makeReq(tttReq ticTacToeReq) (*ticTacToeResp, error) {
	j, err := json.Marshal(&tttReq)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse("http://perfecttictactoe.herokuapp.com/")
	if err != nil {
		return nil, err
	}

	u.Path = "/api/v2/play"

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var tttResp ticTacToeResp
	if err := json.NewDecoder(resp.Body).Decode(&tttResp); err != nil {
		return nil, err
	}

	return &tttResp, nil
}

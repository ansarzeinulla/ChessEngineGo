package main

import (
	"bufio"
	"os"
	"github.com/notnil/chess"
	"fmt"
	"strings"
)



type Engine struct {
	game *chess.Game
}

func NewEngine() *Engine {
	return &Engine{game: chess.NewGame()}
}

// === UCI Engine Core ===

func (e *Engine) HandleInput(input string) {
	switch {
	case input == "uci":
		fmt.Println("id name AlphaBetaEngine")
		fmt.Println("id author You")
		fmt.Println("uciok")
	case input == "isready":
		fmt.Println("readyok")
	case strings.HasPrefix(input, "position"):
		e.setPosition(input)
	case input[:2] == "go":
		e.makeMove()
	case input == "quit":
		os.Exit(0)
	}
	os.Stdout.Sync()
}

func (e *Engine) setPosition(cmd string) {
	tokens := strings.Fields(cmd)
	if len(tokens) < 2 {
		e.game = chess.NewGame()
		return
	}

	switch tokens[1] {
	case "startpos":
		e.game = chess.NewGame()
	case "fen":
		fenParts := []string{}
		i := 2
		for i < len(tokens) && tokens[i] != "moves" {
			fenParts = append(fenParts, tokens[i])
			i++
		}
		fenStr := strings.Join(fenParts, " ")
		pos, err := chess.FEN(fenStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid FEN:", err)
			e.game = chess.NewGame()
		} else {
			e.game = chess.NewGame(pos)
		}
	}
}


func NewScanner(r *os.File) *Scanner {
	return &Scanner{r: r}
}

type Scanner struct {
	r   *os.File
	buf []byte
}

func (s *Scanner) Scan() bool {
	s.buf = make([]byte, 0, 4096)
	var b [1]byte
	for {
		_, err := s.r.Read(b[:])
		if err != nil {
			return false
		}
		if b[0] == '\n' {
			break
		}
		s.buf = append(s.buf, b[0])
	}
	return true
}

func (s *Scanner) Text() string {
	return string(s.buf)
}

func main() {
	engine := NewEngine()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		engine.HandleInput(scanner.Text())
	}
}

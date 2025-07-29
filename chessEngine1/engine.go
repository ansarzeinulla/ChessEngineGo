package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/notnil/chess"
)

type RandomEngine struct {
	game *chess.Game
}

// NewRandomEngine initializes the engine with a fresh game
func NewRandomEngine() *RandomEngine {
	return &RandomEngine{game: chess.NewGame()}
}

// HandleInput routes a single UCI command string
func (e *RandomEngine) HandleInput(input string) {
	switch {
	case input == "uci":
		fmt.Println("id name RandomEngine")
		fmt.Println("id author You")
		fmt.Println("uciok")
	case input == "isready":
		fmt.Println("readyok")
	case strings.HasPrefix(input, "position"):
		e.setPosition(input)
	case input[:2] == "go":
		e.playMove()
	case input == "quit":
		os.Exit(0)
	}
	os.Stdout.Sync()
}

// setPosition handles the "position" command and optionally applies a move list
func (e *RandomEngine) setPosition(command string) {
	tokens := strings.Fields(command)
	e.game = nil

	if len(tokens) < 2 {
		fmt.Fprintln(os.Stderr, "invalid position command")
		e.game = chess.NewGame()
		return
	}

	switch tokens[1] {
	case "startpos":
		e.game = chess.NewGame()
	case "fen":
		// Collect FEN string (6 parts), then optionally parse "moves"
		i := 2
		fenParts := []string{}
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
	default:
		fmt.Fprintln(os.Stderr, "unknown position type:", tokens[1])
		e.game = chess.NewGame()
	}

	if e.game == nil {
		e.game = chess.NewGame()
	}

	// Apply moves (if any)
	// for i, token := range tokens {
	// 	if token == "moves" {
	// 		for _, moveStr := range tokens[i+1:] {
	// 			move := uciToMove(e.game, moveStr)
	// 			if move == nil {
	// 				fmt.Fprintln(os.Stderr, "invalid move:", moveStr)
	// 				continue
	// 			}
	// 			if err := e.game.Move(move); err != nil {
	// 				fmt.Fprintln(os.Stderr, "could not apply move:", moveStr, err)
	// 			}
	// 		}
	// 		break
	// 	}
	// }
}

// playMove selects a random legal move and prints it as the bestmove
func (e *RandomEngine) playMove() {
	if e.game == nil {
		fmt.Fprintln(os.Stderr, "no game initialized")
		fmt.Println("bestmove 0000")
		return
	}

	moves := e.game.ValidMoves()
	if len(moves) == 0 {
		fmt.Println("bestmove 0000")
		return
	}

	rand.Seed(time.Now().UnixNano())
	move := moves[rand.Intn(len(moves))]
	moveStr := move.S1().String() + move.S2().String()
	if move.Promo() != chess.NoPieceType {
		moveStr += strings.ToLower(move.Promo().String())
	}
	fmt.Println("bestmove", moveStr)
	os.Stdout.Sync()
}

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/notnil/chess"
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
	case input == "go":
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

func (e *Engine) makeMove() {
	bestScore := -999999
	var bestMove *chess.Move

	moves := e.game.ValidMoves()
	for _, move := range moves {
		clone := e.game.Clone()
		_ = clone.Move(move)
		score := alphaBeta(clone, 2, -999999, 999999, false, 0)
		if score > bestScore || bestMove == nil {
			bestScore = score
			bestMove = move
		}
	}

	if bestMove == nil {
		fmt.Println("bestmove 0000")
		return
	}

	moveStr := bestMove.S1().String() + bestMove.S2().String()
	if bestMove.Promo() != chess.NoPieceType {
		moveStr += strings.ToLower(bestMove.Promo().String())
	}
	fmt.Println("bestmove", moveStr)
	os.Stdout.Sync()
}

// === Alpha-Beta Pruning ===

func alphaBeta(game *chess.Game, depth, alpha, beta int, maximizing bool, ply int) int {
	if depth == 0 || game.Outcome() != chess.NoOutcome || ply >= 4 {
		return evaluate(game.Position())
	}

	moves := game.ValidMoves()
	if maximizing {
		value := -999999
		for _, move := range moves {
			child := game.Clone()
			_ = child.Move(move)
			nextDepth := adjustedDepth(depth, ply, move)
			score := alphaBeta(child, nextDepth, alpha, beta, false, ply+1)
			value = max(value, score)
			alpha = max(alpha, value)
			if beta <= alpha {
				break
			}
		}
		return value
	} else {
		value := 999999
		for _, move := range moves {
			child := game.Clone()
			_ = child.Move(move)
			nextDepth := adjustedDepth(depth, ply, move)
			score := alphaBeta(child, nextDepth, alpha, beta, true, ply+1)
			value = min(value, score)
			beta = min(beta, value)
			if beta <= alpha {
				break
			}
		}
		return value
	}
}

func adjustedDepth(depth, ply int, move *chess.Move) int {
	if move.HasTag(chess.Capture) || move.HasTag(chess.Check) {
		return depth // keep current depth
	}
	return depth - 1
}

// === Evaluation ===

func evaluate(pos *chess.Position) int {
	score := 0
	board := pos.Board()

	for sq := chess.A1; sq <= chess.H8; sq++ {
		piece := board.Piece(sq)
		if piece == chess.NoPiece {
			continue
		}

		val := pieceValue(piece.Type())
		if piece.Color() == chess.White {
			score += val
		} else {
			score -= val
		}
	}
	return score
}

func pieceValue(t chess.PieceType) int {
	switch t {
	case chess.Pawn:
		return 100
	case chess.Knight, chess.Bishop:
		return 300
	case chess.Rook:
		return 500
	case chess.Queen:
		return 900
	default:
		return 0
	}
}

// === Helpers ===

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

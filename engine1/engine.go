package engine1

import (
	chess "ChessEngineGo/arbiter"
)

type Engine struct{}

// Make sure this matches the interface exactly
func (e *Engine) GetMove(board chess.BoardwithParameters) [2]int {
	return [2]int{0, 1}
}

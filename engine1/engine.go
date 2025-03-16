package engine1

import (
	chess "ChessEngineGo/arbiter"
)

type Engine struct{}

func chessLocationToUint64(notation string) uint64 {
	// Validate input
	if len(notation) != 2 {
		return 0
	}
	col := notation[0]
	row := notation[1]

	// Ensure valid column (a-h) and row (1-8)
	if col < 'a' || col > 'h' || row < '1' || row > '8' {
		return 0
	}

	// Calculate column index (0-7)
	colIndex := col - 'a'

	// Calculate row index (0-7), reverse row numbering (1-8 to 7-0)
	rowIndex := row - '1'

	// Calculate bit position: bit_position = row * 8 + col
	bitPosition := rowIndex*8 + colIndex

	// Set the corresponding bit in uint64 and return
	return 1 << bitPosition
}

// Make sure this matches the interface exactly
func (e *Engine) GetMove(board chess.BoardwithParameters) [3]uint64 {
	return [3]uint64{chessLocationToUint64("e1"), chessLocationToUint64("g1"), 0}
}

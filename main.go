package main

import (
	"fmt"

	// Path to chess.go file
	chess "ChessEngineGo/arbiter"
	engine1 "ChessEngineGo/engine1" // Path to engine1.go file
	engine2 "ChessEngineGo/engine2"
)

func main() {
	// Create two engine instances (dummy engines for now)
	engine1Instance := &engine1.Engine{} // Renamed the variable to avoid conflict
	engine2Instance := &engine2.Engine{} // Renamed the variable to avoid conflict
	fen := "rnb1kbnr/pPp1pppp/3q4/3p4/8/8/PPPP1PPP/RNBQK2R w KQkq - 0 1"

	a := 0
	result := ""
	if a == 0 {
		result = chess.PlayGame(engine1Instance, engine2Instance, fen) // Use renamed variables
	}

	fmt.Println(result)
}

func uint64ToChessLocation(cell uint64) string {
	if cell == 0 {
		return "" // Return an empty string if no cell is selected
	}
	row := 1
	for cell >= 256 {
		cell /= 256
		row++
	}
	// Find the column (divide by 2 until we reach 1)
	col := 0
	for cell > 1 {
		cell /= 2
		col++
	}
	// Convert column index to chess notation (a-h)
	notation := string('a'+col) + fmt.Sprintf("%d", row)
	return notation
}

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
	rowIndex := 8 - (row - '0')

	// Calculate bit position: bit_position = row * 8 + col
	bitPosition := rowIndex*8 + colIndex

	// Set the corresponding bit in uint64 and return
	return 1 << bitPosition
}

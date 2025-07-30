package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/notnil/chess"
)


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

// === Evaluation ===

func evaluate(pos *chess.Position) int {
	score := 0
	board := pos.Board()

	for sq := chess.A1; sq <= chess.H8; sq++ {
		piece := board.Piece(sq)
		if piece == chess.NoPiece {
			continue
		}

		// Evaluate each piece individually
		switch piece.Type() {
		case chess.Pawn:
			score += evaluatePawn(board, sq, piece)
		case chess.Knight:
			score += evaluateKnight(board, sq, piece)
		case chess.Bishop:
			score += evaluateBishop(board, sq, piece)
		case chess.Rook:
			score += evaluateRook(board, sq, piece)
		case chess.Queen:
			score += evaluateQueen(board, sq, piece)
		case chess.King:
			score += evaluateKing(board, sq, piece)
		}
	}

	return score
}

// === Pawn Evaluation ===
func evaluatePawn(board *chess.Board, sq chess.Square, piece chess.Piece) int {
	// Basic value of the pawn
	value := pieceValue(piece.Type())

	// Pawn structure: Isolated pawn penalty or passed pawn bonus
	// For simplicity, we're assuming the pawn's position matters in some cases
	if piece.Color() == chess.White {
		// Example: Pawns on the 7th rank are better
		if sq.Rank() == chess.Rank7 {
			value += 50
		}
	} else {
		// For black pawns, pawns on the 2nd rank are weaker
		if sq.Rank() == chess.Rank2 {
			value -= 50
		}
	}
	return value
}

// === Knight Evaluation ===
func evaluateKnight(board *chess.Board, sq chess.Square, piece chess.Piece) int {
	value := pieceValue(piece.Type())

	// Knights are more valuable in the center (for example)
	if sq.File() > chess.FileD && sq.File() < chess.FileE && sq.Rank() > chess.Rank3 && sq.Rank() < chess.Rank6 {
		value += 50 // Centralized knight bonus
	}

	return value
}

// === Bishop Evaluation ===
func evaluateBishop(board *chess.Board, sq chess.Square, piece chess.Piece) int {
	value := pieceValue(piece.Type())

	// Bishops are more powerful on open boards
	// (i.e., when there are fewer pawns blocking their movement)
	if piece.Color() == chess.White {
		if board.Piece(sq + 8) == chess.NoPiece && board.Piece(sq - 8) == chess.NoPiece {
			value += 30 // Open diagonals bonus
		}
	} else {
		if board.Piece(sq + 8) == chess.NoPiece && board.Piece(sq - 8) == chess.NoPiece {
			value -= 30 // Open diagonals penalty
		}
	}

	return value
}

// === Rook Evaluation ===

func evaluateRook(board *chess.Board, sq chess.Square, piece chess.Piece) int {
	value := pieceValue(piece.Type())

	// Rooks are more valuable on open files
	// (i.e., when there are no pawns on the file)
	if piece.Color() == chess.White {
		// Check if the file is open by scanning through the entire file
		openFile := true
		for rank := chess.Rank1; rank <= chess.Rank8; rank++ {
			// Convert file to int and calculate the square index
			checkSquare := chess.Square(int(sq.File())*8 + int(rank)) // Combine file and rank to form a square
			if board.Piece(checkSquare) != chess.NoPiece {
				openFile = false
				break
			}
		}
		if openFile {
			value += 40 // Rook on open file bonus
		}
	} else {
		// Same logic for black rooks
		openFile := true
		for rank := chess.Rank1; rank <= chess.Rank8; rank++ {
			// Convert file to int and calculate the square index
			checkSquare := chess.Square(int(sq.File())*8 + int(rank)) // Combine file and rank to form a square
			if board.Piece(checkSquare) != chess.NoPiece {
				openFile = false
				break
			}
		}
		if openFile {
			value -= 40 // Rook on open file penalty
		}
	}

	return value
}



// === Queen Evaluation ===
func evaluateQueen(board *chess.Board, sq chess.Square, piece chess.Piece) int {
	value := pieceValue(piece.Type())

	// Queens are powerful in the center
	if sq.File() > chess.FileD && sq.File() < chess.FileE && sq.Rank() > chess.Rank3 && sq.Rank() < chess.Rank6 {
		value += 100 // Queen centralization bonus
	}

	return value
}

// === King Evaluation ===
func evaluateKing(board *chess.Board, sq chess.Square, piece chess.Piece) int {
	value := pieceValue(piece.Type())

	// King safety: Penalize if the king is in the center of the board
	if sq.File() > chess.FileC && sq.File() < chess.FileF && sq.Rank() > chess.Rank3 && sq.Rank() < chess.Rank6 {
		value -= 100 // King in the center penalty
	}

	// King endgame: In the endgame, the king becomes more active, so it's rewarded
	// For simplicity, let's just assume that if depth > 20, it's an endgame phase
	// (You would need to pass this information into the evaluation function or calculate it outside)
	// A simplified way to determine this might be to just check the position of the king
	if piece.Color() == chess.White && sq.Rank() > chess.Rank4 {
		value += 50 // King endgame bonus for white
	} else if piece.Color() == chess.Black && sq.Rank() < chess.Rank5 {
		value -= 50 // King endgame penalty for black
	}

	return value
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

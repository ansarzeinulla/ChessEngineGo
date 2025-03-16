package engine1

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

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

// BoardwithParameters represents the state of a chess board
type BoardwithParameters struct {
	Board          [12]uint64 // Bitboards for each piece type
	TurnOfPlayer   int        // 0 for white, 1 for black
	EnPassantWhite uint64     // Position of en passant square for white
	EnPassantBlack uint64     // Position of en passant square for black
	WhiteCastle    int        // Castling rights for white: 0=none, 1=kingside, 2=queenside, 3=both
	BlackCastle    int        // Castling rights for black: 0=none, 1=kingside, 2=queenside, 3=both
}

// ChessArbiter is the main controller for chess games
type ChessArbiter struct {
	BoardwithParameters BoardwithParameters
}

// Piece indices in the Board array
const (
	WhiteKing = iota
	WhiteQueen
	WhiteRook
	WhiteBishop
	WhiteKnight
	WhitePawn
	BlackKing
	BlackQueen
	BlackRook
	BlackBishop
	BlackKnight
	BlackPawn
)

type ChessEngine interface {
	GetMove(board BoardwithParameters) [3]uint64
}

// IsValidMove checks if a move is valid based on chess rules
func IsValidMove(arbiter *ChessArbiter, move [3]uint64) bool {
	// 1. Get the color of current player
	turnOfPlayer := arbiter.BoardwithParameters.TurnOfPlayer

	// Find the bit positions of the from and to squares
	fromBit := findSetBit(move[0])
	toBit := findSetBit(move[1])
	// Validate that exactly one bit is set in from and to positions
	if countSetBits(move[0]) != 1 || countSetBits(move[1]) != 1 {
		return false
	}

	// 2. Check if the piece at FROM position belongs to the current player
	fromPiece, fromColor := getPieceAtPosition(arbiter, fromBit)
	if fromPiece == -1 || fromColor != turnOfPlayer {
		return false // No piece at FROM or wrong color
	}

	// 3. Check if the TO position doesn't have a piece of the same color
	toPiece, toColor := getPieceAtPosition(arbiter, toBit)
	if toPiece != -1 && toColor == turnOfPlayer {
		return false // Can't capture your own piece
	}

	// 4. Check specific piece movement
	switch fromPiece {
	case WhitePawn, BlackPawn:
		return isValidPawnMove(arbiter, move)

	case WhiteKing, BlackKing:
		return isValidKingMove(arbiter, move)

	case WhiteBishop, BlackBishop:
		return isValidBishopMove(arbiter, move)

	case WhiteRook, BlackRook:
		return isValidRookMove(arbiter, move)

	case WhiteQueen, BlackQueen:
		// Queen moves like bishop or rook
		return isValidBishopMove(arbiter, move) || isValidRookMove(arbiter, move)

	case WhiteKnight, BlackKnight:
		return isValidKnightMove(arbiter, move)
	}

	return false
}

// Helper function to find the position of the set bit in a uint64
func findSetBit(bitmap uint64) int {
	// Returns the index of the set bit (0-63)
	for i := 0; i < 64; i++ {
		if bitmap&(uint64(1)<<i) != 0 {
			return i
		}
	}
	return -1 // No bit is set
}

// Helper function to count the number of set bits in a uint64
func countSetBits(bitmap uint64) int {
	count := 0
	for bitmap != 0 {
		bitmap &= bitmap - 1 // Clear the least significant set bit
		count++
	}
	return count
}

// Helper to get the piece and its color at a position
func getPieceAtPosition(arbiter *ChessArbiter, position int) (int, int) {
	// Position is 0-63, we need to create a bitmask with only that bit set
	bitMask := uint64(1) << position

	// Check white pieces first (0-5 in the Board array)
	for pieceType := WhiteKing; pieceType <= WhitePawn; pieceType++ {
		if arbiter.BoardwithParameters.Board[pieceType]&bitMask != 0 {
			return pieceType, 0 // White piece
		}
	}

	// Check black pieces (6-11 in the Board array)
	for pieceType := BlackKing; pieceType <= BlackPawn; pieceType++ {
		if arbiter.BoardwithParameters.Board[pieceType]&bitMask != 0 {
			return pieceType, 1 // Black piece
		}
	}

	// No piece found
	return -1, -1
}

// Validates if a pawn move is legal
func isValidPawnMove(arbiter *ChessArbiter, move [3]uint64) bool {
	// Get bit positions
	fromPos := findSetBit(move[0])
	toPos := findSetBit(move[1])
	promotionPiece := int(move[2]) // Use int for promotion piece index

	// Get pawn color and determine if it's white or black
	_, color := getPieceAtPosition(arbiter, fromPos)

	// Convert positions to coordinates
	fromRank, fromFile := fromPos/8, fromPos%8
	toRank, toFile := toPos/8, toPos%8

	// Calculate rank and file differences
	fileDiff := abs(toFile - fromFile)
	rankDiff := toRank - fromRank // Note: not using abs() here as direction matters for pawns

	// Different movement rules for white and black pawns
	if color == 0 { // White pawn
		// REGULAR MOVE: Forward 1 square
		if rankDiff == 1 && fileDiff == 0 {
			// Check if destination square is empty
			piece, _ := getPieceAtPosition(arbiter, toPos)
			if piece != -1 {
				return false // Destination square is occupied
			}

			// Check for promotion (white pawn reaches 8th rank)
			if toRank == 7 {
				return isValidPromotion(promotionPiece)
			}

			return true
		}

		// INITIAL MOVE: Forward 2 squares from starting position
		if rankDiff == 2 && fileDiff == 0 && fromRank == 1 {
			// Check if both the next square and destination are empty
			midSquare := (fromRank+1)*8 + fromFile

			piece1, _ := getPieceAtPosition(arbiter, midSquare)
			if piece1 != -1 {
				return false // Square in between is occupied
			}

			piece2, _ := getPieceAtPosition(arbiter, toPos)
			if piece2 != -1 {
				return false // Destination square is occupied
			}

			return true
		}

		// CAPTURE: Diagonal move
		if rankDiff == 1 && fileDiff == 1 {
			// Regular capture - check if destination has an opponent's piece
			piece, pieceColor := getPieceAtPosition(arbiter, toPos)

			// Normal capture
			if piece != -1 && pieceColor == 1 {
				// Check for promotion when capturing
				if toRank == 7 {
					return isValidPromotion(promotionPiece)
				}
				return true
			}

			// EN PASSANT capture - only valid against black pawns' en passant square
			if piece == -1 && move[1] == arbiter.BoardwithParameters.EnPassantBlack {
				// Verify there's actually a black pawn in the correct position to capture
				capturedPawnPos := toPos - 8 // One rank below the en passant square
				capturedPawnBit := uint64(1) << capturedPawnPos
				if arbiter.BoardwithParameters.Board[BlackPawn]&capturedPawnBit != 0 {
					// The square is empty but it's the en passant target square with a capturable pawn
					return true
				}
			}

			return false
		}

		// Any other move is invalid for a white pawn
		return false
	} else { // Black pawn
		// REGULAR MOVE: Forward 1 square
		if rankDiff == -1 && fileDiff == 0 {
			// Check if destination square is empty
			piece, _ := getPieceAtPosition(arbiter, toPos)
			if piece != -1 {
				return false // Destination square is occupied
			}

			// Check for promotion (black pawn reaches 1st rank)
			if toRank == 0 {
				return isValidPromotion(promotionPiece)
			}

			return true
		}

		// INITIAL MOVE: Forward 2 squares from starting position
		if rankDiff == -2 && fileDiff == 0 && fromRank == 6 {
			// Check if both the next square and destination are empty
			midSquare := (fromRank-1)*8 + fromFile

			piece1, _ := getPieceAtPosition(arbiter, midSquare)
			if piece1 != -1 {
				return false // Square in between is occupied
			}

			piece2, _ := getPieceAtPosition(arbiter, toPos)
			if piece2 != -1 {
				return false // Destination square is occupied
			}

			return true
		}

		// CAPTURE: Diagonal move
		if rankDiff == -1 && fileDiff == 1 {
			// Regular capture - check if destination has an opponent's piece
			piece, pieceColor := getPieceAtPosition(arbiter, toPos)

			// Normal capture
			if piece != -1 && pieceColor == 0 {
				// Check for promotion when capturing
				if toRank == 0 {
					return isValidPromotion(promotionPiece)
				}
				return true
			}

			// EN PASSANT capture - only valid against white pawns' en passant square
			if piece == -1 && move[1] == arbiter.BoardwithParameters.EnPassantWhite {
				// Verify there's actually a white pawn in the correct position to capture
				capturedPawnPos := toPos + 8 // One rank above the en passant square
				capturedPawnBit := uint64(1) << capturedPawnPos
				if arbiter.BoardwithParameters.Board[WhitePawn]&capturedPawnBit != 0 {
					// The square is empty but it's the en passant target square with a capturable pawn
					return true
				}
			}

			return false
		}

		// Any other move is invalid for a black pawn
		return false
	}
}

// Helper function to check if promotion is valid
func isValidPromotion(promotionPiece int) bool {
	// Promotion piece cannot be a pawn or king
	if promotionPiece == -1 {
		return false // No promotion piece specified, but we're on the promotion rank
	}

	// Check for valid promotion pieces
	// Can only promote to Queen, Rook, Bishop, or Knight
	validPromotions := []int{
		WhiteQueen, WhiteRook, WhiteBishop, WhiteKnight, // 1, 2, 3, 4
		BlackQueen, BlackRook, BlackBishop, BlackKnight, // 7, 8, 9, 10
	}

	for _, piece := range validPromotions {
		if promotionPiece == piece {
			return true
		}
	}

	return false
}

// Validates if a king's move is legal
func isValidKingMove(arbiter *ChessArbiter, move [3]uint64) bool {
	// Get bit positions
	fromPos := findSetBit(move[0])
	toPos := findSetBit(move[1])

	// Convert to coordinates
	fromRank, fromFile := fromPos/8, fromPos%8
	toRank, toFile := toPos/8, toPos%8

	// Calculate the distance moved
	rankDiff := abs(toRank - fromRank)
	fileDiff := abs(toFile - fromFile)

	// Player color
	kingPiece, kingColor := getPieceAtPosition(arbiter, fromPos)

	// Regular king move: one square in any direction
	if rankDiff <= 1 && fileDiff <= 1 {
		// Create a buffer arbiter to check if the destination square is under attack
		bufferArbiter := *arbiter

		// Temporarily remove the king from its current position
		bufferArbiter.BoardwithParameters.Board[kingPiece] &= ^move[0]

		// Temporarily place the king at the destination
		bufferArbiter.BoardwithParameters.Board[kingPiece] |= move[1]

		// Switch turn to see if opponent can attack the king at this position
		bufferArbiter.BoardwithParameters.TurnOfPlayer = 1 - kingColor

		// Check if the king would be in check at the destination
		if IsCheck(&bufferArbiter) {
			return false // Cannot move into check
		}

		return true
	}

	// If we reach here, it's not a regular king move
	// Check if it might be castling (always on the king's starting rank)
	turnOfPlayer := arbiter.BoardwithParameters.TurnOfPlayer

	// Castling conditions: king moves 2 squares horizontally on its home rank
	if rankDiff == 0 && fileDiff == 2 {
		// White king
		if turnOfPlayer == 0 && fromRank == 0 && fromFile == 4 {
			// First check if the king is currently in check
			if IsCheck(arbiter) {
				return false // Cannot castle out of check
			}

			// Check if castling is allowed according to flags
			if toFile == 6 { // Kingside castling
				// Check if kingside castling is allowed
				if arbiter.BoardwithParameters.WhiteCastle&1 == 0 {
					return false // Kingside castling not allowed for white
				}

				// Check if squares between king and rook are empty
				squareF1 := 5 // f1 square
				pieceF1, _ := getPieceAtPosition(arbiter, squareF1)
				if pieceF1 != -1 {
					return false // Path is not clear
				}

				squareG1 := 6 // g1 square
				pieceG1, _ := getPieceAtPosition(arbiter, squareG1)
				if pieceG1 != -1 {
					return false // Path is not clear
				}

				// Check if king passes through check during castling (F1 square)
				bufferArbiter := *arbiter
				// Move king to f1 temporarily
				f1Bitboard := uint64(1) << squareF1
				bufferArbiter.BoardwithParameters.Board[WhiteKing] &= ^move[0]   // Remove from e1
				bufferArbiter.BoardwithParameters.Board[WhiteKing] |= f1Bitboard // Place on f1

				bufferArbiter.BoardwithParameters.TurnOfPlayer = 1 // Black's turn to check if king would be in check
				if IsCheck(&bufferArbiter) {
					return false // Cannot castle through check
				}

				// Check if king would end up in check after castling (G1 square)
				bufferArbiter = *arbiter // Reset buffer
				g1Bitboard := uint64(1) << squareG1
				bufferArbiter.BoardwithParameters.Board[WhiteKing] &= ^move[0]   // Remove from e1
				bufferArbiter.BoardwithParameters.Board[WhiteKing] |= g1Bitboard // Place on g1

				bufferArbiter.BoardwithParameters.TurnOfPlayer = 1 // Black's turn to check if king would be in check
				if IsCheck(&bufferArbiter) {
					return false // Cannot castle into check
				}

				// Check if rook is actually there
				rookPos := 7 // h1 square
				rookPiece, rookColor := getPieceAtPosition(arbiter, rookPos)
				if rookPiece != WhiteRook || rookColor != 0 {
					return false // Rook not in correct position
				}

				return true
			}

			if toFile == 2 { // Queenside castling
				// Check if queenside castling is allowed
				if arbiter.BoardwithParameters.WhiteCastle&2 == 0 {
					return false // Queenside castling not allowed for white
				}

				// Check if squares between king and rook are empty
				squareB1 := 1 // b1 square
				pieceB1, _ := getPieceAtPosition(arbiter, squareB1)
				if pieceB1 != -1 {
					return false // Path is not clear
				}

				squareC1 := 2 // c1 square
				pieceC1, _ := getPieceAtPosition(arbiter, squareC1)
				if pieceC1 != -1 {
					return false // Path is not clear
				}

				squareD1 := 3 // d1 square
				pieceD1, _ := getPieceAtPosition(arbiter, squareD1)
				if pieceD1 != -1 {
					return false // Path is not clear
				}

				// Check if king passes through check during castling (D1 square)
				bufferArbiter := *arbiter
				// Move king to d1 temporarily
				d1Bitboard := uint64(1) << squareD1
				bufferArbiter.BoardwithParameters.Board[WhiteKing] &= ^move[0]   // Remove from e1
				bufferArbiter.BoardwithParameters.Board[WhiteKing] |= d1Bitboard // Place on d1

				bufferArbiter.BoardwithParameters.TurnOfPlayer = 1 // Black's turn to check if king would be in check
				if IsCheck(&bufferArbiter) {
					return false // Cannot castle through check
				}

				// Check if king would end up in check after castling (C1 square)
				bufferArbiter = *arbiter // Reset buffer
				c1Bitboard := uint64(1) << squareC1
				bufferArbiter.BoardwithParameters.Board[WhiteKing] &= ^move[0]   // Remove from e1
				bufferArbiter.BoardwithParameters.Board[WhiteKing] |= c1Bitboard // Place on c1

				bufferArbiter.BoardwithParameters.TurnOfPlayer = 1 // Black's turn to check if king would be in check
				if IsCheck(&bufferArbiter) {
					return false // Cannot castle into check
				}

				// Check if rook is actually there
				rookPos := 0 // a1 square
				rookPiece, rookColor := getPieceAtPosition(arbiter, rookPos)
				if rookPiece != WhiteRook || rookColor != 0 {
					return false // Rook not in correct position
				}

				return true
			}
		}

		// Black king
		if turnOfPlayer == 1 && fromRank == 7 && fromFile == 4 {
			// First check if the king is currently in check
			if IsCheck(arbiter) {
				return false // Cannot castle out of check
			}

			// Check if castling is allowed according to flags
			if toFile == 6 { // Kingside castling
				// Check if kingside castling is allowed
				if arbiter.BoardwithParameters.BlackCastle&1 == 0 {
					return false // Kingside castling not allowed for black
				}

				// Check if squares between king and rook are empty
				squareF8 := 61 // f8 square
				pieceF8, _ := getPieceAtPosition(arbiter, squareF8)
				if pieceF8 != -1 {
					return false // Path is not clear
				}

				squareG8 := 62 // g8 square
				pieceG8, _ := getPieceAtPosition(arbiter, squareG8)
				if pieceG8 != -1 {
					return false // Path is not clear
				}

				// Check if king passes through check during castling (F8 square)
				bufferArbiter := *arbiter
				// Move king to f8 temporarily
				f8Bitboard := uint64(1) << squareF8
				bufferArbiter.BoardwithParameters.Board[BlackKing] &= ^move[0]   // Remove from e8
				bufferArbiter.BoardwithParameters.Board[BlackKing] |= f8Bitboard // Place on f8

				bufferArbiter.BoardwithParameters.TurnOfPlayer = 0 // White's turn to check if king would be in check
				if IsCheck(&bufferArbiter) {
					return false // Cannot castle through check
				}

				// Check if king would end up in check after castling (G8 square)
				bufferArbiter = *arbiter // Reset buffer
				g8Bitboard := uint64(1) << squareG8
				bufferArbiter.BoardwithParameters.Board[BlackKing] &= ^move[0]   // Remove from e8
				bufferArbiter.BoardwithParameters.Board[BlackKing] |= g8Bitboard // Place on g8

				bufferArbiter.BoardwithParameters.TurnOfPlayer = 0 // White's turn to check if king would be in check
				if IsCheck(&bufferArbiter) {
					return false // Cannot castle into check
				}

				// Check if rook is actually there
				rookPos := 63 // h8 square
				rookPiece, rookColor := getPieceAtPosition(arbiter, rookPos)
				if rookPiece != BlackRook || rookColor != 1 {
					return false // Rook not in correct position
				}

				return true
			}

			if toFile == 2 { // Queenside castling
				// Check if queenside castling is allowed
				if arbiter.BoardwithParameters.BlackCastle&2 == 0 {
					return false // Queenside castling not allowed for black
				}

				// Check if squares between king and rook are empty
				squareB8 := 57 // b8 square
				pieceB8, _ := getPieceAtPosition(arbiter, squareB8)
				if pieceB8 != -1 {
					return false // Path is not clear
				}

				squareC8 := 58 // c8 square
				pieceC8, _ := getPieceAtPosition(arbiter, squareC8)
				if pieceC8 != -1 {
					return false // Path is not clear
				}

				squareD8 := 59 // d8 square
				pieceD8, _ := getPieceAtPosition(arbiter, squareD8)
				if pieceD8 != -1 {
					return false // Path is not clear
				}

				// Check if king passes through check during castling (D8 square)
				bufferArbiter := *arbiter
				// Move king to d8 temporarily
				d8Bitboard := uint64(1) << squareD8
				bufferArbiter.BoardwithParameters.Board[BlackKing] &= ^move[0]   // Remove from e8
				bufferArbiter.BoardwithParameters.Board[BlackKing] |= d8Bitboard // Place on d8

				bufferArbiter.BoardwithParameters.TurnOfPlayer = 0 // White's turn to check if king would be in check
				if IsCheck(&bufferArbiter) {
					return false // Cannot castle through check
				}

				// Check if king would end up in check after castling (C8 square)
				bufferArbiter = *arbiter // Reset buffer
				c8Bitboard := uint64(1) << squareC8
				bufferArbiter.BoardwithParameters.Board[BlackKing] &= ^move[0]   // Remove from e8
				bufferArbiter.BoardwithParameters.Board[BlackKing] |= c8Bitboard // Place on c8

				bufferArbiter.BoardwithParameters.TurnOfPlayer = 0 // White's turn to check if king would be in check
				if IsCheck(&bufferArbiter) {
					return false // Cannot castle into check
				}

				// Check if rook is actually there
				rookPos := 56 // a8 square
				rookPiece, rookColor := getPieceAtPosition(arbiter, rookPos)
				if rookPiece != BlackRook || rookColor != 1 {
					return false // Rook not in correct position
				}

				return true
			}
		}
	}

	// If we've reached here, the move is not valid
	return false
}

// Bishop movement validation
func isValidBishopMove(arbiter *ChessArbiter, move [3]uint64) bool {
	// Get bit positions
	fromPos := findSetBit(move[0])
	toPos := findSetBit(move[1])

	// Convert to coordinates
	fromRank, fromFile := fromPos/8, fromPos%8
	toRank, toFile := toPos/8, toPos%8

	// Bishop moves diagonally, so the absolute difference in rank and file should be equal
	rankDiff := abs(toRank - fromRank)
	fileDiff := abs(toFile - fromFile)

	if rankDiff != fileDiff {
		return false // Not a diagonal move
	}

	// Check if the path is clear
	rankDir := sign(toRank - fromRank)
	fileDir := sign(toFile - fromFile)

	// Check each square along the diagonal path
	for i := 1; i < rankDiff; i++ {
		checkRank := fromRank + i*rankDir
		checkFile := fromFile + i*fileDir
		checkPos := checkRank*8 + checkFile

		// If there's a piece in the way, the move is invalid
		piece, _ := getPieceAtPosition(arbiter, checkPos)
		if piece != -1 {
			return false
		}
	}

	return true
}

// Rook movement validation
func isValidRookMove(arbiter *ChessArbiter, move [3]uint64) bool {
	// Get bit positions
	fromPos := findSetBit(move[0])
	toPos := findSetBit(move[1])

	// Convert to coordinates
	fromRank, fromFile := fromPos/8, fromPos%8
	toRank, toFile := toPos/8, toPos%8

	// Rook moves horizontally or vertically, so either the rank or file must remain the same
	if fromRank != toRank && fromFile != toFile {
		return false // Neither a horizontal nor a vertical move
	}

	// Check if the path is clear
	if fromRank == toRank {
		// Horizontal move
		start, end := min(fromFile, toFile), max(fromFile, toFile)

		// Check each square along the horizontal path
		for file := start + 1; file < end; file++ {
			checkPos := fromRank*8 + file
			piece, _ := getPieceAtPosition(arbiter, checkPos)
			if piece != -1 {
				return false // Piece in the way
			}
		}
	} else {
		// Vertical move
		start, end := min(fromRank, toRank), max(fromRank, toRank)

		// Check each square along the vertical path
		for rank := start + 1; rank < end; rank++ {
			checkPos := rank*8 + fromFile
			piece, _ := getPieceAtPosition(arbiter, checkPos)
			if piece != -1 {
				return false // Piece in the way
			}
		}
	}

	return true
}

// Knight movement validation
func isValidKnightMove(arbiter *ChessArbiter, move [3]uint64) bool {
	// Get bit positions
	fromPos := findSetBit(move[0])
	toPos := findSetBit(move[1])

	// Convert to coordinates
	fromRank, fromFile := fromPos/8, fromPos%8
	toRank, toFile := toPos/8, toPos%8

	// Knights move in an L-shape: 2 squares in one direction and 1 square perpendicular
	rankDiff := abs(toRank - fromRank)
	fileDiff := abs(toFile - fromFile)

	// A valid knight move is either (2,1) or (1,2)
	return (rankDiff == 2 && fileDiff == 1) || (rankDiff == 1 && fileDiff == 2)
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GenerateValidMoves generates all valid moves for the current player
func GenerateValidMoves(arbiter *ChessArbiter) [][3]uint64 {
	var allMoves [][3]uint64

	// Generate moves for each piece type based on whose turn it is
	playerColor := arbiter.BoardwithParameters.TurnOfPlayer

	// Generate king moves
	kingMoves := generateValidKingMoves(arbiter, playerColor)
	allMoves = append(allMoves, kingMoves...)

	// Generate queen moves
	queenMoves := generateValidQueenMoves(arbiter, playerColor)
	allMoves = append(allMoves, queenMoves...)

	// Generate rook moves
	rookMoves := generateValidRookMoves(arbiter, playerColor)
	allMoves = append(allMoves, rookMoves...)

	// Generate bishop moves
	bishopMoves := generateValidBishopMoves(arbiter, playerColor)
	allMoves = append(allMoves, bishopMoves...)

	// Generate knight moves
	knightMoves := generateValidKnightMoves(arbiter, playerColor)
	allMoves = append(allMoves, knightMoves...)

	// Generate pawn moves
	pawnMoves := generateValidPawnMoves(arbiter, playerColor)
	allMoves = append(allMoves, pawnMoves...)

	return allMoves
}

// generateValidKingMoves generates all valid moves for the king of the specified color
func generateValidKingMoves(arbiter *ChessArbiter, playerColor int) [][3]uint64 {
	var kingMoves [][3]uint64

	// Determine which king we're generating moves for
	kingPiece := WhiteKing
	if playerColor == 1 {
		kingPiece = BlackKing
	}

	// Get the king's position
	kingBitboard := arbiter.BoardwithParameters.Board[kingPiece]
	if kingBitboard == 0 {
		return kingMoves // No king found (shouldn't happen in a valid game)
	}

	kingPos := findSetBit(kingBitboard)
	kingBit := uint64(1) << kingPos

	// King can move one square in any direction
	// Convert to rank and file to calculate adjacent squares
	rank, file := kingPos/8, kingPos%8

	// Define the 8 possible king move directions
	directions := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1}, // Top-left, top, top-right
		{0, -1}, {0, 1}, // Left, right
		{1, -1}, {1, 0}, {1, 1}, // Bottom-left, bottom, bottom-right
	}

	// Check each direction
	for _, dir := range directions {
		newRank, newFile := rank+dir[0], file+dir[1]

		// Check if the new position is on the board
		if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
			newPos := newRank*8 + newFile
			targetBit := uint64(1) << newPos

			// Create a move to check
			move := [3]uint64{kingBit, targetBit, 0}

			// Use existing validation function to check if it's valid
			if IsValidMove(arbiter, move) {
				kingMoves = append(kingMoves, move)
			}
		}
	}

	// Check for castling moves
	// For white king
	if playerColor == 0 && kingPos == 4 { // e1 - starting position
		// Check kingside castling
		if arbiter.BoardwithParameters.WhiteCastle&1 != 0 { // Kingside castling available
			kingsideCastle := [3]uint64{kingBit, uint64(1) << 6, 0} // e1 to g1
			if IsValidMove(arbiter, kingsideCastle) {
				kingMoves = append(kingMoves, kingsideCastle)
			}
		}

		// Check queenside castling
		if arbiter.BoardwithParameters.WhiteCastle&2 != 0 { // Queenside castling available
			queensideCastle := [3]uint64{kingBit, uint64(1) << 2, 0} // e1 to c1
			if IsValidMove(arbiter, queensideCastle) {
				kingMoves = append(kingMoves, queensideCastle)
			}
		}
	}

	// For black king
	if playerColor == 1 && kingPos == 60 { // e8 - starting position
		// Check kingside castling
		if arbiter.BoardwithParameters.BlackCastle&1 != 0 { // Kingside castling available
			kingsideCastle := [3]uint64{kingBit, uint64(1) << 62, 0} // e8 to g8
			if IsValidMove(arbiter, kingsideCastle) {
				kingMoves = append(kingMoves, kingsideCastle)
			}
		}

		// Check queenside castling
		if arbiter.BoardwithParameters.BlackCastle&2 != 0 { // Queenside castling available
			queensideCastle := [3]uint64{kingBit, uint64(1) << 58, 0} // e8 to c8
			if IsValidMove(arbiter, queensideCastle) {
				kingMoves = append(kingMoves, queensideCastle)
			}
		}
	}

	return kingMoves
}

// generateValidQueenMoves generates all valid moves for the queens of the specified color
func generateValidQueenMoves(arbiter *ChessArbiter, playerColor int) [][3]uint64 {
	var queenMoves [][3]uint64

	// Determine which queen we're generating moves for
	queenPiece := WhiteQueen
	if playerColor == 1 {
		queenPiece = BlackQueen
	}

	// Get the queen's positions
	queenBitboard := arbiter.BoardwithParameters.Board[queenPiece]

	// For each queen on the board
	for queenBitboard != 0 {
		// Find the position of the least significant bit (a queen)
		queenPos := findSetBit(queenBitboard)
		queenBit := uint64(1) << queenPos

		// Clear this bit so we can find the next queen (if any)
		queenBitboard &= ^queenBit

		// Queen moves like a rook and a bishop combined
		// Generate rook-like moves (horizontal and vertical)
		rank, file := queenPos/8, queenPos%8

		// Check each of the four directions (up, right, down, left)
		// Horizontal moves (left and right)
		for newFile := 0; newFile < 8; newFile++ {
			if newFile == file {
				continue // Skip the queen's current file
			}

			newPos := rank*8 + newFile
			targetBit := uint64(1) << newPos

			// Create a move to check
			move := [3]uint64{queenBit, targetBit, 0}

			// Use existing validation function to check if it's valid
			if IsValidMove(arbiter, move) {
				queenMoves = append(queenMoves, move)
			}
		}

		// Vertical moves (up and down)
		for newRank := 0; newRank < 8; newRank++ {
			if newRank == rank {
				continue // Skip the queen's current rank
			}

			newPos := newRank*8 + file
			targetBit := uint64(1) << newPos

			// Create a move to check
			move := [3]uint64{queenBit, targetBit, 0}

			// Use existing validation function to check if it's valid
			if IsValidMove(arbiter, move) {
				queenMoves = append(queenMoves, move)
			}
		}

		// Generate bishop-like moves (diagonals)
		// Check diagonals in all four directions
		// Direction: top-left to bottom-right
		for offset := -7; offset <= 7; offset++ {
			if offset == 0 {
				continue // Skip the queen's current position
			}

			newRank, newFile := rank+offset, file+offset

			// Check if the new position is on the board
			if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
				newPos := newRank*8 + newFile
				targetBit := uint64(1) << newPos

				// Create a move to check
				move := [3]uint64{queenBit, targetBit, 0}

				// Use existing validation function to check if it's valid
				if IsValidMove(arbiter, move) {
					queenMoves = append(queenMoves, move)
				}
			}
		}

		// Direction: top-right to bottom-left
		for offset := -7; offset <= 7; offset++ {
			if offset == 0 {
				continue // Skip the queen's current position
			}

			newRank, newFile := rank+offset, file-offset

			// Check if the new position is on the board
			if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
				newPos := newRank*8 + newFile
				targetBit := uint64(1) << newPos

				// Create a move to check
				move := [3]uint64{queenBit, targetBit, 0}

				// Use existing validation function to check if it's valid
				if IsValidMove(arbiter, move) {
					queenMoves = append(queenMoves, move)
				}
			}
		}
	}

	return queenMoves
}

// generateValidRookMoves generates all valid moves for the rooks of the specified color
func generateValidRookMoves(arbiter *ChessArbiter, playerColor int) [][3]uint64 {
	var rookMoves [][3]uint64

	// Determine which rook we're generating moves for
	rookPiece := WhiteRook
	if playerColor == 1 {
		rookPiece = BlackRook
	}

	// Get the rook's positions
	rookBitboard := arbiter.BoardwithParameters.Board[rookPiece]

	// For each rook on the board
	for rookBitboard != 0 {
		// Find the position of the least significant bit (a rook)
		rookPos := findSetBit(rookBitboard)
		rookBit := uint64(1) << rookPos

		// Clear this bit so we can find the next rook (if any)
		rookBitboard &= ^rookBit

		// Rook moves horizontally and vertically
		rank, file := rookPos/8, rookPos%8

		// Horizontal moves (left and right)
		for newFile := 0; newFile < 8; newFile++ {
			if newFile == file {
				continue // Skip the rook's current file
			}

			newPos := rank*8 + newFile
			targetBit := uint64(1) << newPos

			// Create a move to check
			move := [3]uint64{rookBit, targetBit, 0}

			// Use existing validation function to check if it's valid
			if IsValidMove(arbiter, move) {
				rookMoves = append(rookMoves, move)
			}
		}

		// Vertical moves (up and down)
		for newRank := 0; newRank < 8; newRank++ {
			if newRank == rank {
				continue // Skip the rook's current rank
			}

			newPos := newRank*8 + file
			targetBit := uint64(1) << newPos

			// Create a move to check
			move := [3]uint64{rookBit, targetBit, 0}

			// Use existing validation function to check if it's valid
			if IsValidMove(arbiter, move) {
				rookMoves = append(rookMoves, move)
			}
		}
	}

	return rookMoves
}

// generateValidBishopMoves generates all valid moves for the bishops of the specified color
func generateValidBishopMoves(arbiter *ChessArbiter, playerColor int) [][3]uint64 {
	var bishopMoves [][3]uint64

	// Determine which bishop we're generating moves for
	bishopPiece := WhiteBishop
	if playerColor == 1 {
		bishopPiece = BlackBishop
	}

	// Get the bishop's positions
	bishopBitboard := arbiter.BoardwithParameters.Board[bishopPiece]

	// For each bishop on the board
	for bishopBitboard != 0 {
		// Find the position of the least significant bit (a bishop)
		bishopPos := findSetBit(bishopBitboard)
		bishopBit := uint64(1) << bishopPos

		// Clear this bit so we can find the next bishop (if any)
		bishopBitboard &= ^bishopBit

		// Bishop moves diagonally
		rank, file := bishopPos/8, bishopPos%8

		// Check diagonals in all four directions
		// Direction: top-left to bottom-right
		for offset := -7; offset <= 7; offset++ {
			if offset == 0 {
				continue // Skip the bishop's current position
			}

			newRank, newFile := rank+offset, file+offset

			// Check if the new position is on the board
			if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
				newPos := newRank*8 + newFile
				targetBit := uint64(1) << newPos

				// Create a move to check
				move := [3]uint64{bishopBit, targetBit, 0}

				// Use existing validation function to check if it's valid
				if IsValidMove(arbiter, move) {
					bishopMoves = append(bishopMoves, move)
				}
			}
		}

		// Direction: top-right to bottom-left
		for offset := -7; offset <= 7; offset++ {
			if offset == 0 {
				continue // Skip the bishop's current position
			}

			newRank, newFile := rank+offset, file-offset

			// Check if the new position is on the board
			if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
				newPos := newRank*8 + newFile
				targetBit := uint64(1) << newPos

				// Create a move to check
				move := [3]uint64{bishopBit, targetBit, 0}

				// Use existing validation function to check if it's valid
				if IsValidMove(arbiter, move) {
					bishopMoves = append(bishopMoves, move)
				}
			}
		}
	}

	return bishopMoves
}

// generateValidKnightMoves generates all valid moves for the knights of the specified color
func generateValidKnightMoves(arbiter *ChessArbiter, playerColor int) [][3]uint64 {
	var knightMoves [][3]uint64

	// Determine which knight we're generating moves for
	knightPiece := WhiteKnight
	if playerColor == 1 {
		knightPiece = BlackKnight
	}

	// Get the knight's positions
	knightBitboard := arbiter.BoardwithParameters.Board[knightPiece]

	// For each knight on the board
	for knightBitboard != 0 {
		// Find the position of the least significant bit (a knight)
		knightPos := findSetBit(knightBitboard)
		knightBit := uint64(1) << knightPos

		// Clear this bit so we can find the next knight (if any)
		knightBitboard &= ^knightBit

		// Knight moves in an L-shape (2 squares in one direction, then 1 square perpendicular)
		rank, file := knightPos/8, knightPos%8

		// Define the 8 possible knight move offsets
		knightOffsets := [][2]int{
			{-2, -1}, {-2, 1}, // Up 2, left/right 1
			{-1, -2}, {-1, 2}, // Up 1, left/right 2
			{1, -2}, {1, 2}, // Down 1, left/right 2
			{2, -1}, {2, 1}, // Down 2, left/right 1
		}

		// Check each possible knight move
		for _, offset := range knightOffsets {
			newRank, newFile := rank+offset[0], file+offset[1]

			// Check if the new position is on the board
			if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
				newPos := newRank*8 + newFile
				targetBit := uint64(1) << newPos

				// Create a move to check
				move := [3]uint64{knightBit, targetBit, 0}

				// Use existing validation function to check if it's valid
				if IsValidMove(arbiter, move) {
					knightMoves = append(knightMoves, move)
				}
			}
		}
	}

	return knightMoves
}

// generateValidPawnMoves generates all valid moves for the pawns of the specified color
func generateValidPawnMoves(arbiter *ChessArbiter, playerColor int) [][3]uint64 {
	var pawnMoves [][3]uint64

	// Determine which pawn we're generating moves for
	pawnPiece := WhitePawn
	if playerColor == 1 {
		pawnPiece = BlackPawn
	}

	// Get the pawn's positions
	pawnBitboard := arbiter.BoardwithParameters.Board[pawnPiece]

	// For each pawn on the board
	for pawnBitboard != 0 {
		// Find the position of the least significant bit (a pawn)
		pawnPos := findSetBit(pawnBitboard)
		pawnBit := uint64(1) << pawnPos

		// Clear this bit so we can find the next pawn (if any)
		pawnBitboard &= ^pawnBit

		rank, file := pawnPos/8, pawnPos%8

		// Movement direction depends on pawn color
		rankDir := 1       // White pawns move up the board (increasing rank)
		promotionRank := 7 // White pawns promote on the 8th rank
		startingRank := 1  // White pawns start on the 2nd rank

		if playerColor == 1 {
			rankDir = -1      // Black pawns move down the board (decreasing rank)
			promotionRank = 0 // Black pawns promote on the 1st rank
			startingRank = 6  // Black pawns start on the 7th rank
		}

		// Forward move (1 square)
		newRank := rank + rankDir

		// Check if the new position is on the board
		if newRank >= 0 && newRank < 8 {
			newPos := newRank*8 + file
			targetBit := uint64(1) << newPos

			// Promotions happen when a pawn reaches the end rank
			if newRank == promotionRank {
				// Create moves for each possible promotion piece
				promotionPieces := []int{}
				if playerColor == 0 {
					promotionPieces = []int{WhiteQueen, WhiteRook, WhiteBishop, WhiteKnight}
				} else {
					promotionPieces = []int{BlackQueen, BlackRook, BlackBishop, BlackKnight}
				}

				for _, promotionPiece := range promotionPieces {
					move := [3]uint64{pawnBit, targetBit, uint64(promotionPiece)}

					// Use existing validation function to check if it's valid
					if IsValidMove(arbiter, move) {
						pawnMoves = append(pawnMoves, move)
					}
				}
			} else {
				// Regular non-promotion move
				move := [3]uint64{pawnBit, targetBit, 0}

				// Use existing validation function to check if it's valid
				if IsValidMove(arbiter, move) {
					pawnMoves = append(pawnMoves, move)
				}
			}
		}

		// Forward move (2 squares, only from starting position)
		if rank == startingRank {
			newRank := rank + (2 * rankDir)

			// Check if the new position is on the board
			if newRank >= 0 && newRank < 8 {
				newPos := newRank*8 + file
				targetBit := uint64(1) << newPos

				move := [3]uint64{pawnBit, targetBit, 0}

				// Use existing validation function to check if it's valid
				if IsValidMove(arbiter, move) {
					pawnMoves = append(pawnMoves, move)
				}
			}
		}

		// Capture moves (diagonal)
		for fileDiff := -1; fileDiff <= 1; fileDiff += 2 { // Check both left and right diagonals
			newFile := file + fileDiff
			newRank := rank + rankDir

			// Check if the new position is on the board
			if newRank >= 0 && newRank < 8 && newFile >= 0 && newFile < 8 {
				newPos := newRank*8 + newFile
				targetBit := uint64(1) << newPos

				// Promotions happen when a pawn reaches the end rank
				if newRank == promotionRank {
					// Create moves for each possible promotion piece
					promotionPieces := []int{}
					if playerColor == 0 {
						promotionPieces = []int{WhiteQueen, WhiteRook, WhiteBishop, WhiteKnight}
					} else {
						promotionPieces = []int{BlackQueen, BlackRook, BlackBishop, BlackKnight}
					}

					for _, promotionPiece := range promotionPieces {
						move := [3]uint64{pawnBit, targetBit, uint64(promotionPiece)}

						// Use existing validation function to check if it's valid
						if IsValidMove(arbiter, move) {
							pawnMoves = append(pawnMoves, move)
						}
					}
				} else {
					// Regular non-promotion capture
					move := [3]uint64{pawnBit, targetBit, 0}

					// Use existing validation function to check if it's valid
					if IsValidMove(arbiter, move) {
						pawnMoves = append(pawnMoves, move)
					}
				}
			}
		}

		// En passant captures
		if playerColor == 0 { // White pawns can only capture black pawns' en passant
			if arbiter.BoardwithParameters.EnPassantBlack != 0 && rank == 4 { // White pawns can en passant from the 5th rank
				// Find the en passant target square
				epSquare := findSetBit(arbiter.BoardwithParameters.EnPassantBlack)
				epFile := epSquare % 8

				// Check if the pawn is adjacent to the en passant square
				if abs(file-epFile) == 1 {
					// Verify there's actually a black pawn to capture
					capturedPawnPos := epSquare - 8 // One rank below the en passant square
					capturedPawnBit := uint64(1) << capturedPawnPos

					if arbiter.BoardwithParameters.Board[BlackPawn]&capturedPawnBit != 0 {
						move := [3]uint64{pawnBit, arbiter.BoardwithParameters.EnPassantBlack, 0}

						// Use existing validation function to check if it's valid
						if IsValidMove(arbiter, move) {
							pawnMoves = append(pawnMoves, move)
						}
					}
				}
			}
		} else if playerColor == 1 { // Black pawns can only capture white pawns' en passant
			if arbiter.BoardwithParameters.EnPassantWhite != 0 && rank == 3 { // Black pawns can en passant from the 4th rank
				// Find the en passant target square
				epSquare := findSetBit(arbiter.BoardwithParameters.EnPassantWhite)
				epFile := epSquare % 8

				// Check if the pawn is adjacent to the en passant square
				if abs(file-epFile) == 1 {
					// Verify there's actually a white pawn to capture
					capturedPawnPos := epSquare + 8 // One rank above the en passant square
					capturedPawnBit := uint64(1) << capturedPawnPos

					if arbiter.BoardwithParameters.Board[WhitePawn]&capturedPawnBit != 0 {
						move := [3]uint64{pawnBit, arbiter.BoardwithParameters.EnPassantWhite, 0}

						// Use existing validation function to check if it's valid
						if IsValidMove(arbiter, move) {
							pawnMoves = append(pawnMoves, move)
						}
					}
				}
			}
		}
	}

	return pawnMoves
}

// BoardParamsToFEN converts a BoardwithParameters to a FEN string
func BoardParamsToFEN(boardParams chess.BoardwithParameters) string {
	var fen strings.Builder

	// 1. Piece placement
	for rank := 7; rank >= 0; rank-- {
		emptySquares := 0

		for file := 0; file < 8; file++ {
			squareIndex := rank*8 + file
			bitMask := uint64(1) << squareIndex
			piece := ' '

			// Check each bitboard to see if a piece is on this square
			if boardParams.Board[WhiteKing]&bitMask != 0 {
				piece = 'K'
			} else if boardParams.Board[WhiteQueen]&bitMask != 0 {
				piece = 'Q'
			} else if boardParams.Board[WhiteRook]&bitMask != 0 {
				piece = 'R'
			} else if boardParams.Board[WhiteBishop]&bitMask != 0 {
				piece = 'B'
			} else if boardParams.Board[WhiteKnight]&bitMask != 0 {
				piece = 'N'
			} else if boardParams.Board[WhitePawn]&bitMask != 0 {
				piece = 'P'
			} else if boardParams.Board[BlackKing]&bitMask != 0 {
				piece = 'k'
			} else if boardParams.Board[BlackQueen]&bitMask != 0 {
				piece = 'q'
			} else if boardParams.Board[BlackRook]&bitMask != 0 {
				piece = 'r'
			} else if boardParams.Board[BlackBishop]&bitMask != 0 {
				piece = 'b'
			} else if boardParams.Board[BlackKnight]&bitMask != 0 {
				piece = 'n'
			} else if boardParams.Board[BlackPawn]&bitMask != 0 {
				piece = 'p'
			}

			if piece == ' ' {
				emptySquares++
			} else {
				if emptySquares > 0 {
					fen.WriteString(strconv.Itoa(emptySquares))
					emptySquares = 0
				}
				fen.WriteRune(piece)
			}
		}

		if emptySquares > 0 {
			fen.WriteString(strconv.Itoa(emptySquares))
		}

		if rank > 0 {
			fen.WriteRune('/')
		}
	}

	// 2. Active color
	if boardParams.TurnOfPlayer == 0 {
		fen.WriteString(" w ")
	} else {
		fen.WriteString(" b ")
	}

	// 3. Castling availability
	castlingRights := ""
	if boardParams.WhiteCastle&1 != 0 { // Kingside
		castlingRights += "K"
	}
	if boardParams.WhiteCastle&2 != 0 { // Queenside
		castlingRights += "Q"
	}
	if boardParams.BlackCastle&1 != 0 { // Kingside
		castlingRights += "k"
	}
	if boardParams.BlackCastle&2 != 0 { // Queenside
		castlingRights += "q"
	}

	if castlingRights == "" {
		fen.WriteString("-")
	} else {
		fen.WriteString(castlingRights)
	}

	// 4. En passant target square
	fen.WriteString(" ")
	enPassantBitboard := uint64(0)
	if boardParams.TurnOfPlayer == 0 && boardParams.EnPassantBlack != 0 {
		enPassantBitboard = boardParams.EnPassantBlack
	} else if boardParams.TurnOfPlayer == 1 && boardParams.EnPassantWhite != 0 {
		enPassantBitboard = boardParams.EnPassantWhite
	}

	if enPassantBitboard != 0 {
		// Find the position of the set bit
		enPassantSquare := findSetBit(enPassantBitboard)
		file := enPassantSquare % 8
		rank := enPassantSquare / 8
		fen.WriteRune(rune('a' + file))
		fen.WriteRune(rune('1' + rank))
	} else {
		fen.WriteString("-")
	}

	// 5. Halfmove clock (not tracked in our implementation)
	fen.WriteString(" 0")

	// 6. Fullmove number (not tracked in our implementation)
	fen.WriteString(" 1")

	return fen.String()
}

func CreateGameArbiter(fen string) (*ChessArbiter, error) {
	if fen == "" {
		fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	}

	arbiter := &ChessArbiter{
		BoardwithParameters: BoardwithParameters{},
	}

	// Initialize all bitboards to 0
	for i := range arbiter.BoardwithParameters.Board {
		arbiter.BoardwithParameters.Board[i] = 0
	}

	// Initialize en passant bitboards to 0
	arbiter.BoardwithParameters.EnPassantWhite = 0
	arbiter.BoardwithParameters.EnPassantBlack = 0

	// Split FEN into its components
	parts := strings.Split(fen, " ")
	if len(parts) < 4 {
		return nil, errors.New("invalid FEN: not enough components")
	}

	// Parse board position
	board := parts[0]
	rank := 7 // Start at the 8th rank (0-indexed)
	file := 0

	for _, char := range board {
		switch char {
		case '/':
			rank--
			file = 0
		case '1', '2', '3', '4', '5', '6', '7', '8':
			file += int(char - '0')
		default:
			// Calculate square index (0-63)
			squareIndex := rank*8 + file

			// Set the appropriate bit in the correct bitboard
			bitMask := uint64(1) << squareIndex

			switch char {
			case 'K':
				arbiter.BoardwithParameters.Board[WhiteKing] |= bitMask
			case 'Q':
				arbiter.BoardwithParameters.Board[WhiteQueen] |= bitMask
			case 'R':
				arbiter.BoardwithParameters.Board[WhiteRook] |= bitMask
			case 'B':
				arbiter.BoardwithParameters.Board[WhiteBishop] |= bitMask
			case 'N':
				arbiter.BoardwithParameters.Board[WhiteKnight] |= bitMask
			case 'P':
				arbiter.BoardwithParameters.Board[WhitePawn] |= bitMask
			case 'k':
				arbiter.BoardwithParameters.Board[BlackKing] |= bitMask
			case 'q':
				arbiter.BoardwithParameters.Board[BlackQueen] |= bitMask
			case 'r':
				arbiter.BoardwithParameters.Board[BlackRook] |= bitMask
			case 'b':
				arbiter.BoardwithParameters.Board[BlackBishop] |= bitMask
			case 'n':
				arbiter.BoardwithParameters.Board[BlackKnight] |= bitMask
			case 'p':
				arbiter.BoardwithParameters.Board[BlackPawn] |= bitMask
			}
			file++
		}
	}

	// Parse active color
	if parts[1] == "w" {
		arbiter.BoardwithParameters.TurnOfPlayer = 0
	} else {
		arbiter.BoardwithParameters.TurnOfPlayer = 1
	}

	// Parse castling availability
	arbiter.BoardwithParameters.WhiteCastle = 0
	arbiter.BoardwithParameters.BlackCastle = 0

	if strings.Contains(parts[2], "K") {
		arbiter.BoardwithParameters.WhiteCastle |= 1 // Kingside (right) castling
	}
	if strings.Contains(parts[2], "Q") {
		arbiter.BoardwithParameters.WhiteCastle |= 2 // Queenside (left) castling
	}
	if strings.Contains(parts[2], "k") {
		arbiter.BoardwithParameters.BlackCastle |= 1 // Kingside (right) castling
	}
	if strings.Contains(parts[2], "q") {
		arbiter.BoardwithParameters.BlackCastle |= 2 // Queenside (left) castling
	}

	// Parse en passant target square
	if parts[3] != "-" {
		file := int(parts[3][0] - 'a')
		rank := int(parts[3][1] - '1')
		enPassantSquare := rank*8 + file
		enPassantBitboard := uint64(1) << enPassantSquare

		if arbiter.BoardwithParameters.TurnOfPlayer == 0 { // White to move, so en passant square is for black
			arbiter.BoardwithParameters.EnPassantBlack = enPassantBitboard
			arbiter.BoardwithParameters.EnPassantWhite = 0
		} else { // Black to move, so en passant square is for white
			arbiter.BoardwithParameters.EnPassantWhite = enPassantBitboard
			arbiter.BoardwithParameters.EnPassantBlack = 0
		}
	} else {
		arbiter.BoardwithParameters.EnPassantWhite = 0
		arbiter.BoardwithParameters.EnPassantBlack = 0
	}
	return arbiter, nil
}

// Modified IsCheck function that doesn't use GenerateValidMoves to avoid recursion
func IsCheck(arbiter *ChessArbiter) bool {
	// Get the current player's color
	currentPlayerColor := arbiter.BoardwithParameters.TurnOfPlayer

	// Find the current player's king position
	kingPiece := WhiteKing
	if currentPlayerColor == 1 {
		kingPiece = BlackKing
	}

	// Get the king's bitboard
	kingBitboard := arbiter.BoardwithParameters.Board[kingPiece]
	if kingBitboard == 0 {
		// This shouldn't happen in a valid game, but return false if no king found
		return false
	}

	// Find the king's position
	kingPos := findSetBit(kingBitboard)

	// Temporarily switch the turn to the opponent
	opponentColor := 1 - currentPlayerColor
	originalTurn := arbiter.BoardwithParameters.TurnOfPlayer
	arbiter.BoardwithParameters.TurnOfPlayer = opponentColor

	// Check if the opponent can attack the king directly
	isInCheck := isSquareAttacked(arbiter, kingPos, opponentColor)

	// Restore the original turn
	arbiter.BoardwithParameters.TurnOfPlayer = originalTurn

	return isInCheck
}

// isSquareAttacked checks if a square is under attack by any piece of the specified color
// This avoids using GenerateValidMoves to prevent recursion
func isSquareAttacked(arbiter *ChessArbiter, square int, attackerColor int) bool {
	// Check pawn attacks
	if attackerColor == 0 { // White attacking
		// Check if black king is attacked by white pawns
		// Pawns attack diagonally forward, so check one rank below and one file to the left/right
		if square > 7 { // Not on the first rank
			// Check if white pawn can attack from bottom-left
			if square%8 > 0 { // Not on the a-file
				pawnPos := square - 9 // One rank down, one file left
				if pawnPos >= 0 {
					pawnBit := uint64(1) << pawnPos
					if arbiter.BoardwithParameters.Board[WhitePawn]&pawnBit != 0 {
						return true
					}
				}
			}

			// Check if white pawn can attack from bottom-right
			if square%8 < 7 { // Not on the h-file
				pawnPos := square - 7 // One rank down, one file right
				if pawnPos >= 0 {
					pawnBit := uint64(1) << pawnPos
					if arbiter.BoardwithParameters.Board[WhitePawn]&pawnBit != 0 {
						return true
					}
				}
			}
		}
	} else { // Black attacking
		// Check if white king is attacked by black pawns
		// Pawns attack diagonally forward, so check one rank above and one file to the left/right
		if square < 56 { // Not on the last rank
			// Check if black pawn can attack from top-left
			if square%8 > 0 { // Not on the a-file
				pawnPos := square + 7 // One rank up, one file left
				if pawnPos < 64 {
					pawnBit := uint64(1) << pawnPos
					if arbiter.BoardwithParameters.Board[BlackPawn]&pawnBit != 0 {
						return true
					}
				}
			}

			// Check if black pawn can attack from top-right
			if square%8 < 7 { // Not on the h-file
				pawnPos := square + 9 // One rank up, one file right
				if pawnPos < 64 {
					pawnBit := uint64(1) << pawnPos
					if arbiter.BoardwithParameters.Board[BlackPawn]&pawnBit != 0 {
						return true
					}
				}
			}
		}
	}

	// Get the knight piece index for the attacker color
	knightPiece := WhiteKnight
	if attackerColor == 1 {
		knightPiece = BlackKnight
	}

	// Check knight attacks
	knightOffsets := []int{-17, -15, -10, -6, 6, 10, 15, 17}
	for _, offset := range knightOffsets {
		attackPos := square + offset

		// Make sure the position is valid and the knight's move is on the board
		// (knights can jump 2 ranks and 1 file or 1 rank and 2 files)
		if attackPos >= 0 && attackPos < 64 {
			rankDiff := abs((attackPos / 8) - (square / 8))
			fileDiff := abs((attackPos % 8) - (square % 8))

			if (rankDiff == 2 && fileDiff == 1) || (rankDiff == 1 && fileDiff == 2) {
				attackBit := uint64(1) << attackPos
				if arbiter.BoardwithParameters.Board[knightPiece]&attackBit != 0 {
					return true
				}
			}
		}
	}

	// Get the pieces indices for the attacker color
	kingPiece := WhiteKing
	queenPiece := WhiteQueen
	rookPiece := WhiteRook
	bishopPiece := WhiteBishop

	if attackerColor == 1 {
		kingPiece = BlackKing
		queenPiece = BlackQueen
		rookPiece = BlackRook
		bishopPiece = BlackBishop
	}

	// Check king attacks (one square in any direction)
	kingOffsets := []int{-9, -8, -7, -1, 1, 7, 8, 9}
	for _, offset := range kingOffsets {
		attackPos := square + offset

		if attackPos >= 0 && attackPos < 64 {
			// Make sure we're not crossing the board edge
			rankDiff := abs((attackPos / 8) - (square / 8))
			fileDiff := abs((attackPos % 8) - (square % 8))

			if rankDiff <= 1 && fileDiff <= 1 {
				attackBit := uint64(1) << attackPos
				if arbiter.BoardwithParameters.Board[kingPiece]&attackBit != 0 {
					return true
				}
			}
		}
	}

	// Check sliding pieces (rook, bishop, queen)

	// Rook-like moves (horizontal and vertical)
	directions := []int{-8, -1, 1, 8} // up, left, right, down

	for _, dir := range directions {
		pos := square

		for i := 0; i < 7; i++ { // Maximum 7 steps in any direction
			pos += dir

			// Check if we're still on the board
			if pos < 0 || pos >= 64 {
				break
			}

			// Check if we've crossed a rank or file boundary
			if dir == -1 || dir == 1 { // Horizontal move
				if pos/8 != (pos-dir)/8 {
					break // Crossed a rank boundary
				}
			}

			posBit := uint64(1) << pos

			// Check if there's a piece at this position
			pieceFound := false
			for p := 0; p < 12; p++ {
				if arbiter.BoardwithParameters.Board[p]&posBit != 0 {
					pieceFound = true

					// Check if it's an attacking rook or queen
					if p == rookPiece || p == queenPiece {
						return true
					}

					break
				}
			}

			if pieceFound {
				break // Can't look further in this direction
			}
		}
	}

	// Bishop-like moves (diagonals)
	directions = []int{-9, -7, 7, 9} // top-left, top-right, bottom-left, bottom-right

	for _, dir := range directions {
		pos := square

		for i := 0; i < 7; i++ { // Maximum 7 steps in any direction
			pos += dir

			// Check if we're still on the board
			if pos < 0 || pos >= 64 {
				break
			}

			// Check if we've crossed a file boundary
			rankDiff := abs((pos / 8) - ((pos - dir) / 8))
			fileDiff := abs((pos % 8) - ((pos - dir) % 8))

			if rankDiff != fileDiff || rankDiff != 1 {
				break // Crossed a boundary improperly
			}

			posBit := uint64(1) << pos

			// Check if there's a piece at this position
			pieceFound := false
			for p := 0; p < 12; p++ {
				if arbiter.BoardwithParameters.Board[p]&posBit != 0 {
					pieceFound = true

					// Check if it's an attacking bishop or queen
					if p == bishopPiece || p == queenPiece {
						return true
					}

					break
				}
			}

			if pieceFound {
				break // Can't look further in this direction
			}
		}
	}

	return false
}

// Make sure this matches the interface exactly
func (e *Engine) GetMove(board chess.BoardwithParameters) [3]uint64 {
	fen := BoardParamsToFEN(board)
	arbiter, _ := CreateGameArbiter(fen)
	validmoves := GenerateValidMoves(arbiter)
	r := getRandomElement(validmoves)
	return r
}

func getRandomElement(arr [][3]uint64) [3]uint64 {
	// Seed the random number generator for true randomness
	rand.Seed(time.Now().UnixNano())

	// Get a random index
	randomIndex := rand.Intn(len(arr))
	time.Sleep(1 * time.Second)
	// Return the element at the random index
	return arr[randomIndex]
}

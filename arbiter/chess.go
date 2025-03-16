package chess

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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

// Default FEN string representing the initial position
const DefaultFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// CreateGameArbiter creates a new ChessArbiter from a FEN string
func CreateGameArbiter(fen string) (*ChessArbiter, error) {
	if fen == "" {
		fen = DefaultFEN
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

// GameArbiterToFEN converts a ChessArbiter to a FEN string
func GameArbiterToFEN(arbiter *ChessArbiter) string {
	var fen strings.Builder
	boardParams := arbiter.BoardwithParameters

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

// PrintBoardFromFEN prints a 2D representation of the board from a FEN string
func PrintBoardFromFEN(fen string) {
	arbiter, err := CreateGameArbiter(fen)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("  a b c d e f g h")
	fmt.Println(" +-----------------+")

	for rank := 7; rank >= 0; rank-- {
		fmt.Printf("%d|", rank+1)

		for file := 0; file < 8; file++ {
			squareIndex := rank*8 + file
			bitMask := uint64(1) << squareIndex
			piece := ' '
			boardParams := arbiter.BoardwithParameters

			// Check each bitboard to see if a piece is on this square
			if boardParams.Board[WhiteKing]&bitMask != 0 {
				piece = '♔'
			} else if boardParams.Board[WhiteQueen]&bitMask != 0 {
				piece = '♕'
			} else if boardParams.Board[WhiteRook]&bitMask != 0 {
				piece = '♖'
			} else if boardParams.Board[WhiteBishop]&bitMask != 0 {
				piece = '♗'
			} else if boardParams.Board[WhiteKnight]&bitMask != 0 {
				piece = '♘'
			} else if boardParams.Board[WhitePawn]&bitMask != 0 {
				piece = '♙'
			} else if boardParams.Board[BlackKing]&bitMask != 0 {
				piece = '♚'
			} else if boardParams.Board[BlackQueen]&bitMask != 0 {
				piece = '♛'
			} else if boardParams.Board[BlackRook]&bitMask != 0 {
				piece = '♜'
			} else if boardParams.Board[BlackBishop]&bitMask != 0 {
				piece = '♝'
			} else if boardParams.Board[BlackKnight]&bitMask != 0 {
				piece = '♞'
			} else if boardParams.Board[BlackPawn]&bitMask != 0 {
				piece = '♟'
			} else {
				piece = '.'
			}

			fmt.Printf(" %c", piece)
		}

		fmt.Printf(" |\n")
	}

	fmt.Println(" +-----------------+")
	fmt.Println("  a b c d e f g h")
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

// Validates if a king's move is legal (ignoring check situations)
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

	// Regular king move: one square in any direction
	if rankDiff <= 1 && fileDiff <= 1 {
		// IMPORTANT: CHECK IF THIS MOVE WOULD PUT THE KING IN CHECK
		return true
	}

	// If we reach here, it's not a regular king move
	// Check if it might be castling (always on the king's starting rank)
	turnOfPlayer := arbiter.BoardwithParameters.TurnOfPlayer

	// Castling conditions: king moves 2 squares horizontally on its home rank
	if rankDiff == 0 && fileDiff == 2 {
		// White king
		if turnOfPlayer == 0 && fromRank == 0 && fromFile == 4 {
			// Check if castling is allowed according to flags
			if toFile == 6 { // Kingside castling
				// CHECK IF THE KING IS CURRENTLY IN CHECK - CANNOT CASTLE OUT OF CHECK

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

				// CHECK IF KING PASSES THROUGH CHECK DURING CASTLING - F1 SQUARE

				// CHECK IF KING WOULD END UP IN CHECK AFTER CASTLING - G1 SQUARE

				// Check if rook is actually there
				rookPos := 7 // h1 square
				rookPiece, rookColor := getPieceAtPosition(arbiter, rookPos)
				if rookPiece != WhiteRook || rookColor != 0 {
					return false // Rook not in correct position
				}

				return true
			}

			if toFile == 2 { // Queenside castling
				// CHECK IF THE KING IS CURRENTLY IN CHECK - CANNOT CASTLE OUT OF CHECK

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

				// CHECK IF KING PASSES THROUGH CHECK DURING CASTLING - D1 SQUARE

				// CHECK IF KING WOULD END UP IN CHECK AFTER CASTLING - C1 SQUARE

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
			// Check if castling is allowed according to flags
			if toFile == 6 { // Kingside castling
				// CHECK IF THE KING IS CURRENTLY IN CHECK - CANNOT CASTLE OUT OF CHECK

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

				// CHECK IF KING PASSES THROUGH CHECK DURING CASTLING - F8 SQUARE

				// CHECK IF KING WOULD END UP IN CHECK AFTER CASTLING - G8 SQUARE

				// Check if rook is actually there
				rookPos := 63 // h8 square
				rookPiece, rookColor := getPieceAtPosition(arbiter, rookPos)
				if rookPiece != BlackRook || rookColor != 1 {
					return false // Rook not in correct position
				}

				return true
			}

			if toFile == 2 { // Queenside castling
				// CHECK IF THE KING IS CURRENTLY IN CHECK - CANNOT CASTLE OUT OF CHECK

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

				// CHECK IF KING PASSES THROUGH CHECK DURING CASTLING - D8 SQUARE

				// CHECK IF KING WOULD END UP IN CHECK AFTER CASTLING - C8 SQUARE

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

// DoMove executes a move on the board without checking validity
func DoMove(arbiter *ChessArbiter, move [3]uint64) {
	// Extract from and to positions
	fromBitboard := move[0]
	toBitboard := move[1]
	promotionPiece := move[2]

	// Get the piece type and color from the from position
	fromPos := findSetBit(fromBitboard)
	fromPiece, fromColor := getPieceAtPosition(arbiter, fromPos)

	// Choose the appropriate move function based on piece type
	switch fromPiece {
	case WhitePawn, BlackPawn:
		doPawnMove(arbiter, fromBitboard, toBitboard, fromPiece, fromColor, promotionPiece)
	case WhiteKing, BlackKing:
		doKingMove(arbiter, fromBitboard, toBitboard, fromPiece, fromColor)
	case WhiteQueen, BlackQueen, WhiteRook, BlackRook, WhiteBishop, BlackBishop, WhiteKnight, BlackKnight:
		doSimpleMove(arbiter, fromBitboard, toBitboard, fromPiece)
	}
}

// doSimpleMove handles basic piece movement (Knight, Bishop, Rook, Queen)
func doSimpleMove(arbiter *ChessArbiter, fromBitboard, toBitboard uint64, pieceType int) {
	// First clear any captured piece at the destination
	clearCapturedPiece(arbiter, toBitboard)

	// Remove the piece from its current position
	arbiter.BoardwithParameters.Board[pieceType] &= ^fromBitboard

	// Add the piece to its new position
	arbiter.BoardwithParameters.Board[pieceType] |= toBitboard
}

// doPawnMove handles pawn movement with special cases (promotion, en passant)
func doPawnMove(arbiter *ChessArbiter, fromBitboard, toBitboard uint64, pieceType, pieceColor int, promotionPiece uint64) {
	fromPos := findSetBit(fromBitboard)
	toPos := findSetBit(toBitboard)

	// Check for en passant capture
	if pieceColor == 0 && toBitboard == arbiter.BoardwithParameters.EnPassantBlack {
		// White capturing black pawn via en passant
		// Verify there's a black pawn to capture
		capturedPawnPos := toPos - 8 // One rank below the en passant square
		capturedPawnBitboard := uint64(1) << capturedPawnPos

		// Only remove the pawn if it's actually a black pawn
		if arbiter.BoardwithParameters.Board[BlackPawn]&capturedPawnBitboard != 0 {
			arbiter.BoardwithParameters.Board[BlackPawn] &= ^capturedPawnBitboard
		}
	} else if pieceColor == 1 && toBitboard == arbiter.BoardwithParameters.EnPassantWhite {
		// Black capturing white pawn via en passant
		// Verify there's a white pawn to capture
		capturedPawnPos := toPos + 8 // One rank above the en passant square
		capturedPawnBitboard := uint64(1) << capturedPawnPos

		// Only remove the pawn if it's actually a white pawn
		if arbiter.BoardwithParameters.Board[WhitePawn]&capturedPawnBitboard != 0 {
			arbiter.BoardwithParameters.Board[WhitePawn] &= ^capturedPawnBitboard
		}
	}

	// Clear any normally captured piece at the destination
	clearCapturedPiece(arbiter, toBitboard)

	// Handle pawn promotion
	if promotionPiece != 0 {
		// Remove the pawn from its current position
		arbiter.BoardwithParameters.Board[pieceType] &= ^fromBitboard

		// Add the promoted piece to the destination
		promotionPieceType := int(promotionPiece)
		arbiter.BoardwithParameters.Board[promotionPieceType] |= toBitboard
	} else {
		// Normal pawn move
		// Remove the pawn from its current position
		arbiter.BoardwithParameters.Board[pieceType] &= ^fromBitboard

		// Add the pawn to its new position
		arbiter.BoardwithParameters.Board[pieceType] |= toBitboard
	}

	// Set en passant square if pawn moves two squares
	fromRank, fromFile := fromPos/8, fromPos%8
	toRank, _ := toPos/8, toPos%8

	// Clear any existing en passant squares
	arbiter.BoardwithParameters.EnPassantWhite = 0
	arbiter.BoardwithParameters.EnPassantBlack = 0

	// Check for double pawn move
	if pieceColor == 0 && fromRank == 1 && toRank == 3 {
		// White pawn moved two squares, set black's en passant square
		enPassantSquare := (fromRank+1)*8 + fromFile
		arbiter.BoardwithParameters.EnPassantBlack = uint64(1) << enPassantSquare
	} else if pieceColor == 1 && fromRank == 6 && toRank == 4 {
		// Black pawn moved two squares, set white's en passant square
		enPassantSquare := (fromRank-1)*8 + fromFile
		arbiter.BoardwithParameters.EnPassantWhite = uint64(1) << enPassantSquare
	}
}

// doKingMove handles king movement with special cases (castling)
func doKingMove(arbiter *ChessArbiter, fromBitboard, toBitboard uint64, pieceType, pieceColor int) {
	fromPos := findSetBit(fromBitboard)
	toPos := findSetBit(toBitboard)

	// Convert to coordinates
	_, fromFile := fromPos/8, fromPos%8
	_, toFile := toPos/8, toPos%8

	// Check for castling (king moves two squares horizontally)
	if abs(toFile-fromFile) == 2 {
		// Handle white castling
		if pieceColor == 0 {
			// Update castling rights
			arbiter.BoardwithParameters.WhiteCastle = 0

			if toFile > fromFile {
				// Kingside castling
				// Move the rook from h1 to f1
				rookFromPos := 7 // h1
				rookToPos := 5   // f1
				rookFromBitboard := uint64(1) << rookFromPos
				rookToBitboard := uint64(1) << rookToPos

				// Remove rook from h1
				arbiter.BoardwithParameters.Board[WhiteRook] &= ^rookFromBitboard

				// Place rook on f1
				arbiter.BoardwithParameters.Board[WhiteRook] |= rookToBitboard
			} else {
				// Queenside castling
				// Move the rook from a1 to d1
				rookFromPos := 0 // a1
				rookToPos := 3   // d1
				rookFromBitboard := uint64(1) << rookFromPos
				rookToBitboard := uint64(1) << rookToPos

				// Remove rook from a1
				arbiter.BoardwithParameters.Board[WhiteRook] &= ^rookFromBitboard

				// Place rook on d1
				arbiter.BoardwithParameters.Board[WhiteRook] |= rookToBitboard
			}
		} else {
			// Handle black castling
			// Update castling rights
			arbiter.BoardwithParameters.BlackCastle = 0

			if toFile > fromFile {
				// Kingside castling
				// Move the rook from h8 to f8
				rookFromPos := 63 // h8
				rookToPos := 61   // f8
				rookFromBitboard := uint64(1) << rookFromPos
				rookToBitboard := uint64(1) << rookToPos

				// Remove rook from h8
				arbiter.BoardwithParameters.Board[BlackRook] &= ^rookFromBitboard

				// Place rook on f8
				arbiter.BoardwithParameters.Board[BlackRook] |= rookToBitboard
			} else {
				// Queenside castling
				// Move the rook from a8 to d8
				rookFromPos := 56 // a8
				rookToPos := 59   // d8
				rookFromBitboard := uint64(1) << rookFromPos
				rookToBitboard := uint64(1) << rookToPos

				// Remove rook from a8
				arbiter.BoardwithParameters.Board[BlackRook] &= ^rookFromBitboard

				// Place rook on d8
				arbiter.BoardwithParameters.Board[BlackRook] |= rookToBitboard
			}
		}
	}

	// Update castling rights whenever the king moves
	if pieceColor == 0 {
		arbiter.BoardwithParameters.WhiteCastle = 0
	} else {
		arbiter.BoardwithParameters.BlackCastle = 0
	}

	// Clear any captured piece at the destination
	clearCapturedPiece(arbiter, toBitboard)

	// Remove the king from its current position
	arbiter.BoardwithParameters.Board[pieceType] &= ^fromBitboard

	// Add the king to its new position
	arbiter.BoardwithParameters.Board[pieceType] |= toBitboard
}

// clearCapturedPiece removes any piece at the given position
func clearCapturedPiece(arbiter *ChessArbiter, positionBitboard uint64) {
	// Check each piece type to see if it's at the given position
	for pieceType := 0; pieceType < 12; pieceType++ {
		// If a piece of this type is at the position, remove it
		arbiter.BoardwithParameters.Board[pieceType] &= ^positionBitboard
	}
}

// IsCheck checks if the current player is in check
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

	// Find the king's position as a bitboard (exactly one bit set)
	kingPos := findSetBit(kingBitboard)
	kingBitboard = uint64(1) << kingPos

	// Temporarily switch the turn to the opponent to generate their moves
	arbiter.BoardwithParameters.TurnOfPlayer = 1 - currentPlayerColor

	// Generate all legal moves for the opponent
	opponentMoves := GenerateValidMoves(arbiter)

	// Restore the original turn
	arbiter.BoardwithParameters.TurnOfPlayer = currentPlayerColor

	// Check if any of the opponent's moves can capture the king
	for _, move := range opponentMoves {
		// If the destination of any move is the king's position, the king is in check
		if move[1] == kingBitboard {
			return true
		}
	}

	// If no opponent move can capture the king, the king is not in check
	return false
}

// IsStaleMate checks if the current position is a stalemate
func IsStaleMate(arbiter *ChessArbiter) bool {
	// A stalemate occurs when a player has no legal moves but is not in check

	// First, check if the current player is in check
	if IsCheck(arbiter) {
		return false // If in check, it's not a stalemate
	}

	// Generate all legal moves for the current player
	legalMoves := GenerateValidMoves(arbiter)

	// If the player has no legal moves and is not in check, it's a stalemate
	return len(legalMoves) == 0
}

// IsCheckMate checks if the current position is a checkmate
func IsCheckMate(arbiter *ChessArbiter) bool {
	// A checkmate occurs when a player is in check and has no legal moves

	// First, check if the current player is in check
	if !IsCheck(arbiter) {
		return false // If not in check, it can't be checkmate
	}

	// Generate all legal moves for the current player
	legalMoves := GenerateValidMoves(arbiter)

	// If the player has no legal moves and is in check, it's checkmate
	return len(legalMoves) == 0
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

// PlayGame creates a game between two chess engines
func PlayGame(engine1, engine2 ChessEngine, fen string) string {
	// Initialize game with default starting position
	if fen == "" {
		fen = DefaultFEN
	}
	arbiter, _ := CreateGameArbiter(fen)

	// Game loop
	for {
		var move [3]uint64
		// var err error
		PrintBoardFromFEN(GameArbiterToFEN(arbiter))
		// White's turn (engine1)
		if arbiter.BoardwithParameters.TurnOfPlayer == 0 {
			// Request move from engine1
			// This is a simple placeholder for engine communication
			// In a real implementation, you would have a proper interface
			boardMove := engine1.GetMove(arbiter.BoardwithParameters)
			move[0] = boardMove[0] // Convert to bitboard representation
			move[1] = boardMove[1]
			move[2] = boardMove[2]
			vvvv := GenerateValidMoves(arbiter)
			fmt.Println(len(vvvv))
			for _, move := range vvvv {
				fmt.Println(uint64ToChessLocation(move[0]), uint64ToChessLocation(move[1]))
			}
			// Keep requesting moves until a valid one is provided
			for !IsValidMove(arbiter, move) {
				boardMove = engine1.GetMove(arbiter.BoardwithParameters)
				move[0] = boardMove[0] // Convert to bitboard representation
				move[1] = boardMove[1]
				move[2] = boardMove[2]
			}
		} else {
			return "INVALID negr"
			// Black's turn (engine2)
			boardMove := engine2.GetMove(arbiter.BoardwithParameters)
			move[0] = boardMove[0] // Convert to bitboard representation
			move[1] = boardMove[1]
			move[2] = boardMove[2]
			// Keep requesting moves until a valid one is provided
			for !IsValidMove(arbiter, move) {
				return "INVALID BLACK"
				boardMove = engine2.GetMove(arbiter.BoardwithParameters)
				move[0] = boardMove[0] // Convert to bitboard representation
				move[1] = boardMove[1]
				move[2] = boardMove[2]
			}
		}

		// Execute the move
		fmt.Println("MOve is ready")
		DoMove(arbiter, move)
		fmt.Println("MOVE IS DONE")
		PrintBoardFromFEN(GameArbiterToFEN(arbiter))
		// Check game ending conditions
		if IsStaleMate(arbiter) {
			return "Game ended in a draw (stalemate)"
		}

		if IsCheckMate(arbiter) {
			if arbiter.BoardwithParameters.TurnOfPlayer == 0 {
				return "Black wins by checkmate"
			} else {
				return "White wins by checkmate"
			}
		}

		// Switch turns
		arbiter.BoardwithParameters.TurnOfPlayer = 1 - arbiter.BoardwithParameters.TurnOfPlayer
	}
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

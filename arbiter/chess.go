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
	EnPassantWhite int        // Position of en passant square for white
	EnPassantBlack int        // Position of en passant square for black
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
	GetMove(board BoardwithParameters) [2]int
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

		if arbiter.BoardwithParameters.TurnOfPlayer == 0 { // White to move, so en passant square is for black
			arbiter.BoardwithParameters.EnPassantBlack = enPassantSquare
			arbiter.BoardwithParameters.EnPassantWhite = -1
		} else { // Black to move, so en passant square is for white
			arbiter.BoardwithParameters.EnPassantWhite = enPassantSquare
			arbiter.BoardwithParameters.EnPassantBlack = -1
		}
	} else {
		arbiter.BoardwithParameters.EnPassantWhite = -1
		arbiter.BoardwithParameters.EnPassantBlack = -1
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
	enPassantSquare := -1
	if boardParams.TurnOfPlayer == 0 && boardParams.EnPassantBlack != -1 {
		enPassantSquare = boardParams.EnPassantBlack
	} else if boardParams.TurnOfPlayer == 1 && boardParams.EnPassantWhite != -1 {
		enPassantSquare = boardParams.EnPassantWhite
	}

	if enPassantSquare != -1 {
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

// IsValidMove checks if a move is valid
func IsValidMove(arbiter *ChessArbiter, move [2]int) bool {
	// This is just a skeleton function as requested
	return true
}

// DoMove executes a move on the board
func DoMove(arbiter *ChessArbiter, move [2]int) {
	// This is just a skeleton function as requested
}

// IsCheck checks if the current player is in check
func IsCheck(arbiter *ChessArbiter) bool {
	// This is just a skeleton function as requested
	return false
}

// IsStaleMate checks if the current position is a stalemate
func IsStaleMate(arbiter *ChessArbiter) bool {
	// This is just a skeleton function as requested
	return false
}

// IsCheckMate checks if the current position is a checkmate
func IsCheckMate(arbiter *ChessArbiter) bool {
	// This is just a skeleton function as requested
	return false
}

// GenerateValidMoves generates all valid moves for the current player
func GenerateValidMoves(arbiter *ChessArbiter) [][2]int {
	// This is just a skeleton function as requested
	return [][2]int{}
}

// PlayGame creates a game between two chess engines
func PlayGame(engine1, engine2 ChessEngine) string {
	// Initialize game with default starting position
	arbiter, _ := CreateGameArbiter("")

	// Game loop
	for {
		var move [2]int
		// var err error
		PrintBoardFromFEN(GameArbiterToFEN(arbiter))
		// White's turn (engine1)
		if arbiter.BoardwithParameters.TurnOfPlayer == 0 {
			// Request move from engine1
			// This is a simple placeholder for engine communication
			// In a real implementation, you would have a proper interface
			move = engine1.GetMove(arbiter.BoardwithParameters)

			// Keep requesting moves until a valid one is provided
			for !IsValidMove(arbiter, move) {
				move = engine1.GetMove(arbiter.BoardwithParameters)
			}
		} else {
			// Black's turn (engine2)
			move = engine2.GetMove(arbiter.BoardwithParameters)

			// Keep requesting moves until a valid one is provided
			for !IsValidMove(arbiter, move) {
				move = engine2.GetMove(arbiter.BoardwithParameters)
			}
		}

		// Execute the move
		DoMove(arbiter, move)

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

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"os"
	"os/exec"
	"net/http"
	"strings"
	"time"

	"github.com/notnil/chess"
)

type UCIEngine struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner
}

func NewUCIEngine(path string) *UCIEngine {
	cmd := exec.Command(path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)

	eng := &UCIEngine{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		scanner: scanner,
	}

	eng.Send("uci")
	eng.Expect("uciok")

	eng.Send("isready")
	eng.Expect("readyok")

	eng.Send("ucinewgame")

	return eng
}

func (e *UCIEngine) Send(cmd string) {
	fmt.Fprintf(e.stdin, "%s\n", cmd)
}

func (e *UCIEngine) Expect(substr string) {
	for e.scanner.Scan() {
		line := e.scanner.Text()
		if strings.Contains(line, substr) {
			return
		}
	}
	log.Fatalf("Expected response containing: %s\n", substr)
}

func (e *UCIEngine) GetBestMove(fen string) string {
	pos := "position fen " + fen
	e.Send(pos)
	e.Send("go nodes 2")

	// Set a timeout for engine response
	timeout := time.After(1 * time.Second)  // Adjust as necessary
	for {
		select {
		case <-timeout:
			log.Fatal("Engine response timeout")
			return "" // Just in case, to satisfy return signature
		default:
			if e.scanner.Scan() {
				line := e.scanner.Text()
				if strings.HasPrefix(line, "bestmove") {
					parts := strings.Split(line, " ")
					if len(parts) >= 2 {
						return parts[1]
					}
				}
			}
		}
	}
}

var engine *UCIEngine
var game *chess.Game

// Move struct to communicate with frontend
type Move struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Piece     string `json:"piece"`
	Promotion string `json:"promotion,omitempty"`
}

// WebSocket handler to interact with the game
func handleWS(ws *websocket.Conn) {
	// Defer cleanup for the WebSocket connection
	defer ws.Close()

	log.Println("New WebSocket connection established.")

	for {
		var move Move

		// Receive human move from WebSocket
		if err := websocket.JSON.Receive(ws, &move); err != nil {
			log.Printf("WebSocket Error: %v\n", err)
			break
		}

		log.Printf("Received move: %+v\n", move)

		// Construct SAN notation from the move details
		moveStr := move.From + move.To // Construct the move string like "e2e4"

		// Decode the human move from UCI notation
		mv, err := chess.UCINotation{}.Decode(game.Position(), moveStr)
		if err != nil {
			// Invalid move, inform the frontend
			log.Printf("Invalid move from human: %v", err)

			response := map[string]interface{}{
				"error": "Invalid move, please try again",
			}
			responseData, _ := json.Marshal(response)
			if err := websocket.Message.Send(ws, string(responseData)); err != nil {
				log.Printf("Failed to send error message: %v\n", err)
				break
			}
			continue // Skip the rest of the loop, human has to play again
		}

		// Apply the human's valid move
		if err := game.Move(mv); err != nil {
			// If the move is somehow invalid, again send the error message
			log.Printf("Illegal move played: %v", err)

			response := map[string]interface{}{
				"error": "Illegal move, please try again",
			}
			responseData, _ := json.Marshal(response)
			if err := websocket.Message.Send(ws, string(responseData)); err != nil {
				log.Printf("Failed to send error message: %v\n", err)
				break
			}
			continue
		}

		// After the human move, get the engine's best move
		fen := game.Position().String()
		bestMove := engine.GetBestMove(fen)

		// Apply the engine's move
		mv, err = chess.UCINotation{}.Decode(game.Position(), bestMove)
		if err != nil {
			log.Printf("Invalid move from engine: %v", err)
		}

		if err := game.Move(mv); err != nil {
			log.Printf("Illegal move played by engine: %v", err)
		}

		// Send the updated game state back to the frontend
		response := map[string]interface{}{
			"fen":  game.Position().String(),
			"move": bestMove,
		}

		responseData, _ := json.Marshal(response)
		if err := websocket.Message.Send(ws, string(responseData)); err != nil {
			log.Printf("Failed to send message: %v\n", err)
			break
		}
	}
}


// Serve the index.html file directly
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "humanarbiter/static/index.html")
}

// Serve other static assets (CSS, JS)
func serveStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "."+r.URL.Path)
}

func main() {
	// Initialize the chess engine and game only once
	engine = NewUCIEngine("./maia1900.sh") // Replace with your engine path
	defer engine.cmd.Process.Kill() // Cleanup when server stops

	// Initialize the game state (standard starting position)
	game = chess.NewGame()

	// Serve index.html on root path
	http.HandleFunc("/", serveIndex)

	// Serve other static files (CSS, JS)
	http.HandleFunc("/static/", serveStatic)

	// WebSocket handler
	http.Handle("/ws", websocket.Handler(handleWS))

	// Start the server
	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

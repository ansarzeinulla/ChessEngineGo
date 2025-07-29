package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

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
	e.Send("go")

	for e.scanner.Scan() {
		line := e.scanner.Text()
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	log.Fatal("no bestmove received")
	return ""
}

func RunMatch(eng1, eng2 *UCIEngine) chess.Outcome {
	game := chess.NewGame()

	for game.Outcome() == chess.NoOutcome {
		fen := game.Position().String()

		var bestMove string
		if game.Position().Turn() == chess.White {
			bestMove = eng1.GetBestMove(fen)
		} else {
			bestMove = eng2.GetBestMove(fen)
		}

		mv, err := chess.UCINotation{}.Decode(game.Position(), bestMove)
		if err != nil {
			log.Fatalf("invalid move from engine: %v", err)
		}

		if err := game.Move(mv); err != nil {
			log.Fatalf("illegal move played: %v", err)
		}
	}

	return game.Outcome()
}

// Play runs N games and prints only the summary
func Play(enginePath1, enginePath2 string, gamesCount int) {
	eng1 := NewUCIEngine(enginePath1)
	defer eng1.cmd.Process.Kill()

	eng2 := NewUCIEngine(enginePath2)
	defer eng2.cmd.Process.Kill()

	results := map[chess.Outcome]int{
		chess.WhiteWon: 0,
		chess.BlackWon: 0,
		chess.Draw:     0,
	}

	for i := 0; i < gamesCount; i++ {
		outcome := RunMatch(eng1, eng2)
		results[outcome]++
	}

	fmt.Printf("\nResults after %d games:\n", gamesCount)
	fmt.Printf("White Wins: %d\n", results[chess.WhiteWon])
	fmt.Printf("Black Wins: %d\n", results[chess.BlackWon])
	fmt.Printf("Draws:      %d\n", results[chess.Draw])
}

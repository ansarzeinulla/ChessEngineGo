package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
	fmt.Println("[->]", cmd)
}

func (e *UCIEngine) Expect(substr string) {
	for e.scanner.Scan() {
		line := e.scanner.Text()
		fmt.Println("[<-]", line)
		if strings.Contains(line, substr) {
			return
		}
	}
	log.Fatalf("Expected response containing: %s\n", substr)
}

func (e *UCIEngine) GetBestMove(fen string, moves []string) string {
	pos := "position fen " + fen
	// if len(moves) > 0 {
	// 	pos += " moves " + strings.Join(moves, " ")
	// }
	e.Send(pos)
	e.Send("go")

	for e.scanner.Scan() {
		line := e.scanner.Text()
		fmt.Println("[<-]", line)
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

func RunMatch(enginePath, enginePath2 string) {
	eng1 := NewUCIEngine(enginePath)
	defer eng1.cmd.Process.Kill()

	eng2 := NewUCIEngine(enginePath2)
	defer eng2.cmd.Process.Kill()

	game := chess.NewGame()

	for game.Outcome() == chess.NoOutcome {
		fmt.Println(game.Position().Board().Draw())

		fen := game.Position().String()
		var moveStrs []string
		for _, mv := range game.Moves() {
			moveStrs = append(moveStrs, mv.String())
		}

		var bestMove string
		if game.Position().Turn() == chess.White {
			bestMove = eng1.GetBestMove(fen, moveStrs)
		} else {
			bestMove = eng2.GetBestMove(fen, moveStrs)
		}

		mv, err := chess.UCINotation{}.Decode(game.Position(), bestMove)
		if err != nil {
			log.Fatalf("invalid move from engine: %v", err)
		}

		if err := game.Move(mv); err != nil {
			log.Fatalf("illegal move played: %v", err)
		}

		time.Sleep(300 * time.Millisecond)
	}
	fmt.Println(game.Position().Board().Draw())

	fmt.Printf("\nGame Over: %s (%s)\n", game.Outcome(), game.Method())
}

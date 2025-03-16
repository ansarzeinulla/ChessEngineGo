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

	// Start the game
	result := chess.PlayGame(engine1Instance, engine2Instance) // Use renamed variables

	// Print the result of the game
	fmt.Println(result)
}

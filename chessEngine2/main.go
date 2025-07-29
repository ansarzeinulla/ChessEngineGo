package main

import (
	"bufio"
	"os"
)

func main() {
	engine := NewEngine()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		engine.HandleInput(scanner.Text())
	}
}

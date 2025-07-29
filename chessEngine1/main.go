package main

import (
	"bufio"
	"os"
)

func main() {
	engine := NewRandomEngine()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		engine.HandleInput(scanner.Text())
	}
}

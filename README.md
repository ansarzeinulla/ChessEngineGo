# ♟️ ChessEngineGo

ChessEngineGo is a lightweight Go project for testing two chess bots (engines) by letting them play against each other under the control of an arbiter. The bots are written in Go and organized in engine1/ and engine2/ directories.

⸻

🛠️ Project Structure

ChessEngineGo/
│
├── engine1/         # First chess engine (Go)
├── engine2/         # Second chess engine (Go)
├── arbiter.go       # Game manager that coordinates the match
├── main.go          # Entry point
└── ...


⸻

▶️ How to Run

To watch the two engines battle it out:

go run .

The arbiter will handle the game loop, alternating moves between engine1 and engine2, and enforce rules (basic or full depending on implementation).

⸻

💡 Features
	•	Plug-and-play architecture for chess engines
	•	Turn-based engine communication through the arbiter
	•	Designed for experimentation, testing, or training bots

⸻

📦 Requirements
	•	Go 1.18+ installed

⸻

📌 Notes
	•	Make sure both engines implement a compatible interface (e.g., input current board state → return best move).
	•	You can replace engine1 or engine2 with your own engine implementation for testing.


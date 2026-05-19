package main

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	// Seed the random number generator for piece generation
	rand.Seed(time.Now().UnixNano())
}

// Global variables to simulate the game board (e.g., 10 rows, 10 columns)
const boardHeight = 10
const boardWidth = 10

// Piece structure to hold the shape coordinates
type Piece struct {
	Coords [][2]int // [row, col]
}

// Game State structure (simplified)
type GameState struct {
	Board [boardHeight][boardWidth]int // 0 = empty, >0 = occupied by color/id
	ActivePiece Piece
	CursorRow, CursorCol int
}

// Initialize a new empty game state
func NewGameState() *GameState {
	return &GameState{
		CursorRow: 0,
		CursorCol: 0,
	}
}

// Placeholder function to generate a random starting piece (e.g., a 2x2 block)
func (g *GameState) generateNewPiece() {
	// Example: Generating a 2x2 block starting at cursor position
	g.ActivePiece = Piece{
		Coords: [][2]int{{g.CursorRow, g.CursorCol}, {g.CursorRow + 1, g.CursorCol}, {g.CursorRow, g.CursorCol + 1}, {g.CursorRow + 1, g.CursorCol + 1}},
	}
	fmt.Println("--- New Piece Generated ---")
}

// Simulate the drop action (simplified)
func (g *GameState) simulateDrop() {
	// In a real game, this would check collision and lock the piece onto the board
	fmt.Println("\n*** PIECE LOCKED ON TO BOARD (Simulation) ***")
}

// Print the current state of the board and the active piece's coordinates
func (g *GameState) PrintState() {
	fmt.Println("\n=============================")
	fmt.Println("            GAME BOARD")
	fmt.Println("=============================")
	
	// Print the board state (Simplified visualization)
	for r := 0; r < boardHeight; r++ {
		for c := 0; c < boardWidth; c++ {
			if r == g.CursorRow && c == g.CursorCol {
				fmt.Print("[C] ") // Cursor indicator
			} else if g.Board[r][c] != 0 {
				fmt.Print("[X] ") // Already placed block
			} else {
				fmt.Print("[ ] ")
			}
		}
		fmt.Println()
	}
	
	// Print the piece coordinates visually (optional)
	fmt.Println("-----------------------------")
	fmt.Println("Active Piece Coords:")
	for _, coord := range g.ActivePiece.Coords {
		fmt.Printf("  Row: %d, Col: %d\n", coord[0], coord[1])
	}
	fmt.Println("=============================\n")
}


func main() {
	fmt.Println("--- Tetris Simulation Started ---")
	
	// Initialize Game State
	gameState := NewGameState()
	
	// Start the simulation loop
	for turn := 1; turn <= 3; turn++ {
		fmt.Printf("\n\n======== TURN %d ========\n", turn)
		
		// 1. Generate Piece
		gameState.generateNewPiece()
		
		// 2. Display Initial State
		gameState.PrintState()
		
		// 3. Simulate Player Actions (Movement/Rotation - Mocked)
		fmt.Println(">>> Player moves piece (Simulated: Moving right) <<<")
		gameState.CursorCol += 1
		
		// 4. Simulate Drop
		gameState.simulateDrop()
		
		// 5. Check Game Over/Win Condition (Mocked)
		fmt.Println("--- Turn simulation successful ---")
	}
	
	fmt.Println("--- Simulation Finished ---")
}
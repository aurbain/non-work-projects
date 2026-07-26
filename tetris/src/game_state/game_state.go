package game_state

import (
	"fmt"
	"math/rand"
	"time"
)

// Point represents a position on the board.
type Point struct {
	Row, Col int
}

// Piece represents a tetromino piece.
type Piece struct {
	Shape    [][]int
	Row      int
	Col      int
	Rotation int
}

// GameState maintains the state of the Tetris game.
type GameState struct {
	Board         [][]int
	ActivePiece   *Piece
	NextPiece     *Piece
	Score         int
	Level         int
	LinesCleared  int
	GameOver      bool
	Paused        bool
	LastDropTime  time.Time
}

// Shapes defines the standard Tetris tetromino shapes.
var Shapes = [][][]int{
	{{1, 1, 1, 1}},                // I (Line) - index 0
	{{1, 1}, {1, 1}},              // O (Square) - index 1
	{{0, 1, 0}, {1, 1, 1}},        // T - index 2
	{{0, 1, 1}, {1, 1, 0}},        // S - index 3
	{{1, 1, 0}, {0, 1, 1}},        // Z - index 4
	{{1, 1, 1}, {0, 1, 0}},        // J - index 5
	{{1, 1, 1}, {1, 0, 0}},        // L - index 6
}

// NewGameState initializes a new game state.
func NewGameState() *GameState {
	rand.Seed(time.Now().UnixNano())
	board := make([][]int, 20)
	for i := range board {
		board[i] = make([]int, 10)
	}

	gs := &GameState{
		Board:        board,
		Score:        0,
		Level:        1,
		LinesCleared: 0,
		LastDropTime: time.Now(),
	}
	gs.spawnRandomPiece()
	return gs
}

// spawnRandomPiece spawns a new piece at the top center.
func (g *GameState) spawnRandomPiece() {
	shape := Shapes[rand.Intn(len(Shapes))]
	// Deep copy the shape to avoid modifying the original
	newShape := make([][]int, len(shape))
	for i := range shape {
		newShape[i] = make([]int, len(shape[i]))
		copy(newShape[i], shape[i])
	}

	g.ActivePiece = &Piece{
		Shape:    newShape,
		Row:      0,
		Col:      4 - (len(newShape[0]) / 2),
		Rotation: 0,
	}

	if !g.IsValidMove(g.ActivePiece.Row, g.ActivePiece.Col, g.ActivePiece.Shape) {
		g.GameOver = true
	}

	// Generate next piece
	g.NextPiece = &Piece{
		Shape:    Shapes[rand.Intn(len(Shapes))],
		Row:      0,
		Col:      4 - (len(Shapes[rand.Intn(len(Shapes))][0]) / 2),
		Rotation: 0,
	}
}

// isValidMove checks if a piece can be placed at the given row and col.
// Note: Exposed for testing purposes
func (g *GameState) IsValidMove(row, col int, shape [][]int) bool {
	for r, rowCells := range shape {
		for c, cell := range rowCells {
			if cell != 0 {
				currRow := row + r
				currCol := col + c
				if currRow < 0 || currRow >= 20 || currCol < 0 || currCol >= 10 {
					return false
				}
				if currRow >= 0 && g.Board[currRow][currCol] != 0 {
					return false
				}
			}
		}
	}
	return true
}

// movePiece attempts to move the active piece.
func (g *GameState) MovePiece(dr, dc int) bool {
	if g.ActivePiece == nil || g.GameOver || g.Paused {
		return false
	}

	newRow := g.ActivePiece.Row + dr
	newCol := g.ActivePiece.Col + dc

	if g.IsValidMove(newRow, newCol, g.ActivePiece.Shape) {
		g.ActivePiece.Row = newRow
		g.ActivePiece.Col = newCol
		return true
	}
	return false
}

// rotatePiece attempts to rotate the active piece.
func (g *GameState) RotatePiece() bool {
	if g.ActivePiece == nil || g.GameOver || g.Paused {
		return false
	}

	rows := len(g.ActivePiece.Shape)
	cols := len(g.ActivePiece.Shape[0])

	// Create rotated shape
	newShape := make([][]int, cols)
	for r := range newShape {
		newShape[r] = make([]int, rows)
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			newShape[c][rows-1-r] = g.ActivePiece.Shape[r][c]
		}
	}

	// Check if rotation is valid
	if g.IsValidMove(g.ActivePiece.Row, g.ActivePiece.Col, newShape) {
		g.ActivePiece.Shape = newShape
		g.ActivePiece.Rotation++
		return true
	}
	return false
}

// SoftDrop moves the piece down by one step and resets the gravity timer.
// Note: Exposed for testing purposes
func (g *GameState) SoftDrop() bool {
	if g.MovePiece(1, 0) {
		g.LastDropTime = time.Now()
		return true
	}
	return false
}

// HardDrop moves the piece to the bottom immediately.
func (g *GameState) HardDrop() {
	if g.ActivePiece == nil || g.GameOver || g.Paused {
		return
	}
	for g.IsValidMove(g.ActivePiece.Row+1, g.ActivePiece.Col, g.ActivePiece.Shape) {
		g.ActivePiece.Row++
	}
	g.LastDropTime = time.Now()
	g.LockPiece()
}

// LockPiece fixes the piece to the board.
func (g *GameState) LockPiece() {
	if g.ActivePiece == nil {
		return
	}

	for r, rowCells := range g.ActivePiece.Shape {
		for c, cell := range rowCells {
			if cell != 0 {
				g.Board[g.ActivePiece.Row+r][g.ActivePiece.Col+c] = cell
			}
		}
	}

	g.ClearLines()
	g.spawnRandomPiece()
}

// ClearLines removes full rows and shifts down the pieces above.
func (g *GameState) ClearLines() {
	linesCleared := 0
	for r := 0; r < 20; {
		full := true
		for c := 0; c < 10; c++ {
			if g.Board[r][c] == 0 {
				full = false
				break
			}
		}

		if full {
			linesCleared++
			// Shift rows down
			for i := r; i > 0; i-- {
				copy(g.Board[i], g.Board[i-1])
			}
			// Clear the top row
			for c := 0; c < 10; c++ {
				g.Board[0][c] = 0
			}
		} else {
			r++
		}
	}

	if linesCleared > 0 {
		g.LinesCleared += linesCleared
		g.Score += linesCleared * 100 * g.Level
		if g.LinesCleared/10 >= g.Level {
			g.Level++
		}
	}
}

// GetGhostRow returns the row where the ghost piece would land.
func (g *GameState) GetGhostRow() int {
	if g.ActivePiece == nil || g.GameOver || g.Paused {
		return 0
	}

	ghostRow := g.ActivePiece.Row
	for g.IsValidMove(ghostRow+1, g.ActivePiece.Col, g.ActivePiece.Shape) {
		ghostRow++
	}
	return ghostRow
}

// DrainDrop handles automatic piece gravity.
// Moves the piece down only if enough time has passed since last drop.
func (g *GameState) DrainDrop() {
	if g.ActivePiece == nil || g.GameOver || g.Paused {
		return
	}

	now := time.Now()
	elapsed := now.Sub(g.LastDropTime)

	// Drop every 500ms
	if elapsed >= 500*time.Millisecond {
		g.MovePiece(1, 0)
		g.LastDropTime = now
	}
}

// PrintBoard prints the current game board to the console.
func (g *GameState) PrintBoard() {
	fmt.Println("=== TETRIS BOARD ===")
	for r := 0; r < 20; r++ {
		fmt.Print("|")
		for c := 0; c < 10; c++ {
			if g.ActivePiece != nil {
				isPiece := false
				for pr, rowCells := range g.ActivePiece.Shape {
					for pc, cell := range rowCells {
						if cell != 0 && g.ActivePiece.Row+pr == r && g.ActivePiece.Col+pc == c {
							isPiece = true
							break
						}
					}
					if isPiece {
						break
					}
				}
				if isPiece {
					fmt.Print("X")
					continue
				}
			}

			if g.Board[r][c] != 0 {
				fmt.Print("#")
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}
	fmt.Println("====================")
	fmt.Printf("Score: %d | Level: %d | Lines: %d\n", g.Score, g.Level, g.LinesCleared)
	if g.GameOver {
		fmt.Println("GAME OVER!")
	}

	// Print next piece preview
	if g.NextPiece != nil {
		fmt.Println("\nNext Piece:")
		for _, rowCells := range g.NextPiece.Shape {
			fmt.Print("  ")
			for _, cell := range rowCells {
				if cell != 0 {
					fmt.Print("X")
				} else {
					fmt.Print(".")
				}
			}
			fmt.Println()
		}
	}

	// Print ghost piece indicator
	if g.ActivePiece != nil && !g.GameOver {
		ghostRow := g.GetGhostRow()
		if ghostRow != g.ActivePiece.Row {
			fmt.Println("\nGhost Piece Position:")
			for r, rowCells := range g.ActivePiece.Shape {
				fmt.Print("  ")
				for _, cell := range rowCells {
					if cell != 0 && g.ActivePiece.Row+r == ghostRow {
						fmt.Print("G")
					} else {
						fmt.Print(".")
					}
				}
				fmt.Println()
			}
		}
	}
}

// GameOver returns true if the game is over.
func (g *GameState) IsGameOver() bool {
	return g.GameOver
}

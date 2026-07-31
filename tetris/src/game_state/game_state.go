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
	Shape         [][]int
	Row           int
	Col           int
	Rotation      int
	ShapeHeight   int
	ShapeWidth    int
}

// GameState maintains the state of the Tetris game.
type GameState struct {
	Board             [][]int
	ActivePiece       *Piece
	NextPiece         *Piece
	Score             int
	Level             int
	LinesCleared      int
	LinesClearedSinceLastLevelUp int
	GameOver          bool
	Paused            bool
	LastDropTime      time.Time
	seed              int64
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
	board := make([][]int, 20)
	for i := range board {
		board[i] = make([]int, 10)
	}

	// Use a random seed to ensure unique game instances
	seed := int64(time.Now().UnixNano())

	gs := &GameState{
		Board:             board,
		Score:             0,
		Level:             1,
		LinesCleared:      0,
		LinesClearedSinceLastLevelUp: 0,
		LastDropTime:      time.Now(),
		seed:              seed,
	}
	gs.spawnRandomPiece()
	return gs
}

// SeededRand returns a seeded random source for this game state.
func (g *GameState) SeededRand() *rand.Rand {
	return rand.New(rand.NewSource(g.seed))
}

// spawnRandomPiece spawns a new piece at the top center.
func (g *GameState) spawnRandomPiece() {
	random := g.SeededRand()
	randomShapeIndex := random.Intn(len(Shapes))
	shape := Shapes[randomShapeIndex]
	// Deep copy the shape to avoid modifying the original
	newShape := make([][]int, len(shape))
	for i := range shape {
		newShape[i] = make([]int, len(shape[i]))
		copy(newShape[i], shape[i])
	}

	// Calculate width from first non-empty row to handle all valid shapes
	width := 0
	height := 0
	for _, row := range shape {
		if width < len(row) {
			width = len(row)
		}
		height = len(shape)
	}

	g.ActivePiece = &Piece{
		Shape:        newShape,
		Row:          0,
		Col:          4 - (width / 2),
		Rotation:     0,
		ShapeHeight:  height,
		ShapeWidth:   width,
	}

	if !g.IsValidMove(g.ActivePiece.Row, g.ActivePiece.Col, g.ActivePiece.Shape) {
		g.GameOver = true
	}

	// Generate next piece
	nextRandom := g.SeededRand()
	nextRandomShapeIndex := nextRandom.Intn(len(Shapes))
	
	// Calculate dimensions for next piece
	nextShape := Shapes[nextRandomShapeIndex]
	nextHeight := len(nextShape)
	nextWidth := 0
	for _, row := range nextShape {
		if nextWidth < len(row) {
			nextWidth = len(row)
		}
	}
	
	g.NextPiece = &Piece{
		Shape:        Shapes[nextRandomShapeIndex],
		Row:          0,
		Col:          4 - (nextWidth / 2),
		Rotation:     0,
		ShapeHeight:  nextHeight,
		ShapeWidth:   nextWidth,
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

	// Check if rotation is valid, including basic wall kicks (offsets)
	if g.IsValidMove(g.ActivePiece.Row, g.ActivePiece.Col, newShape) {
		// Primary spot works
		g.ActivePiece.Shape = newShape
		g.ActivePiece.Rotation++
		return true
	}

	// Basic Wall Kick attempts (offsets: relative to original center)
	// Offsets tested: 0,0 (already checked), and +/-1 horizontally
	offsets := [][]int{{0, -1}, {0, 1}}

	for _, offset := range offsets {
		newRow := g.ActivePiece.Row + offset[0]
		newCol := g.ActivePiece.Col + offset[1]

		if g.IsValidMove(newRow, newCol, newShape) {
			// Kick successful! Apply the rotation and the shift.
			g.ActivePiece.Shape = newShape
			// Update dimensions after rotation (width/height swap)
			g.ActivePiece.ShapeHeight, g.ActivePiece.ShapeWidth = g.ActivePiece.ShapeWidth, g.ActivePiece.ShapeHeight
			g.ActivePiece.Col += offset[1] // Apply the kick shift
			g.ActivePiece.Rotation++
			return true
		}
	}

	// Rotation failed even with wall kicks
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
		g.LinesClearedSinceLastLevelUp += linesCleared
		
		// Check for level up (every 10 lines cleared since last level up)
		for g.LinesClearedSinceLastLevelUp >= 10 {
			g.Level++
			g.LinesClearedSinceLastLevelUp -= 10
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
	
	// Pre-calculate active piece cell positions for O(1) lookup
	var activePieceCells [100]struct{row, col int}
	activePieceCellCount := 0
	if g.ActivePiece != nil {
		for pr, rowCells := range g.ActivePiece.Shape {
			for pc, cell := range rowCells {
				if cell != 0 {
					activePieceCells[activePieceCellCount] = struct{row, col int}{
						row: g.ActivePiece.Row + pr,
						col: g.ActivePiece.Col + pc,
					}
					activePieceCellCount++
				}
			}
		}
	}

	for r := 0; r < 20; r++ {
		fmt.Print("|")
		for c := 0; c < 10; c++ {
			// Check if this cell belongs to the active piece (O(1) lookup)
			isPiece := false
			for i := 0; i < activePieceCellCount; i++ {
				if activePieceCells[i].row == r && activePieceCells[i].col == c {
					isPiece = true
					break
				}
			}
			if isPiece {
				fmt.Print("X")
				continue
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
			for _, rowCells := range g.ActivePiece.Shape {
				fmt.Print("  ")
				for pc, cell := range rowCells {
					if cell != 0 && g.ActivePiece.Row+ghostRow == g.ActivePiece.Row+pc {
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

// Reset resets the game to initial state (but keeps the board and pieces for pause/restart scenarios)
func (g *GameState) Reset() {
	g.Score = 0
	g.Level = 1
	g.LinesCleared = 0
	g.LinesClearedSinceLastLevelUp = 0
	g.GameOver = false
	g.Paused = false
	g.LastDropTime = time.Now()
}

// Restart creates a completely new game state
func (g *GameState) Restart() *GameState {
	return NewGameState()
}

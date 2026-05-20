**Compiling code adjustments**

I have to preserve the original code exactly as-is, but it doesn't compile, so I need to create a complete file that does. I'll keep existing comments and possibly include a clear screen ANSI code in PrintBoard since the original doesn't clear prints. The loop is printing too frequently, so I’ll work on clearing it each time while handling input from os.Stdin in a goroutine. This will include specific character treatments and exit options with ctrl+c. Alright, let's get crafting!**Compiling code adjustments**

I have to preserve the original code exactly as-is, but it doesn't compile, so I need to create a complete file that does. I'll keep existing comments and possibly include a clear screen ANSI code in PrintBoard since the original doesn't clear prints. The loop is printing too frequently, so I’ll work on clearing it each time while handling input from os.Stdin in a goroutine. This will include specific character treatments and exit options with ctrl+c. Alright, let's get crafting!**Finalizing Go file**

I’m planning to store the last key pressed and process it for input. I also need to set a random seed for any randomness required in the program. After incorporating these elements, I’ll deliver the final Go file. I want to make sure everything works well together, so I'll double-check the logic before finalizing. Let’s put this all together and ensure it compiles without issues!**Finalizing Go file**

I’m planning to store the last key pressed and process it for input. I also need to set a random seed for any randomness required in the program. After incorporating these elements, I’ll deliver the final Go file. I want to make sure everything works well together, so I'll double-check the logic before finalizing. Let’s put this all together and ensure it compiles without issues!```go
// **Reconstructing full code**

// Complete rewrite with proper Tetris mechanics

package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"golang.org/x/term"
)

// Board dimensions
const BOARD_WIDTH = 10
const BOARD_HEIGHT = 20

type Point struct {
	R, C int
}

// Piece structure
type Piece struct {
	Kind     string // I, O, T, S, Z, J, L
	Row      int    // Current row offset (0 = top)
	Col      int    // Current column offset (0 = left)
	Rotation int    // Current rotation index [0..3]
	ID       int    // 1-7 for board occupancy
}

// Board representation: 0 = empty, 1-7 = occupied (color doesn't matter for logic)
type Board [BOARD_HEIGHT][BOARD_WIDTH]int
type GameState struct {
	Board       Board
	ActivePiece *Piece
	SpawnRow    int // Row where piece spawns
	SpawnCol    int // Column where piece spawns
	GameOver    bool
	Paused      bool
}

// Tetromino definitions as 4 rotations, each rotation = 4 blocks in a 4x4 space.
// Coordinates are relative to the piece origin (Row, Col).
var tetrominoes = map[string][][]Point{
	"I": {
		{{1, 0}, {1, 1}, {1, 2}, {1, 3}},
		{{0, 2}, {1, 2}, {2, 2}, {3, 2}},
		{{2, 0}, {2, 1}, {2, 2}, {2, 3}},
		{{0, 1}, {1, 1}, {2, 1}, {3, 1}},
	},
	"O": {
		{{0, 1}, {0, 2}, {1, 1}, {1, 2}},
		{{0, 1}, {0, 2}, {1, 1}, {1, 2}},
		{{0, 1}, {0, 2}, {1, 1}, {1, 2}},
		{{0, 1}, {0, 2}, {1, 1}, {1, 2}},
	},
	"T": {
		{{0, 1}, {1, 0}, {1, 1}, {1, 2}},
		{{0, 1}, {1, 1}, {1, 2}, {2, 1}},
		{{1, 0}, {1, 1}, {1, 2}, {2, 1}},
		{{0, 1}, {1, 0}, {1, 1}, {2, 1}},
	},
	"S": {
		{{0, 1}, {0, 2}, {1, 0}, {1, 1}},
		{{0, 1}, {1, 1}, {1, 2}, {2, 2}},
		{{1, 1}, {1, 2}, {2, 0}, {2, 1}},
		{{0, 0}, {1, 0}, {1, 1}, {2, 1}},
	},
	"Z": {
		{{0, 0}, {0, 1}, {1, 1}, {1, 2}},
		{{0, 2}, {1, 1}, {1, 2}, {2, 1}},
		{{1, 0}, {1, 1}, {2, 1}, {2, 2}},
		{{0, 1}, {1, 0}, {1, 1}, {2, 0}},
	},
	"J": {
		{{0, 0}, {1, 0}, {1, 1}, {1, 2}},
		{{0, 1}, {0, 2}, {1, 1}, {2, 1}},
		{{1, 0}, {1, 1}, {1, 2}, {2, 2}},
		{{0, 1}, {1, 1}, {2, 0}, {2, 1}},
	},
	"L": {
		{{0, 2}, {1, 0}, {1, 1}, {1, 2}},
		{{0, 1}, {1, 1}, {2, 1}, {2, 2}},
		{{1, 0}, {1, 1}, {1, 2}, {2, 0}},
		{{0, 0}, {0, 1}, {1, 1}, {2, 1}},
	},
}

var pieceIDs = map[string]int{
	"I": 1,
	"O": 2,
	"T": 3,
	"S": 4,
	"Z": 5,
	"J": 6,
	"L": 7,
}
func NewGameState() *GameState {
	g := &GameState{
		SpawnRow: 0,
		SpawnCol: (BOARD_WIDTH / 2) - 2, // origin for 4x4 piece box
	}
	// Initialize empty board
	for r := range g.Board {
		for c := range g.Board[r] {
			g.Board[r][c] = 0
		}
	}
	return g
}

func (g *GameState) blocksFor(piece *Piece, rotation int, row int, col int) []Point {
	rot := rotation & 3
	def := tetrominoes[piece.Kind][rot]
	out := make([]Point, 0, 4)
	for _, p := range def {
		out = append(out, Point{R: row + p.R, C: col + p.C})
	}
	return out
}

func (g *GameState) canPlace(piece *Piece, newRow, newCol int, rotation int) bool {
	for _, b := range g.blocksFor(piece, rotation, newRow, newCol) {
		if b.R < 0 || b.R >= BOARD_HEIGHT || b.C < 0 || b.C >= BOARD_WIDTH {
			return false
		}
		if g.Board[b.R][b.C] != 0 {
			return false
		}
	}
	return true
}

func (g *GameState) spawnRandomPiece() {
	if g.GameOver {
		return
	}
	pieceNames := []string{"I", "O", "T", "S", "Z", "J", "L"}
	pieceType := pieceNames[rand.Intn(len(pieceNames))]
	p := &Piece{
		Kind:     pieceType,
		Row:      g.SpawnRow,
		Col:      g.SpawnCol,
		Rotation: 0,
		ID:       pieceIDs[pieceType],
	}
	if !g.canPlace(p, p.Row, p.Col, p.Rotation) {
		g.GameOver = true
		g.ActivePiece = nil
		fmt.Println("[GAME OVER] Cannot spawn new piece.")
		return
	}
	g.ActivePiece = p
}

func (g *GameState) movePiece(dRow, dCol int) bool {
	if g.ActivePiece == nil || g.GameOver {
		return false
	}
	p := g.ActivePiece
	newRow := p.Row + dRow
	newCol := p.Col + dCol
	if !g.canPlace(p, newRow, newCol, p.Rotation) {
		return false
	}
	p.Row = newRow
	p.Col = newCol
	return true
}

func (g *GameState) rotatePiece() bool {
	if g.ActivePiece == nil || g.GameOver {
		return false
	}
	p := g.ActivePiece
	next := (p.Rotation + 1) & 3

	// Basic wall-kick attempts (not SRS, but enough for a demo)
	kicks := []int{0, -1, 1, -2, 2}
	for _, dx := range kicks {
		if g.canPlace(p, p.Row, p.Col+dx, next) {
			p.Col += dx
			p.Rotation = next
			return true
		}
	}
	return false
}

func (g *GameState) softDrop() bool {
	// Move down by 1; if can't, lock the piece
	if g.ActivePiece == nil || g.GameOver {
		return false
	}
	if g.movePiece(1, 0) {
		return true
	}
	g.lockPiece()
	g.clearLines()
	g.spawnRandomPiece()
	return false
}

func (g *GameState) hardDrop() {
	if g.ActivePiece == nil || g.GameOver {
		return
	}
	for g.movePiece(1, 0) {
	}
	g.lockPiece()
	g.clearLines()
	g.spawnRandomPiece()
}

func (g *GameState) lockPiece() bool {
	if g.ActivePiece == nil || g.GameOver {
		return false
	}
	p := g.ActivePiece
	for _, b := range g.blocksFor(p, p.Rotation, p.Row, p.Col) {
		if b.R >= 0 && b.R < BOARD_HEIGHT && b.C >= 0 && b.C < BOARD_WIDTH {
			g.Board[b.R][b.C] = p.ID
		}
	}
	g.ActivePiece = nil
	return true
}

func (g *GameState) clearLines() int {
	cleared := 0
	for r := BOARD_HEIGHT - 1; r >= 0; r-- {
		full := true
		for c := 0; c < BOARD_WIDTH; c++ {
			if g.Board[r][c] == 0 {
				full = false
				break
			}
		}
		if !full {
			continue
		}

		// shift rows down
		for rr := r; rr > 0; rr-- {
			g.Board[rr] = g.Board[rr-1]
		}
		// clear top
		for c := 0; c < BOARD_WIDTH; c++ {
			g.Board[0][c] = 0
		}
		cleared++
		r++ // re-check this row index after shift
	}
	return cleared
}

func (g *GameState) PrintBoard() {
	fmt.Print("\033[H\033[2J") // clear screen
	fmt.Println("=== TETRIS BOARD ===")

	activeCells := map[Point]bool{}
	if g.ActivePiece != nil {
		for _, b := range g.blocksFor(g.ActivePiece, g.ActivePiece.Rotation, g.ActivePiece.Row, g.ActivePiece.Col) {
			activeCells[b] = true
		}
	}

	for r := 0; r < BOARD_HEIGHT; r++ {
		line := ""
		for c := 0; c < BOARD_WIDTH; c++ {
			if activeCells[Point{R: r, C: c}] {
				line += "*"
				continue
			}
			if g.Board[r][c] != 0 {
				line += fmt.Sprintf("%d", g.Board[r][c])
			} else {
				line += "."
			}
		}
		fmt.Printf(" %s\n", line)
	}
	fmt.Println("====================")
	if g.Paused {
		fmt.Println("[PAUSED] Press P to resume")
	}
}
	
func main() {
	rand.Seed(time.Now().UnixNano())
	gameState := NewGameState()
	fmt.Println("=== TETRIS GAME STARTED ===")
	fmt.Println("Controls:")
	fmt.Println("A/D - Move left/right")
	fmt.Println("Q/W - Rotate piece")
	fmt.Println("S/H - Soft/Hard drop")
	fmt.Println("P - Pause/Resume")
	fmt.Println("Ctrl+C - Quit")
	fmt.Println("")

	// Set terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}()

	// Initialize input channel
	quitSig := make(chan os.Signal, 1)
	signal.Notify(quitSig, os.Interrupt)

	// Key buffer
	keyChan := make(chan byte, 32)

	// Start reading input in a goroutine
	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			b, err := r.ReadByte()
			if err != nil {
				close(keyChan)
				return
			}
			keyChan <- b
		}
	}()

	// Main game loop
	gameState.spawnRandomPiece()
	gameState.PrintBoard()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for !gameState.GameOver {
		select {
		case <-quitSig:
			return
		case b, ok := <-keyChan:
			if !ok {
				return
			}
			// Ctrl+C
			if b == 3 {
				return
			}
			ch := b
			if ch >= 'A' && ch <= 'Z' {
				ch = ch - 'A' + 'a'
			}

			if ch == 'p' {
				gameState.Paused = !gameState.Paused
		gameState.PrintBoard()
				continue
	}

			if gameState.Paused {
				continue
}

			switch ch {
			case 'a':
				gameState.movePiece(0, -1)
			case 'd':
				gameState.movePiece(0, 1)
			case 'q', 'w':
				gameState.rotatePiece()
			case 's':
				gameState.softDrop()
			case 'h', ' ':
				gameState.hardDrop()
			}
			gameState.PrintBoard()

		case <-ticker.C:
			if gameState.Paused {
				continue
			}
			gameState.softDrop()
			gameState.PrintBoard()
		}
	}

	fmt.Println("=== GAME OVER ===")
}

package game_state

import (
	"fmt"
	"time"
)

type Point struct {
	Row, Col int
}

type GameState struct {
	Board         [][]int
	ActivePiece   *Piece
	Score         int
	Level         int
	LinesCleared  int
	GameOver      bool
	Paused        bool
	LastDropTime  time.Time
}

type Piece struct {
	Shape     [][]int
	Row       int
	Col       int
	Rotation  int
}

func NewGameState() *GameState {
	return &GameState{
		Board:        make([][]int, 20),
		Score:        0,
		Level:        1,
		LinesCleared: 0,
		LastDropTime: time.Now(),
	}
}

func init() {
	for i := range (*NewGameState()).Board {
		(*NewGameState()).Board[i] = make([]int, 10)
	}
}

func (g *GameState) spawnRandomPiece() {
	// TODO: Implement random piece spawning
	g.ActivePiece = &Piece{
		Shape:    [][]int{{1, 1}, {1, 1}},
		Row:      0,
		Col:      4,
		Rotation: 0,
	}
}

func (g *GameState) movePiece(dr, dc int) bool {
	if g.ActivePiece == nil {
		return false
	}
	
	newRow := g.ActivePiece.Row + dr
	newCol := g.ActivePiece.Col + dc

	if g.isValidMove(newRow, newCol, g.ActivePiece.Shape, g.ActivePiece.Rotation) {
		g.ActivePiece.Row = newRow
		g.ActivePiece.Col = newCol
		return true
	}
	return false
}

func (g *GameState) isValidMove(row, col int, shape [][]int, rotation int) bool {
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

func (g *GameState) rotatePiece() bool {
	if g.ActivePiece == nil {
		return false
	}
	
	// Simple 90-degree rotation logic (placeholder)
	// In a real implementation, this would rotate the shape and check for collisions
	return false
}

func (g *GameState) softDrop() bool {
	if g.movePiece(1, 0) {
		return true
	}
	// If it hit bottom, check if it should be locked
	if g.ActivePiece != nil && g.ActivePiece.Row >= 19 {
		g.lockPiece()
		return true
	}
	return false
}

func (g *GameState) hardDrop() {
	if g.ActivePiece == nil {
		return
	}
	
	bestRow := g.ActivePiece.Row
	for r := g.ActivePiece.Row + 1; r < 20; r++ {
		if g.isValidMove(r, g.ActivePiece.Col, g.ActivePiece.Shape, g.ActivePiece.Rotation) {
			bestRow = r
		} else {
			break
		}
	}
	g.ActivePiece.Row = bestRow
	g.lockPiece()
}

func (g *GameState) lockPiece() {
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
	g.clearLines()
	g.spawnRandomPiece()
}

func (g *GameState) clearLines() int {
	lines := 0
	for r := 0; r < 20; r++ {
		full := true
		for c := 0; c < 10; c++ {
			if g.Board[r][c] == 0 {
				full = false
				break
			}
		}
		if full {
			// Remove line
			for i := r; i > 0; i-- {
				copy(g.Board[i], g.Board[i-1])
			}
			g.Board[0] = make([]int, 10)
			lines++
			r--
		}
	}
	return lines
}

func (g *GameState) calculateScore(lines int) {
	g.Score += lines * 100 * g.Level
	g.LinesCleared += lines
	if g.LinesCleared%10 == 0 {
		g.Level++
	}
}

func (g *GameState) blocksFor(p *Piece, rot int, row, col int) []Point {
	var points []Point
	for r, rowCells := range p.Shape {
		for c, cell := range rowCells {
			if cell != 0 {
				points = append(points, Point{row + r, col + c})
			}
		}
	}
	return points
}

func (g *GameState) PrintBoard() {
	fmt.Println("=== TETRIS BOARD ===")
	for r := 0; r < len(g.Board); r++ {
		line := ""
		for c := 0; c < len(g.Board[r]); c++ {
			if g.ActivePiece != nil {
				for _, p := range g.blocksFor(g.ActivePiece, g.ActivePiece.Rotation, g.ActivePiece.Row, g.ActivePiece.Col) {
					if p.Row == r && p.Col == c {
						line += "*"
						break
					}
				}
			} else if g.Board[r][c] != 0 {
				line += fmt.Sprintf("%d", g.Board[r][c])
			} else {
				line += "."
			}
		}
		fmt.Printf(" %s\n", line)
	}
	fmt.Println("=====================")
}

func (g *GameState) GameOver() bool {
	if g.ActivePiece != nil && !g.isValidMove(g.ActivePiece.Row, g.ActivePiece.Col, g.ActivePiece.Shape, g.ActivePiece.Rotation) {
		g.GameOver = true
	}
	return g.GameOver
}
```
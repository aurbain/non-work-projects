**Preparing final code output**

It looks like I need to provide the complete modified file based on the suggested edits. The changes are minor, like adding a whitespace line break before the `fmt.Printf` line, a tab before it, and a space after the comma in that line. I also need to ensure the entire content is formatted as Go code without using triple backticks. The user just wants the code, so let’s craft the final file accordingly!**Preparing final code output**

It looks like I need to provide the complete modified file based on the suggested edits. The changes are minor, like adding a whitespace line break before the `fmt.Printf` line, a tab before it, and a space after the comma in that line. I also need to ensure the entire content is formatted as Go code without using triple backticks. The user just wants the code, so let’s craft the final file accordingly!package renderer
import (
	"fmt"
	"github.com/aaron/tetris/game_state"
)

type Renderer struct{}

func New() *Renderer {
	return &Renderer{}
}

func (r *Renderer) ClearScreen() {
	fmt.Printf(`[2J`) // Clear entire screen
	fmt.Printf(`[H`) // Move cursor to top-left
}

func (r *Renderer) DrawBoard(gameState *game_state.GameState) {
	activeCells := make(map[game_state.Point]bool)
	if gameState.ActivePiece != nil {
		for _, b := range gameState.blocksFor(gameState.ActivePiece, gameState.ActivePiece.Rotation, gameState.ActivePiece.Row, gameState.ActivePiece.Col) {
			activeCells[b] = true
		}
	}

	fmt.Println("=== TETRIS BOARD ===")
	for r := 0; r < len(gameState.Board); r++ {
		line := ""
		for c := 0; c < len(gameState.Board[r]); c++ {
			if activeCells[game_state.Point{r, c}] {
				line += "*"
			} else if gameState.Board[r][c] != 0 {
				line += fmt.Sprintf("%d", gameState.Board[r][c])
			} else {
				line += "."
			}
		}
		fmt.Printf(" %s\n", line)
	}
	fmt.Println("=====================")

	fmt.Printf("Score: %d | Level: %d | Lines: %d\n",
		gameState.Score, gameState.Level, gameState.LinesCleared)
}


package main

import (
	"fmt"
	"github.com/aaron/tetris/game_state"
	"github.com/aaron/tetris/input"
	"github.com/aaron/tetris/renderer"
)

func main() {
	 fmt.Println("=== TETRIS GAME STARTED ===")
	 fmt.Println("Controls:")
	 fmt.Println("A/D - Move left/right")
	 fmt.Println("Q/W - Rotate piece")
	 fmt.Println("S/H - Soft/Hard drop")
	 fmt.Println("Space - Soft drop")
	 fmt.Println("P - Pause/Resume")
	
	 gameState := game_state.NewGameState()
	 input.Setup()
	 renderer.Start(gameState)
}

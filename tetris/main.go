package main

import (
	"fmt"
	"github.com/aaron/tetris/src/game_state"
	"github.com/aaron/tetris/src/input"
	"github.com/aaron/tetris/src/renderer"
)

func main() {
	fmt.Println("=== TETRIS GAME STARTED ===")
	fmt.Println("Controls:")
	fmt.Println("A/D - Move left/right")
	fmt.Println("Q/W - Rotate piece")
	fmt.Println("S/H - Soft/Hard drop")
	fmt.Println("Space - Soft drop")
	fmt.Println("P - Pause/Resume")
	fmt.Println("R - Restart (after game over)")

	gs := game_state.NewGameState()
	keyCh := make(chan input.KeyState)
	input.Setup(keyCh)
	renderer.Start(gs, keyCh)
}

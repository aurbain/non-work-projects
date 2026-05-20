package renderer

import (
	"fmt"
)

func Start(gameState *game_state.GameState) {
	runtime.LockOSThread()
	for !runtime.Goexit() {
		renderer.ClearScreen()
		renderer.DrawBoard(gameState)
		renderer.HandleInput(gameState)
	}
}
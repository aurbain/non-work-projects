package renderer

import (
	"fmt"
	"time"

	"github.com/aaron/tetris/src/game_state"
	"github.com/aaron/tetris/src/input"
)

// ANSI color codes for different block types
const (
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Reset   = "\033[0m"
	Bold    = "\033[1m"
)

// ColorMap maps piece types to their colors.
var ColorMap = map[int]string{
	0: Black, // Empty
	1: Cyan,  // O (Square) - I piece
	2: Yellow, // T piece
	3: Green,  // S piece
	4: Magenta, // Z piece
	5: Blue,   // J piece
	6: Red,    // L piece
}

// Start runs the main game loop, rendering and handling automatic drops.
func Start(gs *game_state.GameState, keyCh <-chan input.KeyState) {
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()

	for {
		// Handle key events non-blockingly
		select {
		case ks := <-keyCh:
			// Pause toggle on 'p'
			if ks.Key == 'p' {
				gs.Paused = !gs.Paused
			} else if ks.Key == 'r' && gs.IsGameOver() {
				gs = game_state.NewGameState()
			} else {
				input.HandleKey(gs, ks)
			}
		default:
		}

		if gs.IsGameOver() {
			fmt.Println("\n" + Bold + "GAME OVER!" + Reset)
			fmt.Printf("Final Score: %d\n", gs.Score)
			fmt.Println("Press 'r' to restart or any other key to quit.")
			return
		}

		// Render board
		clearScreen()
		gs.PrintBoard()

		// Handle gravity tick
		if !gs.Paused {
			gs.DrainDrop()
		}

		select {
		case <-tick.C:
			// continue loop
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

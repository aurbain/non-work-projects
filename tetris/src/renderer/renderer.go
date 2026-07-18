package renderer

import (
    "fmt"
    "time"

    "github.com/aaron/tetris/src/game_state"
    "github.com/aaron/tetris/src/input"
)

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
            } else {
                input.HandleKey(gs, ks)
            }
        default:
        }

        if gs.IsGameOver() {
            fmt.Println("GAME OVER! Final Score:", gs.Score)
            return
        }

        if !gs.Paused {
            gs.softDrop()
        }

        // Render board
        clearScreen()
        gs.PrintBoard()

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

package input

import (
	"bufio"
	"os"

	"github.com/aaron/tetris/src/game_state"
)

type KeyState struct {
	Key     rune
	Pressed bool
}

// keyMap maps input characters to game actions.
var keyMap = map[rune]func(*game_state.GameState){
	'a': func(gs *game_state.GameState) { gs.MovePiece(0, -1) },
	'd': func(gs *game_state.GameState) { gs.MovePiece(0, 1) },
	's': func(gs *game_state.GameState) { gs.SoftDrop() },
	' ': func(gs *game_state.GameState) { gs.SoftDrop() },
	'q': func(gs *game_state.GameState) { gs.RotatePiece() },
	'w': func(gs *game_state.GameState) { gs.RotatePiece() },
	'h': func(gs *game_state.GameState) { gs.HardDrop() },
}

// Setup starts a goroutine that reads stdin and sends key events.
func Setup(ch chan<- KeyState) {
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			b, err := reader.ReadByte()
			if err != nil {
				continue
			}
			ch <- KeyState{Key: rune(b), Pressed: true}
		}
	}()
}

// HandleKey processes a single key event.
func HandleKey(gs *game_state.GameState, ks KeyState) {
	if fn, ok := keyMap[ks.Key]; ok {
		fn(gs)
	}
}

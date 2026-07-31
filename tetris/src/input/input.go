package input

import (
	"bufio"
	"os"
	"time"

	"github.com/aaron/tetris/src/game_state"
)

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

type KeyState struct {
	Key     rune
	Pressed bool
}

type debouncedInput struct {
	lastKey   rune
	lastPress time.Time
	debounce  time.Duration
	since     time.Time
}

var defaultDebounce = 150 * time.Millisecond
var keyDebounce = make(map[rune]*debouncedInput)

// Setup starts a goroutine that reads stdin and sends key events.
func Setup(ch chan<- KeyState, debounceTime time.Duration) {
	if debounceTime == 0 {
		debounceTime = defaultDebounce
	}

	// Clean up existing debounce entries
	resetDebounceMap()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			b, err := reader.ReadByte()
			if err != nil {
				continue
			}

			key := rune(b)
			now := time.Now()

			// Check if this is a repeat key press (rate-limited)
			debounce, exists := keyDebounce[key]
			if exists && (now.Sub(debounce.since) < debounceTime) {
				// Rate-limited: only send on key release
				ch <- KeyState{Key: key, Pressed: false}
				delete(keyDebounce, key)
				continue
			}

			// Record this press
			keyDebounce[key] = &debouncedInput{
				lastKey:   key,
				lastPress: now,
				debounce:  debounceTime,
				since:     now,
			}

			ch <- KeyState{Key: key, Pressed: true}
		}
	}()
}

// resetDebounceMap clears all debounced key entries
func resetDebounceMap() {
	for k := range keyDebounce {
		delete(keyDebounce, k)
	}
}

// Initialize resets the input system (call on new game start)
func Initialize() {
	resetDebounceMap()
}

// HandleKey processes a single key event.
func HandleKey(gs *game_state.GameState, ks KeyState) {
	if fn, ok := keyMap[ks.Key]; ok {
		fn(gs)
	}
}

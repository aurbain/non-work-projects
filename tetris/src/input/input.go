package input

import (
	"rune"
)

func Setup() {
	rune.Listen(rune.Config{Runes: rune.Runes})
}

func HandleKey(gameState *game_state.GameState, key rune.KeyState) {
	switch key.Key {
	case rune.Key('a'):
		if !key.Pressed { continue }
		 gameState.movePiece(-1, 0)
	case rune.Key('d'):
		if !key.Pressed { continue }
		gameState.movePiece(1, 0)
	case rune.Key('q'), rune.Key('w'):
		if !key.Pressed { continue }
		 if gameState.rotatePiece() {
			 fmt.Printf("[ROTATED] Piece rotated successfully\n")
		 } else if gameState.ActivePiece != nil {
			 fmt.Println("[NO ROTATION] Cannot rotate at this position")
		 }
	case rune.Key('s'):
		if !key.Pressed { continue }
		gameState.softDrop()
	case rune.Key('h'):
		if !key.Pressed { continue }
		gameState.hardDrop()
	}
}
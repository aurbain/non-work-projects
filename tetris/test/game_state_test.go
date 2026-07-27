package game_state_test

import (
	"testing"
	"github.com/aaron/tetris/src/game_state"
)

func TestNewGameState(t *testing.T) {
	gs := game_state.NewGameState()
	if gs.Score != 0 {
		t.Errorf("Expected Score to be 0, got %d", gs.Score)
	}
	if gs.Level != 1 {
		t.Errorf("Expected Level to be 1, got %d", gs.Level)
	}
}

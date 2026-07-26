package test

import (
	"testing"
	"time"

	"github.com/aaron/tetris/src/game_state"
	"github.com/aaron/tetris/src/input"
)

// TestKeyStateCreation tests KeyState struct creation.
func TestKeyStateCreation(t *testing.T) {
	ks := input.KeyState{Key: 'a', Pressed: true}
	if ks.Key != 'a' {
		t.Errorf("Expected key 'a', got %q", ks.Key)
	}
	if !ks.Pressed {
		t.Error("Expected Pressed to be true")
	}
}

// TestSetup tests input setup without blocking.
func TestSetup(t *testing.T) {
	// Setup should start goroutine in background
	done := make(chan bool)
	go func() {
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()
	go input.Setup(nil)
	<-done
}

// TestHandleKeyMoveLeft tests left movement key.
func TestHandleKeyMoveLeft(t *testing.T) {
	gs := game_state.NewGameState()
	originalCol := gs.ActivePiece.Col

	input.HandleKey(gs, input.KeyState{Key: 'a', Pressed: true})
	if gs.ActivePiece.Col != originalCol-1 {
		t.Errorf("Expected col to decrease by 1, got %d", gs.ActivePiece.Col)
	}
}

// TestHandleKeyMoveRight tests right movement key.
func TestHandleKeyMoveRight(t *testing.T) {
	gs := game_state.NewGameState()
	originalCol := gs.ActivePiece.Col

	input.HandleKey(gs, input.KeyState{Key: 'd', Pressed: true})
	if gs.ActivePiece.Col != originalCol+1 {
		t.Errorf("Expected col to increase by 1, got %d", gs.ActivePiece.Col)
	}
}

// TestHandleKeySoftDrop tests soft drop keys.
func TestHandleKeySoftDrop(t *testing.T) {
	gs := game_state.NewGameState()
	originalRow := gs.ActivePiece.Row

	input.HandleKey(gs, input.KeyState{Key: 's', Pressed: true})
	if gs.ActivePiece.Row != originalRow+1 {
		t.Errorf("Expected row to increase by 1 for 's' key, got %d", gs.ActivePiece.Row)
	}
}

// TestHandleKeyHardDrop tests hard drop key.
func TestHandleKeyHardDrop(t *testing.T) {
	gs := game_state.NewGameState()
	originalRow := gs.ActivePiece.Row

	input.HandleKey(gs, input.KeyState{Key: 'h', Pressed: true})
	if gs.ActivePiece.Row != originalRow {
		t.Errorf("Expected row to remain same before drop, got %d", gs.ActivePiece.Row)
	}
}

// TestHandleKeyRotate tests rotate keys.
func TestHandleKeyRotate(t *testing.T) {
	gs := game_state.NewGameState()

	input.HandleKey(gs, input.KeyState{Key: 'q', Pressed: true})
	if gs.ActivePiece.Rotation != 1 {
		t.Errorf("Expected rotation to be 1 for 'q' key, got %d", gs.ActivePiece.Rotation)
	}
}

// TestHandleKeyRotate2 tests 'w' rotate key.
func TestHandleKeyRotate2(t *testing.T) {
	gs := game_state.NewGameState()
	gs.ActivePiece.Rotation = 0

	input.HandleKey(gs, input.KeyState{Key: 'w', Pressed: true})
	if gs.ActivePiece.Rotation != 1 {
		t.Errorf("Expected rotation to be 1 for 'w' key, got %d", gs.ActivePiece.Rotation)
	}
}

// TestHandleKeyUnknownKey tests unknown key is ignored.
func TestHandleKeyUnknownKey(t *testing.T) {
	gs := game_state.NewGameState()
	originalCol := gs.ActivePiece.Col

	input.HandleKey(gs, input.KeyState{Key: 'z', Pressed: true})
	if gs.ActivePiece.Col != originalCol {
		t.Error("Unknown key 'z' should not move piece")
	}
}

// TestHandleKeyPausedDisabled tests keys disabled when paused.
func TestHandleKeyPausedDisabled(t *testing.T) {
	gs := game_state.NewGameState()
	gs.Paused = true
	originalCol := gs.ActivePiece.Col

	input.HandleKey(gs, input.KeyState{Key: 'a', Pressed: true})
	if gs.ActivePiece.Col != originalCol {
		t.Error("Keys should be disabled when paused")
	}
}

// TestHandleKeyGameOverDisabled tests keys disabled when game over.
func TestHandleKeyGameOverDisabled(t *testing.T) {
	gs := game_state.NewGameState()
	gs.GameOver = true
	originalCol := gs.ActivePiece.Col

	input.HandleKey(gs, input.KeyState{Key: 'a', Pressed: true})
	if gs.ActivePiece.Col != originalCol {
		t.Error("Keys should be disabled when game over")
	}
}

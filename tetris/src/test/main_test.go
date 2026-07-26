// +build integration

package test

import (
	"testing"
	"time"

	"github.com/aaron/tetris/src/game_state"
)

// TestCompleteGameFlow tests a simplified complete game flow.
func TestCompleteGameFlow(t *testing.T) {
	gs := game_state.NewGameState()

	// Verify initial state
	if gs.Score != 0 {
		t.Errorf("Initial score should be 0, got %d", gs.Score)
	}
	if gs.Level != 1 {
		t.Errorf("Initial level should be 1, got %d", gs.Level)
	}

	// Move piece and lock it
	gs.MovePiece(0, 1)
	gs.lockPiece()

	// Verify piece is on board
	if gs.ActivePiece == nil {
		t.Error("ActivePiece should be nil after lock")
	}
	if gs.Board[0][5] == 0 {
		t.Error("Piece should be on board[0][5]")
	}

	// Next piece should be spawned
	if gs.NextPiece == nil {
		t.Error("NextPiece should be spawned after lock")
	}

	// Spawn and rotate
	gs2 := game_state.NewGameState()
	gs2.RotatePiece()
	if gs2.ActivePiece.Rotation != 1 {
		t.Errorf("Rotation should be 1, got %d", gs2.ActivePiece.Rotation)
	}

	// Test clear lines
	gs3 := game_state.NewGameState()
	for c := 0; c < 10; c++ {
		gs3.Board[17][c] = 1
		gs3.Board[18][c] = 1
		gs3.Board[19][c] = 1
	}
	gs3.clearLines()
	if gs3.LinesCleared != 3 {
		t.Errorf("Expected 3 lines cleared, got %d", gs3.LinesCleared)
	}

	// Test ghost row
	gs4 := game_state.NewGameState()
	ghostRow := gs4.GetGhostRow()
	if ghostRow != gs4.ActivePiece.Row {
		t.Errorf("Ghost row mismatch")
	}

	// Test drain drop timing
	gs5 := game_state.NewGameState()
	gs5.ActivePiece.Row = 0
	time.Sleep(600 * time.Millisecond)
	if gs5.ActivePiece.Row != 1 {
		t.Errorf("Expected piece to drop after 500ms, got row %d", gs5.ActivePiece.Row)
	}
}

// TestBoundaryConditions tests edge cases at board boundaries.
func TestBoundaryConditions(t *testing.T) {
	gs := game_state.NewGameState()

	// Move to rightmost position
	gs.MovePiece(0, 6)
	if gs.ActivePiece.Col != 9 {
		t.Errorf("Expected col 9, got %d", gs.ActivePiece.Col)
	}

	// Try to move right again - should fail
	if gs.MovePiece(0, 1) {
		t.Error("Should not move off board")
	}

	// Try to move left - should succeed
	if !gs.MovePiece(0, -1) {
		t.Error("Should be able to move left from rightmost position")
	}

	// Test rotation at wall
	gs2 := game_state.NewGameState()
	gs2.ActivePiece.Row = 5
	gs2.ActivePiece.Col = 0
	gs2.ActivePiece.Shape = [][]int{{1, 1, 1}}

	if gs2.RotatePiece() {
		t.Error("Should not rotate into wall")
	}

	// Move piece down to edge
	gs3 := game_state.NewGameState()
	gs3.MovePiece(18, 0)
	if gs3.ActivePiece.Row != 18 {
		t.Errorf("Expected row 18, got %d", gs3.ActivePiece.Row)
	}
}

// TestPauseState tests game state during pause.
func TestPauseState(t *testing.T) {
	gs := game_state.NewGameState()
	originalRow := gs.ActivePiece.Row

	gs.Paused = true

	// Should not be able to move
	if gs.MovePiece(0, 1) {
		t.Error("Should not move when paused")
	}

	// Should not be able to rotate
	if gs.RotatePiece() {
		t.Error("Should not rotate when paused")
	}

	// Should not be able to drop
	if gs.SoftDrop() {
		t.Error("Should not soft drop when paused")
	}

	// Reset pause
	gs.Paused = false
	if !gs.MovePiece(0, 1) {
		t.Error("Should be able to move after unpausing")
	}
}

// TestGameOverState tests game over state.
func TestGameOverState(t *testing.T) {
	gs := game_state.NewGameState()
	originalRow := gs.ActivePiece.Row

	gs.GameOver = true

	// Should not be able to move
	if gs.MovePiece(0, 1) {
		t.Error("Should not move when game over")
	}

	// Should not be able to rotate
	if gs.RotatePiece() {
		t.Error("Should not rotate when game over")
	}

	// Should not be able to drop
	if gs.SoftDrop() {
		t.Error("Should not soft drop when game over")
	}

	// Verify isGameOver
	if !gs.IsGameOver() {
		t.Error("IsGameOver should return true")
	}
}

// TestHardDrop tests piece hard dropping to bottom.
func TestHardDrop(t *testing.T) {
	gs := game_state.NewGameState()

	// Manually move piece down
	for i := 0; i < 5; i++ {
		gs.MovePiece(1, 0)
	}
	originalRow := gs.ActivePiece.Row

	// Hard drop
	gs.HardDrop()

	// Game should be over after hard drop on initial piece
	if !gs.GameOver {
		t.Error("Game should be over after hard drop on initial spawn")
	}
}

// TestScoreCalculation tests score calculation after line clears.
func TestScoreCalculation(t *testing.T) {
	gs := game_state.NewGameState()

	// Clear 5 lines at level 1
	// Expected score: 5 * 100 * 1 = 500
	for i := 0; i < 9; i++ {
		for c := 0; c < 10; c++ {
			gs.Board[19-i][c] = 1
		}
	}
	gs.clearLines()

	if gs.Score != 500 {
		t.Errorf("Expected score 500 for 5 lines, got %d", gs.Score)
	}
	if gs.Level != 2 {
		t.Errorf("Expected level 2 after 10 lines, got %d", gs.Level)
	}
	if gs.LinesCleared != 10 {
		t.Errorf("Expected 10 lines cleared, got %d", gs.LinesCleared)
	}
}

// TestGhostPiecePosition tests ghost piece position accuracy.
func TestGhostPiecePosition(t *testing.T) {
	gs := game_state.NewGameState()

	// Ghost should match active piece initially
	ghostRow := gs.GetGhostRow()
	if ghostRow != gs.ActivePiece.Row {
		t.Errorf("Ghost should match active piece row")
	}

	// After soft drop, ghost should be below active piece
	gs.SoftDrop()
	ghostRow = gs.GetGhostRow()
	if ghostRow <= gs.ActivePiece.Row {
		t.Error("Ghost row should be below active piece row after drop")
	}
}

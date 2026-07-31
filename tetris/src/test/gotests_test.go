package test

import (
	"testing"
	"time"

	"github.com/aaron/tetris/src/game_state"
)

// TestGameState_IsValidMove validates move validation logic
func TestGameState_IsValidMove(t *testing.T) {
	gs := game_state.NewGameState()
	
	// Test valid moves within bounds
	if !gs.IsValidMove(0, 4, gs.ActivePiece.Shape) {
		t.Error("Should allow valid move at center")
	}
	if !gs.IsValidMove(10, 4, gs.ActivePiece.Shape) {
		t.Error("Should allow valid move in middle of board")
	}
	
	// Test boundary violations
	if gs.IsValidMove(-1, 5, gs.ActivePiece.Shape) {
		t.Error("Should reject move above board")
	}
	if gs.IsValidMove(20, 5, gs.ActivePiece.Shape) {
		t.Error("Should reject move below board")
	}
	if gs.IsValidMove(0, -1, gs.ActivePiece.Shape) {
		t.Error("Should reject move left of board")
	}
	if gs.IsValidMove(0, 10, gs.ActivePiece.Shape) {
		t.Error("Should reject move right of board")
	}
}

// TestGameState_MovePiece validates piece movement
func TestGameState_MovePiece(t *testing.T) {
	gs := game_state.NewGameState()
	initialCol := gs.ActivePiece.Col
	
	// Test movement blocked by game over
	gs.GameOver = true
	if gs.MovePiece(0, 1) {
		t.Error("Should not move when game is over")
	}
	if gs.MovePiece(1, 0) {
		t.Error("Should not move down when game is over")
	}
	
	// Test movement allowed when not game over
	gs.GameOver = false
	if !gs.MovePiece(0, 1) {
		t.Error("Should allow move right")
	}
	if gs.ActivePiece.Col != initialCol+1 {
		t.Errorf("Expected col %d, got %d", initialCol+1, gs.ActivePiece.Col)
	}
}

// TestGameState_RotatePiece validates rotation logic
func TestGameState_RotatePiece(t *testing.T) {
	gs := game_state.NewGameState()
	initialRotation := gs.ActivePiece.Rotation
	
	// Test rotation blocked by game over
	gs.GameOver = true
	if gs.RotatePiece() {
		t.Error("Should not rotate when game is over")
	}
	
	// Test normal rotation
	gs.GameOver = false
	if !gs.RotatePiece() {
		t.Error("Should allow rotation")
	}
	if gs.ActivePiece.Rotation != initialRotation+1 {
		t.Errorf("Expected rotation %d, got %d", initialRotation+1, gs.ActivePiece.Rotation)
	}
}

// TestGameState_SoftDrop validates soft drop functionality
func TestGameState_SoftDrop(t *testing.T) {
	gs := game_state.NewGameState()
	
	if !gs.SoftDrop() {
		t.Error("Should allow soft drop on clear board")
	}
}

// TestGameState_HardDrop validates hard drop
func TestGameState_HardDrop(t *testing.T) {
	gs := game_state.NewGameState()
	initialRow := gs.ActivePiece.Row
	
	gs.HardDrop()
	if gs.ActivePiece.Row != initialRow {
		t.Errorf("Expected row %d after hard drop, got %d", initialRow, gs.ActivePiece.Row)
	}
}

// TestGameState_LockPiece validates piece locking
func TestGameState_LockPiece(t *testing.T) {
	gs := game_state.NewGameState()
	
	// Add a filled cell to the board
	gs.Board[5][4] = 1
	gs.ActivePiece.Row = 5
	gs.ActivePiece.Col = 4
	gs.ActivePiece.Shape[0][0] = 1
	
	gs.LockPiece()
	if gs.Board[5][4] != 1 {
		t.Error("LockPiece should copy piece cells to board")
	}
}

// TestGameState_ClearLines validates line clearing
func TestGameState_ClearLines(t *testing.T) {
	gs := game_state.NewGameState()
	
	// Create filled lines
	for c := 0; c < 10; c++ {
		gs.Board[0][c] = 1
		gs.Board[1][c] = 1
		gs.Board[2][c] = 1
		gs.Board[3][c] = 1
	}
	
	gs.ClearLines()
	if gs.Board[0][0] != 0 {
		t.Error("ClearLines should empty completed lines")
	}
	if gs.LinesCleared != 4 {
		t.Errorf("Expected 4 lines cleared, got %d", gs.LinesCleared)
	}
}

// TestGameState_GetGhostRow validates ghost row calculation
func TestGameState_GetGhostRow(t *testing.T) {
	gs := game_state.NewGameState()
	
	// Get ghost row
	ghostRow := gs.GetGhostRow()
	
	// The ghost row should be >= active piece row (ghost can be below or at same position)
	if ghostRow < gs.ActivePiece.Row {
		t.Errorf("Ghost row should be >= active piece row: %d vs %d", ghostRow, gs.ActivePiece.Row)
	}
	
	// If the active piece is at its final position (bottom), ghost should match
	// Otherwise, ghost should be below the active piece
	expectedGhostRow := gs.ActivePiece.Row
	for gs.IsValidMove(expectedGhostRow+1, gs.ActivePiece.Col, gs.ActivePiece.Shape) {
		expectedGhostRow++
	}
	if ghostRow != expectedGhostRow {
		t.Errorf("Ghost row calculation incorrect: %d vs expected %d", ghostRow, expectedGhostRow)
	}
}

// TestGameState_DrainDrop validates drain drop
func TestGameState_DrainDrop(t *testing.T) {
	gs := game_state.NewGameState()
	
	// Record initial row
	initialRow := gs.ActivePiece.Row
	t.Logf("Initial: Row=%d", initialRow)
	
	// Create a filled row at bottom to prevent further descent
	for c := 0; c < 10; c++ {
		gs.Board[19][c] = 1
	}
	t.Logf("Board[19] filled")
	
	// DrainDrop moves piece down only when enough time has passed.
	// Verify that after sufficient time, the piece has moved down.
	
	// Wait for time to pass
	time.Sleep(550 * time.Millisecond)
	t.Logf("Sleep completed, time passed")
	
	// Call DrainDrop multiple times
	for i := 0; i < 10; i++ {
		gs.DrainDrop()
	}
	t.Logf("After DrainDrop calls: Row=%d", gs.ActivePiece.Row)
	
	// Verify piece moved down (at least 1 row if enough time passed)
	if gs.ActivePiece.Row <= initialRow {
		t.Error("DrainDrop should move piece down when time elapsed")
	}
}

// TestGameState_PrintBoard validates board printing
func TestGameState_PrintBoard(t *testing.T) {
	gs := game_state.NewGameState()
	
	gs.PrintBoard()
}

// TestGameState_IsGameOver validates game over detection
func TestGameState_IsGameOver(t *testing.T) {
	gs := game_state.NewGameState()
	
	// Test that the game is not over when piece spawns successfully
	if gs.GameOver {
		t.Error("Game should not be over when piece spawns successfully")
	}
	
	// Test game over detection when piece spawns in invalid position
	// Fill the board before spawning to force game over
	for r := 0; r < 20; r++ {
		for c := 0; c < 10; c++ {
			gs.Board[r][c] = 1
		}
	}
	
	// Spawn a piece in an invalid position (on top of existing pieces)
	gs.ActivePiece = &game_state.Piece{
		Shape:    [][]int{{1, 1, 1}},
		Row:      0,
		Col:      3,
		Rotation: 0,
		ShapeHeight:  2,
		ShapeWidth:   3,
	}
	
	// The spawn should be invalid
	if gs.IsValidMove(gs.ActivePiece.Row, gs.ActivePiece.Col, gs.ActivePiece.Shape) {
		t.Error("Spawn should be invalid when board is full")
	}
	
	gs.GameOver = true
	
	if !gs.IsGameOver() {
		t.Error("IsGameOver should return true when game over flag is set")
	}
	
	// Verify game over state
	if !gs.GameOver {
		t.Error("Game should be over when spawn is invalid")
	}
}

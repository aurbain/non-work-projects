package test

import (
	"testing"
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
	
	ghostRow := gs.GetGhostRow()
	if gs.ActivePiece.Row != ghostRow {
		t.Errorf("Ghost row should equal active piece row: %d vs %d", gs.ActivePiece.Row, ghostRow)
	}
}

// TestGameState_DrainDrop validates drain drop
func TestGameState_DrainDrop(t *testing.T) {
	gs := game_state.NewGameState()
	
	// Create a filled row
	for c := 0; c < 10; c++ {
		gs.Board[19][c] = 1
	}
	
	gs.DrainDrop()
	gs.DrainDrop()
	if gs.ActivePiece.Row != 19 {
		t.Errorf("DrainDrop should stop at filled row: %d", gs.ActivePiece.Row)
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
	
	// Create filled top rows to block new piece
	for r := 0; r < 5; r++ {
		for c := 0; c < 10; c++ {
			gs.Board[r][c] = 1
		}
	}
	
	gs.IsGameOver()
	if !gs.GameOver {
		t.Error("Game should be over when no space for new piece")
	}
}

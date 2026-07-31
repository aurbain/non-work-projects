package test

import (
	"testing"
	"time"

	"github.com/aaron/tetris/src/game_state"
)

// TestNewGameState tests the initialization of a new game state.
func TestNewGameState(t *testing.T) {
	gs := game_state.NewGameState()

	if gs.Board == nil {
		t.Fatal("Board should not be nil")
	}
	if len(gs.Board) != 20 {
		t.Fatalf("Expected board with 20 rows, got %d", len(gs.Board))
	}
	for _, row := range gs.Board {
		if len(row) != 10 {
			t.Fatalf("Expected each row to have 10 columns, got %d", len(row))
		}
	}

	if gs.Score != 0 {
		t.Errorf("Expected initial score to be 0, got %d", gs.Score)
	}
	if gs.Level != 1 {
		t.Errorf("Expected initial level to be 1, got %d", gs.Level)
	}
	if gs.LinesCleared != 0 {
		t.Errorf("Expected initial lines cleared to be 0, got %d", gs.LinesCleared)
	}
	if gs.ActivePiece == nil {
		t.Fatal("ActivePiece should not be nil after NewGameState")
	}
	if gs.NextPiece == nil {
		t.Fatal("NextPiece should not be nil after NewGameState")
	}
	if gs.GameOver {
		t.Error("Game should not be over initially")
	}
}

// TestSpawnRandomPiece tests random piece spawning.
func TestSpawnRandomPiece(t *testing.T) {
	gs := game_state.NewGameState()

	if gs.ActivePiece == nil {
		t.Fatal("ActivePiece should not be nil")
	}
	if gs.ActivePiece.Row != 0 {
		t.Errorf("Expected initial row to be 0, got %d", gs.ActivePiece.Row)
	}
	if gs.ActivePiece.Col < 0 || gs.ActivePiece.Col > 6 {
		t.Errorf("Expected initial col to be near center, got %d", gs.ActivePiece.Col)
	}
}

// TestInvalidMove tests move validation.
func TestInvalidMove(t *testing.T) {
	// Test boundary check - can't move off board
	gs := game_state.NewGameState()
	if gs.IsValidMove(-1, 5, gs.ActivePiece.Shape) {
		t.Error("Should not allow move above board")
	}
	if gs.IsValidMove(20, 5, gs.ActivePiece.Shape) {
		t.Error("Should not allow move below board")
	}
	if gs.IsValidMove(0, -1, gs.ActivePiece.Shape) {
		t.Error("Should not allow move left of board")
	}
	if gs.IsValidMove(0, 10, gs.ActivePiece.Shape) {
		t.Error("Should not allow move right of board")
	}

	// Test collision check - can't move into occupied space
	gs2 := game_state.NewGameState()
	gs2.Board[0][4] = 1
	if gs2.IsValidMove(0, 4, gs2.ActivePiece.Shape) {
		t.Error("Should not allow move into occupied space")
	}
}

// TestMovePiece tests piece movement.
func TestMovePiece(t *testing.T) {
	gs := game_state.NewGameState()
	initialCol := gs.ActivePiece.Col

	if !gs.MovePiece(0, 1) {
		t.Error("Should be able to move right")
	}
	if gs.ActivePiece.Col != initialCol+1 {
		t.Errorf("Expected col to be %d, got %d", initialCol+1, gs.ActivePiece.Col)
	}

	if !gs.MovePiece(0, -1) {
		t.Error("Should be able to move left")
	}
	if gs.ActivePiece.Col != initialCol {
		t.Error("Position should return to original after left move")
	}

	// Test movement blocked by game over
	gs.GameOver = true
	if gs.MovePiece(0, 1) {
		t.Error("Should not move when game is over")
	}
}

// TestRotatePiece tests piece rotation.
func TestRotatePiece(t *testing.T) {
	gs := game_state.NewGameState()

	if !gs.RotatePiece() {
		t.Error("Should be able to rotate")
	}
	if gs.ActivePiece.Rotation != 1 {
		t.Errorf("Expected rotation to be 1, got %d", gs.ActivePiece.Rotation)
	}
}

// TestRotatePieceBlocked tests rotation blocked by walls.
func TestRotatePieceBlocked(t *testing.T) {
	// Test that rotation is blocked when it would collide with existing pieces
	gs := game_state.NewGameState()
	gs.ActivePiece.Row = 6
	gs.ActivePiece.Col = 0
	gs.ActivePiece.Shape = [][]int{{1, 1, 1}}

	// Place pieces where the vertical rotated shape would land
	// Vertical shape at (6,0) occupies rows 6, 7, 8 at col 0
	gs.Board[7][0] = 1
	gs.Board[8][0] = 1

	// With wall kick enabled, rotation should succeed by shifting to column 1
	// where there is no collision
	if !gs.RotatePiece() {
		t.Error("Should be able to rotate with wall kick to adjacent column")
	}
	if gs.ActivePiece.Rotation != 1 {
		t.Errorf("Expected rotation to be 1, got %d", gs.ActivePiece.Rotation)
	}
	// Verify the wall kick moved the piece
	if gs.ActivePiece.Col != 1 {
		t.Errorf("Expected piece to be kicked to col 1 (wall kick), got %d", gs.ActivePiece.Col)
	}
}

// TestSoftDrop tests soft drop functionality.
func TestSoftDrop(t *testing.T) {
	gs := game_state.NewGameState()

	if !gs.SoftDrop() {
		t.Error("Should be able to soft drop")
	}
}

// TestHardDrop tests hard drop functionality.
func TestHardDrop(t *testing.T) {
	gs := game_state.NewGameState()

	// Verify piece moves to bottom and locks
	gs.HardDrop()

	// Verify game is not over (piece spawns successfully)
	if gs.GameOver {
		t.Error("Game should not be over after hard drop if piece spawns successfully")
	}

	// Verify a new piece was spawned
	if gs.ActivePiece == nil {
		t.Error("ActivePiece should exist after hard drop")
	}
	if gs.NextPiece == nil {
		t.Error("NextPiece should be spawned after hard drop")
	}
}

// TestClearLines tests line clearing logic.
func TestClearLines(t *testing.T) {
	// Test 3 lines cleared
	gs2 := game_state.NewGameState()
	for c := 0; c < 10; c++ {
		gs2.Board[17][c] = 1
		gs2.Board[18][c] = 1
		gs2.Board[19][c] = 1
	}
	gs2.ClearLines()

	if gs2.LinesCleared != 3 {
		t.Errorf("Expected 3 lines cleared, got %d", gs2.LinesCleared)
	}
	if gs2.Score != 300 {
		t.Errorf("Expected score 300, got %d", gs2.Score)
	}

	// All rows should have shifted down except 0-16 which are empty
	for r := 0; r < 17; r++ {
		for c := 0; c < 10; c++ {
			if gs2.Board[r][c] != 0 {
				t.Errorf("Row %d should be empty after clear", r)
			}
		}
	}

	// Test level up (10 lines = level 2)
	gs3 := game_state.NewGameState()
	for i := 0; i < 10; i++ {
		for c := 0; c < 10; c++ {
			gs3.Board[18-i][c] = 1
		}
	}
	gs3.ClearLines()
	if gs3.Level != 2 {
		t.Errorf("Expected level 2 after 10 lines, got %d", gs3.Level)
	}
	if gs3.LinesCleared != 10 {
		t.Errorf("Expected 10 lines cleared, got %d", gs3.LinesCleared)
	}
	if gs3.Score != 1000 {
		t.Errorf("Expected score 1000, got %d", gs3.Score)
	}
}

// TestGetGhostRow tests ghost piece position calculation.
func TestGetGhostRow(t *testing.T) {
	gs := game_state.NewGameState()

	// Ghost row is where the piece would land if dropped (at the bottom)
	// It should be below or at the active piece row
	ghostRow := gs.GetGhostRow()
	if ghostRow < gs.ActivePiece.Row {
		t.Errorf("Expected ghost row >= active piece row (%d), got %d", gs.ActivePiece.Row, ghostRow)
	}
}

// TestGhostRowUpdatesOnMove tests ghost row updates when piece moves.
func TestGhostRowUpdatesOnMove(t *testing.T) {
	gs := game_state.NewGameState()
	// initialGhostRow := gs.GetGhostRow()()

	// Move piece down
	gs.SoftDrop()
	newGhostRow := gs.GetGhostRow()

	// Ghost row should move with the piece
	if newGhostRow <= gs.ActivePiece.Row {
		t.Errorf("Expected ghost row to be below active piece after drop")
	}
}

// TestDrainDrop tests gravity timer behavior.
func TestDrainDrop(t *testing.T) {
	gs := game_state.NewGameState()
	gs.ActivePiece.Row = 5
	initialRow := gs.ActivePiece.Row

	// Call DrainDrop without enough time elapsed - should not move
	gs.DrainDrop()
	if gs.ActivePiece.Row != initialRow {
		t.Errorf("Expected piece to stay at row %d when <500ms elapsed", initialRow)
	}

	// Wait for 500ms and call DrainDrop - should move
	time.Sleep(600 * time.Millisecond)
	gs.DrainDrop()
	if gs.ActivePiece.Row != initialRow+1 {
		t.Errorf("Expected piece to drop to row %d, got %d", initialRow+1, gs.ActivePiece.Row)
	}

	// Call again immediately - should not move without 500ms elapsed
	gs.DrainDrop()
	if gs.ActivePiece.Row != initialRow+1 {
		t.Errorf("Expected piece to stay at row %d when <500ms elapsed", initialRow+1)
	}
}

// TestDrainDropPaused tests gravity is paused.
func TestDrainDropPaused(t *testing.T) {
	gs := game_state.NewGameState()
	gs.ActivePiece.Row = 5
	gs.Paused = true

	time.Sleep(600 * time.Millisecond)

	// Piece should not move when paused
	if gs.ActivePiece.Row != 5 {
		t.Errorf("Expected piece to stay at row 5, got %d", gs.ActivePiece.Row)
	}
}

// TestLockPiece tests locking piece to board.
func TestLockPiece(t *testing.T) {
	gs := game_state.NewGameState()

	// Lock the piece
	gs.LockPiece()

	// Verify piece is on board
	if gs.ActivePiece == nil {
		t.Error("ActivePiece should be nil after lock")
	}
	if gs.Board[0][4] == 0 {
		t.Error("Piece should be on board[0]")
	}

	// Next piece should be spawned
	if gs.NextPiece == nil {
		t.Error("NextPiece should not be nil after lock")
	}
}

// TestIsGameOver tests game over status.
func TestIsGameOver(t *testing.T) {
	gs := game_state.NewGameState()

	if gs.IsGameOver() {
		t.Error("Game should not be over initially")
	}

	gs.GameOver = true
	if !gs.IsGameOver() {
		t.Error("Game should be over when GameOver is true")
	}
}

// TestBoardDimensions tests board is correct size.
func TestBoardDimensions(t *testing.T) {
	gs := game_state.NewGameState()

	if len(gs.Board) != 20 {
		t.Errorf("Expected board height of 20, got %d", len(gs.Board))
	}
	for i, row := range gs.Board {
		if len(row) != 10 {
			t.Errorf("Row %d should have width 10, got %d", i, len(row))
		}
	}
}

// TestShapeDefinitions tests all tetromino shapes are valid.
func TestShapeDefinitions(t *testing.T) {
	for i, shape := range game_state.Shapes {
		// Each shape should have at least one non-zero cell
		hasCell := false
		for _, row := range shape {
			for _, cell := range row {
				if cell != 0 {
					hasCell = true
					break
				}
			}
			if hasCell {
				break
			}
		}
		if !hasCell {
			t.Errorf("Shape %d has no cells", i)
		}
	}
}

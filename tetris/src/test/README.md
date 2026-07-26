# Tetris Test Harness

This test suite provides comprehensive testing for the Tetris game engine.

## Running Tests

```bash
# Run all tests
go test -v ./src/...

# Run tests with coverage
go test -coverprofile=coverage.out ./src/...
go tool cover -html=coverage.out -o coverage.html

# Run a specific test
go test -v -run TestMovePiece ./src/test
```

## Test Coverage

### game_state_test.go

Tests for the `GameState` implementation:

| Test | Coverage |
|------|----------|
| `TestNewGameState` | Initialization, random piece spawning, ghost row calculation |
| `TestSpawnRandomPiece` | Piece spawning at correct position |
| `TestInvalidMove` | Boundary and collision validation |
| `TestMovePiece` | Piece movement (left/right) |
| `TestRotatePiece` | Piece rotation |
| `TestRotatePieceBlocked` | Rotation blocked by walls/pieces |
| `TestSoftDrop` | Soft drop functionality |
| `TestHardDrop` | Hard drop (instant placement) |
| `TestClearLines` | Line clearing, scoring, level progression |
| `TestGetGhostRow` | Ghost piece position calculation |
| `TestGhostRowUpdatesOnMove` | Ghost row updates with piece movement |
| `TestDrainDrop` | Gravity timer behavior |
| `TestDrainDropPaused` | Gravity disabled when paused |
| `TestLockPiece` | Piece locking to board |
| `TestIsGameOver` | Game over detection |
| `TestBoardDimensions` | Board 20x10 dimensions |
| `TestShapeDefinitions` | All 7 tetromino shapes |
| `TestSetup` | Full game setup (clear lines, spawn piece) |
| `TestHandleKeyMoveLeft` | Left arrow key handling |
| `TestHandleKeyMoveRight` | Right arrow key handling |
| `TestHandleKeySoftDrop` | Down arrow key handling |
| `TestHandleKeyHardDrop` | Space key handling |
| `TestHandleKeyRotate` | Up arrow key handling |
| `TestHandleKeyRotate2` | Down arrow rotate (alternate) |
| `TestHandleKeyUnknownKey` | Unknown key handling |
| `TestHandleKeyPausedDisabled` | Pause when already paused |
| `TestHandleKeyGameOverDisabled` | Input blocked when game over |

### renderer_test.go

Tests for the renderer package:

| Test | Coverage |
|------|----------|
| `TestColorMap` | All 7 piece colors (Cyan, Yellow, Green, Magenta, Blue, Red) |
| `TestColorMapEmpty` | Empty slot color (Black) |
| `TestColorMapPieceType` | Color mapping for all piece types |
| `TestANSIColorCodes` | Correct ANSI color codes |
| `TestColorFormat` | Proper format with color start and reset codes |
| `TestColorMapCoverage` | All indices 0-6 mapped |
| `TestColorMapExtraIndices` | Indices beyond 6 return Black |

## Key Features Tested

### GameState Mechanics
- **Piece spawning**: Random piece generation with proper centering
- **Movement**: Left/right movement with boundary checks
- **Rotation**: 90° counter-clockwise rotation with wall collision
- **Gravity**: 500ms drop timer per hard drop
- **Hard drop**: Instant piece placement
- **Line clearing**: Detection and removal of complete rows
- **Scoring**: Points per line (100/300/500/800 for 1/2/3/4 lines)
- **Level progression**: Level increases every 10 lines
- **Collision detection**: Prevents movement into walls or locked pieces
- **Game over**: Detection when piece spawns on locked pieces
- **Pause state**: Disables input and gravity when paused

### Input Handling
- **Arrow keys**: Movement and rotation
- **Space**: Hard drop
- **Enter**: Pause
- **Unknown keys**: Ignored

### Rendering
- **ANSI colors**: Proper escape codes for terminal colors
- **Color reset**: Proper formatting with color reset codes
- **Piece colors**: Distinct colors for each tetromino

## Adding New Tests

1. Create a test file in `test/` directory
2. Import the packages being tested:
   ```go
   import (
       "testing"
       "github.com/aaron/tetris/src/game_state"
   )
   ```
3. Follow Go test naming convention: `func TestXxx(t *testing.T)`
4. Use `t.Errorf()` or `t.Error()` for failures
5. Run tests with `go test -v ./src/...`

## Test Dependencies

All tests are self-contained and do not require:
- User input (uses mock input handling)
- External resources
- Specific timing (except where noted for gravity tests)

Tests can be run in any order and are independent of each other.

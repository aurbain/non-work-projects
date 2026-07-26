# Tetris Test Suite

This project includes a comprehensive test suite for the Tetris game implementation.

## Running Tests

### Quick Test Run

```bash
cd /Users/aaron/Desktop/software-projects/on-github/non-work-projects/tetris
go test -v ./src/...
```

### Run Specific Package Tests

```bash
# Game state tests
go test -v ./src/game_state/

# Input tests
go test -v ./src/input/

# Renderer tests
go test -v ./src/renderer/

# Integration tests
go test -v -tags=integration ./src/...
```

### Run with Coverage Report

```bash
go test -v -coverprofile=coverage.out ./src/...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## Test Coverage

### `src/test/game_state_test.go`

Tests for the game state logic:
- `TestNewGameState` - Board initialization
- `TestSpawnRandomPiece` - Piece spawning
- `TestInvalidMove` - Move validation
- `TestMovePiece` - Horizontal movement
- `TestRotatePiece` - Piece rotation
- `TestRotatePieceBlocked` - Blocked rotation
- `TestSoftDrop` - Soft drop
- `TestHardDrop` - Hard drop
- `TestClearLines` - Line clearing and scoring
- `TestGetGhostRow` - Ghost piece calculation
- `TestDrainDrop` - Automatic gravity
- `TestDrainDropPaused` - Gravity pause behavior
- `TestLockPiece` - Piece locking to board
- `TestIsGameOver` - Game over state
- `TestBoardDimensions` - Board size validation
- `TestShapeDefinitions` - Tetromino shapes

### `src/test/input_test.go`

Tests for input handling:
- `TestKeyStateCreation` - KeyState struct
- `TestSetup` - Input setup goroutine
- `TestHandleKeyMoveLeft` - Left movement ('a')
- `TestHandleKeyMoveRight` - Right movement ('d')
- `TestHandleKeySoftDrop` - Soft drop ('s', ' ')
- `TestHandleKeyHardDrop` - Hard drop ('h')
- `TestHandleKeyRotate` - Rotation ('q', 'w')
- `TestHandleKeyUnknownKey` - Unknown key handling
- `TestHandleKeyPausedDisabled` - Input disabled when paused
- `TestHandleKeyGameOverDisabled` - Input disabled when game over

### `src/test/renderer_test.go`

Tests for renderer configuration:
- `TestColorMap` - All color mappings
- `TestColorMapEmpty` - Empty slot color
- `TestColorMapPieceType` - All piece type colors
- `TestANSIColorCodes` - ANSI code format
- `TestColorFormat` - Escape character format
- `TestColorMapCoverage` - Game state piece coverage
- `TestColorMapExtraIndices` - Extra index handling

### `src/test/main_test.go`

Integration tests:
- `TestCompleteGameFlow` - Complete game scenario
- `TestBoundaryConditions` - Edge cases at board boundaries
- `TestPauseState` - Game state during pause
- `TestGameOverState` - Game over behavior
- `TestHardDrop` - Hard drop behavior
- `TestScoreCalculation` - Scoring logic
- `TestGhostPiecePosition` - Ghost piece accuracy

## Test Architecture

The test suite is organized by package:
- **Unit Tests**: Individual function behavior
- **Integration Tests**: Combined behavior across packages
- **Edge Case Tests**: Boundary conditions and error handling

All tests are tagged with `integration` for selective running:
- Run all tests by default
- Skip integration tests with `-run=^$ -tags=integration`

## Build Requirements

- Go 1.21+
- No external dependencies beyond standard library

## Adding New Tests

1. Add test to appropriate package's test file
2. Follow existing test naming conventions
3. Include unit and integration tests as applicable
4. Update this README if adding new test categories

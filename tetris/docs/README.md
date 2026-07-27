# Tetris Game

A Tetris implementation written in Go with a comprehensive test suite.

## Project Structure

```
tetris/
├── main.go              # Game entry point
├── go.mod               # Go module definition
├── docs/                # Project documentation
├── src/
│   ├── main/            # Main package utilities
│   ├── game_state/      # Core game logic
│   ├── input/           # Input handling
│   ├── renderer/        # Console rendering
│   └── test/            # Test harness
└── test/                # External test directory
```

## Game Controls

- `A`/`D` - Move left/right
- `Q`/`W` - Rotate piece
- `S`/`H` - Soft/Hard drop
- `Space` - Soft drop
- `P` - Pause/Resume
- `R` - Restart (after game over)

## Core Components

### Game State (`src/game_state/game_state.go`)

- **20x10 Board**: Classic Tetris dimensions
- **7 Tetromino Shapes**: I, O, T, S, Z, J, L
- **Key Features**:
  - Ghost piece preview
  - Next piece preview
  - Level progression (increases every 10 lines)
  - Scoring: `Lines * 100 * Level`

### Input (`src/input/input.go`)

- Maps keyboard input to game actions
- Non-blocking input handling
- Supports both soft drop keys (`S`, `Space`)

### Renderer (`src/renderer/renderer.go`)

- ANSI color-coded board rendering
- Color mapping:
  - O (Square): Cyan
  - T: Yellow
  - S: Green
  - Z: Magenta
  - J: Blue
  - L: Red
- Displays ghost piece position

## Running the Game

```bash
go run main.go
```

## Running Tests

### Quick Test Run
```bash
go test -v ./src/...
```

### Run Specific Packages
```bash
# Game state tests
go test -v ./src/game_state/

# Input tests
go test -v ./src/input/

# Renderer tests
go test -v ./src/renderer/
```

### Run with Coverage
```bash
go test -v -coverprofile=coverage.out ./src/...
go tool cover -html=coverage.out -o coverage.html
```

### Use gotests
```bash
# Generate test functions
gotests

# Run tests
go test -v ./src/...
```

## Test Architecture

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test interactions between packages
- **Edge Case Tests**: Boundary conditions and error handling

## Development Guidelines

1. Follow Go naming conventions
2. Write tests before adding new features
3. Keep functions focused (single responsibility)
4. Document public functions with godoc comments

# Tetris Game Plan

## Phase 1: Project Structure and Foundation
1. **Fix Module Configuration**: Update `go.mod` to ensure the module path matches the imports used in the code (e.g., `github.com/aaron/tetris`).
2. **Reorganize Project Structure**:
    * Create the `tetris/src/game_state` directory.
    * Cleanup `tetris/src/main.go` (it seems to contain redundant or misplaced code).
3. **Define Game State**: Implement the core game logic in `tetris/src/game_state/game_state.go`, including:
    * Board representation (2D grid).
    * Piece definitions (Tetrominoes) and rotations.
    * Movement logic (left, right, down, hard drop).
    * Rotation logic (checking for collisions).
    * Line clearing and scoring.
    * Game over conditions.

## Phase 2: Input and Rendering
4. **Implement Input Handling**: Refine `tetris/src/input/input.go` to correctly capture key presses from the terminal and communicate them to the game state.
5. **Implement Rendering**: Complete `tetris/src/renderer/renderer.go` to:
    * Clear the terminal screen.
    * Draw the board and the current active piece.
    * Display scores and levels.
6. **Integrate Main Loop**: Refine `tetris/src/main/main.go` to orchestrate the game loop:
    * Initialize game state, input, and renderer.
    * Run a loop that handles timing (gravity), processes input, updates the game state, and renders the frame.

## Phase 3: Polish and Testing
7. **Refine Gameplay Mechanics**: Adjust gravity speed based on the level, add "soft drop" mechanics, and ensure smooth piece movement.
8. **Final Cleanup**: Remove any remaining temporary or broken files.
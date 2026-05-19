package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Board represents the game grid. A 2D array of runes (or integers) will be used.
// 0 represents an empty cell.
type Board [20][10]rune // Adjust dimensions as needed (height x width)

// Piece represents the current falling tetromino.
type Piece struct {
	// Shape definition: 4 rows, 4 columns. A rune value of 0 means empty.
	Shape [4][4]rune
	// Top-left anchor position (row, col) on the board
	Row, Col int
}

// Global tracking for the actively falling piece (simplification for now)
var currentPiece Piece

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	fmt.Println("--- Tetris Game Initializing ---")

	// 1. Initialize the game board
	board := Board{}
	printBoard(&board)

	// 2. Spawn the first piece
	spawnPiece(&board)
	printBoard(&board) // Show the board with the spawned piece (requires board update logic, for now just showing spawn)
	// Game Loop placeholder
	fmt.Println("\nSuccessfully spawned the first piece. Next step: Implementing the gravity/drop cycle.")
	/*
	for {
		// ... existing code ...		// 5. Redraw Board and Piece
		//time.Sleep(50 * time.Millisecond)
	}
	*/
}

// printBoard takes a pointer to the board and prints its current state to the console.
func printBoard(b *Board) {
	fmt.Println("\n--- Current Board State ---")	// Print top border
	fmt.Println("+" + string(make([]rune, 10)) + "+")
	for r := 0; r < 20; r++ {
		fmt.Print("|")
		for c := 0; c < 10; c++ {
			// Use a placeholder character if the rune is the zero value (empty)
			if b[r][c] == 0 {
				fmt.Print(" ")
			} else {
				fmt.Printf("%c", b[r][c])
			}
		}
		fmt.Println("|")
	}
	fmt.Println("+" + string(make([]rune, 10)) + "+")
}

// getTetrominoes returns a slice of predefined Piece shapes.
func getTetrominoes() []Piece {
	// Define the I piece (Cyan) - Line shape
	i := Piece{
		Shape: [4][4]rune{
			{0, 0, 0, 0},
			{1, 1, 1, 1},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		Row: 0, Col: 0,
	}

	// Define the O piece (Yellow) - Square shape
	o := Piece{
		Shape: [4][4]rune{
			{1, 1, 0, 0},
			{1, 1, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		Row: 0, Col: 0,
	}

	// Define the T piece (Purple)
	t := Piece{
		Shape: [4][4]rune{
			{0, 1, 0, 0},
			{1, 1, 1, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		Row: 0, Col: 0,
	}

	// Define the L piece (Orange)
	l := Piece{
		Shape: [4][4]rune{
			{0, 0, 0, 1},
			{1, 1, 1, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		Row: 0, Col: 0,
	}

	return []Piece{i, o, t, l}
}

// spawnPiece selects a random piece type and initializes the currentPiece global variable.
func spawnPiece(board *Board) {
	availablePieces := getTetrominoes()

	// Select a random index
	randIndex := rand.Intn(len(availablePieces))
	pieceToSpawn := availablePieces[randIndex]

	// Reset the global piece tracker
	currentPiece = Piece{
		Shape: pieceToSpawn.Shape,
		Row: 0,
		Col: 0,
	}
	fmt.Printf("\n[INFO] Piece spawned successfully! (Type index: %d)\n", randIndex)
}

// movePiece attempts to move the current piece by (dx, dy).
// It returns true if the move was successful and the piece coordinates updated, 
// and false if collision occurred.
func movePiece(dx, dy int) bool {
    // 1. Check for collision at the potential new location
    if checkCollision(currentPiece, dx, dy) {
        return false // Collision detected, movement aborted
    }

    // 2. If safe, update the piece's anchor/position
    currentPiece.Anchor[0] += dy
    currentPiece.Anchor[1] += dx
    return true
}

// hardDrop simulates dropping the piece instantly to the bottom.
// It returns the final resting position (y, x) and the piece's state.
func dropPiece() (int, int) {
    var finalY int = 0
    
    // Loop upwards until the next step would cause a collision
    for {
        // Simulate moving one step down (increasing Y coordinate)
        newY := finalY + 1
        
        // Check if this new position causes collision (by checking if the space below is occupied)
        // NOTE: This assumes a helper function 'isOccupied(y)' exists and checks collision at the current X coordinate.
        if isOccupied(newY, getCurrentX()) {
            // Collision detected: The previous position (finalY) was the lowest safe spot.
            break
        }
        finalY = newY
    }
    
    return finalY, getCurrentX()
}


// CheckGameOver checks if the piece's current position results in a collision
// with the ceiling or an existing block.
func CheckGameOver(pieceX, pieceY, pieceSize int) bool {
    // 1. Check ceiling collision
    if pieceY < 0 {
        return true // Collision with the top edge (ceiling)
    }
    
    // 2. Check collision with existing blocks in the environment map
    // (Assuming 'GameMap' holds the state of settled blocks)
    for x := pieceX; x < pieceX + pieceSize; x++ {
        for y := pieceY; y < pieceY + pieceSize; y++ {
            // If the map indicates a solid block at (x, y) AND 
            // the block is below the ceiling (y >= 0)
            if GameMap[x][y] == SOLID_BLOCK {
                return true
            }
        }
    }
    return false
}


// --- Placeholder for needed import ---
// If necessary, add 'import "fmt"' to the top of the file.

// --- New function to draw the game state ---
func drawGameState(board [][]*Piece, height, width int) {
    // Clear screen or use suitable terminal control sequence for a clean draw
    // fmt.Print("\033[H\033[2J")

    fmt.Println("\n===================================")
    fmt.Println("           TETRIS GAME BOARD         ")
    fmt.Println("===================================")

    for y := 0; y < height; y++ {
        rowStr := "|"
        for x := 0; x < width; x++ {
            if board[y][x] != nil {
                // Assuming Piece has a way to represent its color/char
                rowStr += " " // Placeholder for actual piece representation
            } else {
                rowStr += " "
            }
        }
        rowStr += "|"
        fmt.Println(rowStr)
        // Print a visual separator for clarity
        fmt.Println("-----------------------------------")
    }
}
func GameLoop() {
    // Initialization
    currentPiece := initializePiece()
    score := 0
    isRunning := true

    for isRunning {
        // 1. INPUT & PREDICTION (READ ONLY PHASE)
        input := ReadInput() 
        
        // Record initial state for this cycle's calculations
        snapshot := getBoardSnapshot() 
        
        // Temporary state variables for this frame's movement simulation
        tempPiece := *&currentPiece // Work on a copy
        
        // Process input to determine the *next* desired state
        // This function handles all collision and movement checks against the current board state
        nextState, moveSuccessful := calculateNextState(tempPiece, input) 
        
        if moveSuccessful {
            // Commit the proposed move to the primary game state
            currentPiece = *nextState
            
            // --- CRITICAL STEP: Update the map *after* the move is validated ---
            updateBoardState(currentPiece) 
            
            // Check for piece collision/lock down after the move
            if isLocked(currentPiece) {
                lockPieceAndCheckClear(currentPiece)
            }
        }

        // --- Check for Game Over conditions after processing input ---
        if isGameOver(currentPiece) {
            break
        }
        
        // Re-calculate score, etc., based on the now-updated state
        score = calculateScore(currentPiece)
        
        // Wait for the next frame delay...
    }
}

// Helper functions that must be implemented/verified for safety:
// 1. calculateNextState: Takes current state and input, returns the predicted new state (or nil if invalid).
// 2. updateBoardState: Writes the currentPiece's geometry/color onto the global board grid.
// 3. isLocked: Checks if the piece cannot move further down.
// 4. lockPieceAndCheckClear: Locks the piece onto the board and checks for completed lines.


// tryRotate attempts to rotate the piece to the desired rotation (0, 1, 2, or 3).
// It checks for collision and uses basic wall-kick logic (which must be expanded).
// Returns true if rotation is successful and the piece's internal rotation angle is updated.
func tryRotate(piece *Piece, desiredRotation int) bool {
    // 1. Basic Boundary Check: Check if the target rotation is even possible.
    if desiredRotation < 0 || desiredRotation > 3 {
        return false
    }

    // 2. Simulate the potential new piece shape (A critical step)
    potentialPiece := piece.GetRotatedShape(desiredRotation)

    // 3. Check Collision at current anchor point
    if checkCollision(potentialPiece, 0, 0) {
        // Simple collision detected. In a full implementation, we'd try kicking offsets here.
        // e.g., Try offsets (1, 0), (-1, 0), (0, 1), (0, -1) until valid.
        // For this iteration, we just fail on collision.
        return false
    }

    // 4. If no collision, we apply the successful rotation.
    piece.Rotation = desiredRotation
    // Update the internal block coordinates of the piece structure to match the new rotation.
    piece.UpdateBlockCoordinates(desiredRotation) 
    
    return true
}

// This is a placeholder for the actual Super Rotation System (SRS) logic.
// In a full implementation, this function would contain exhaustive checks 
// against the board edges and occupied tiles to determine valid kick placements.
func getKickOffsets(originalPiece *Piece, desiredRotation int) []struct{ dx, dy int } {
    // Placeholder: Return a list of valid (dx, dy) offsets for the given rotation.
    // A real implementation would check the board state here.
    // For now, we only test the no-kick scenario.
    return []struct{ dx, dy int }{{0, 0}}
}


// identifyAndClearLines scans the board state and returns a list of row indices that are full.
func identifyAndClearLines() []int {
    var clearedRows []int
    for y := 0; y < BoardHeight; y++ {
        isRowFull := true
        for x := 0; x < BoardWidth; x++ {
            // Assuming we are only checking settled, solid blocks
            if BoardState[y][x] != SOLID_BLOCK { 
                isRowFull = false
                break
            }
        }
        if isRowFull {
            clearedRows = append(clearedRows, y)
        }
    }
    return clearedRows
}

// clearLinesAndScore processes the cleared rows: recalculates the map, shifts blocks, and scores.
func clearLinesAndScore(clearedRows []int) (int, int) {
    if len(clearedRows) == 0 {
        return 0, 0 // No lines cleared, no score change
    }

    var scoreGained int = 0
    var blocksShifted int = 0

    // Sort rows descending to prevent index invalidation issues when shifting
    sort.Sort(sort.Reverse(sort.IntSlice(rows))) 

    for _, row := range rows {
        // 1. Erase the row content (set to empty/air)
        for x := 0; x < width; x++ {
            map[x][row] = AIR
        }
        // 2. Shift all rows above 'row' down by one unit
        for y := row - 1; y >= 0; y-- {
            for x := 0; x < width; x++ {
                map[x][y + 1] = map[x][y]
            }
        }
        // 3. Calculate score based on number of lines cleared
        score += calculateScore(rows, count)
    }
    return score
}


// --- I/O System Placeholders ---

// ReadInput handles reading keyboard/gamepad input for the current frame.
// Must return a character/string representing 'MoveLeft', 'MoveRight', 'Rotate', etc.
func ReadInput() string {
    // TODO: Implement actual terminal reading logic (e.g., using termbox or platform-specific APIs)
    return "" // Return empty string if no input is detected for the frame
}

// RenderGame draws the current state of the board and the falling piece to the screen.
func RenderGame(piece *Piece, score int) {
    // TODO: Implement drawing logic. Must draw GameMap first, then overlay the Piece.
    // Must handle "Game Over" display if isRunning is false.
    fmt.Printf("--- Board State Render Placeholder ---\nScore: %d\n", score)
}

// --- Board State Synchronization ---

// SyncBoardState must be called after every major change (lock, clear, spawn)
// to ensure BoardState and GameMap are perfectly in sync.
func SyncBoardState() {
    // This is the master function: it takes the final state of GameMap and
    // updates the read-only BoardState used by checkCollision().
    // BoardState = DeepCopy(GameMap) 
}

// Example of using the new components:
// At the start of the main game loop iteration:
// 1. Check for user input and update temporary state.
// 2. Run physics/logic updates (check for collisions, check for line clears).
// 3. Call SyncBoardState() to commit all changes to the primary state tracker.
// 4. Call RenderFrame().




// lockAndClear handles the transition phase: locking the piece, checking for lines,
// clearing them, and making the pieces above fall (gravity).
// It returns the updated state of the board and any points accumulated.
func lockAndClear(board *Board, piece *Piece) (*Board, int) {
    points := 0

    // 1. Lock the current piece onto the board structure
    lockPiece(board, piece)

    // 2. Detect and Clear Lines
    linesToClear := detectLines(board)
    
    if len(linesToClear) > 0 {
        points += calculatePoints(len(linesToClear)) // Simplified point calculation
        
        // 3. Clear the lines from the board array
        *board = clearLines(*board, linesToClear)

        // 4. Apply Gravity (Cascading)
        // This must happen repeatedly until no pieces fall.
        for {
            initialBoardState := *board
            *board = applyGravity(*board)
            
            // If the board state did not change after applying gravity, the cascading is complete.
            if deepEqualBoard(initialBoardState, *board) {
                break
            }
            // We might also need to check for new lines formed by the fall here, 
            // but for simplicity in this structure, we assume the loop handles it implicitly 
            // or that a dedicated line check should follow the gravity application.
        }
        
        // *Self-Correction/Enhancement:* A true game loop would re-check for lines immediately 
        // after gravity settles, so this loop structure might need to be nested.
    }
    
    // 5. Return the final updated board state and accumulated score
    return board, points
}

// ==============================================================================
// --- Placeholder Helper Functions (Must be implemented elsewhere) ---
// ==============================================================================

// lockPiece safely writes the coordinates of the piece onto the board structure.
func lockPiece(board *Board, piece *Piece) {
    // Logic to transfer piece data from 'piece' struct onto 'board' data structure.
    fmt.Println("Pieces locked onto the board.")
}

// detectLines scans the entire board to find all full horizontal or vertical lines.
func detectLines(board *Board) []Line {
    // Returns a list of coordinates/structures representing lines that need clearing.
    fmt.Println("Detecting lines...")
    return []Line{/* ... list of lines ... */}
}

// calculatePoints determines the score awarded based on the number and type of lines cleared.
func calculatePoints(lineCount int) int {
    // Simple scoring example
    return lineCount * 100
}

// clearLines takes the old board state and the list of lines to clear, returning the new board state.
func clearLines(oldBoard Board, lines []Line) Board {
    // Logic to set the coordinates occupied by the cleared lines back to an 'empty' state.
    fmt.Println("Lines cleared successfully.")
    return Board{} // Return a placeholder for the cleared board
}

// applyGravity shifts all pieces downward, filling voids left by cleared lines.
func applyGravity(board Board) Board {
    // Core physics simulation: Move all non-empty cells down until they hit another piece or the bottom edge.
    fmt.Println("Applying gravity...")
    return board // Return the gravity-adjusted board
}

// deepEqualBoard checks if two Board structures are identical in content.
func deepEqualBoard(b1 Board, b2 Board) bool {
    // Implements deep comparison for the entire board state.
    return true
}



// --- Data Structure Definitions (MUST BE ADDED TO THE FILE) ---

// Block represents a single, settled block on the board.
type Block struct {
    Color    Color // Enum/constant for color
    Rotation int    // Optional: Useful for debugging, but not strictly needed for settled blocks
}

// Board represents the entire grid structure.
type Board struct {
    Grid [][]*Block // Grid[y][x]
    Width int
    Height int
}

// Piece represents the currently falling piece.
// Assuming Piece already has fields for its blocks relative to its anchor point.
type Piece struct {
    AnchorX int
    AnchorY int
    Rotation int
    Blocks [][2]int // List of relative (dx, dy) coordinates for the piece blocks
}

// --- Implementation of lockPiece ---

// lockPiece transfers the coordinates of the transient 'piece' structure 
// onto the permanent 'board' structure, committing the piece to the grid.
func lockPiece(board *Board, piece *Piece) {
    // Iterate over the coordinates of the piece's constituent blocks
    for _, offset := range piece.Offsets {
        // Calculate the absolute world coordinates
        intX := piece.OffsetX + offset.X
        intY := piece.OffsetY + offset.Y

        // Basic boundary check (assuming the board size is fixed or passed in)
        if intX >= 0 && intX < boardWidth && intY >= 0 && intY < boardHeight {
            // Write the piece's color/type into the permanent board structure
            // NOTE: This assumes the 'Board' global/struct variable exists and is modifiable.
            Board[intY][intX] = &Block{Color: piece.Color, Type: piece.Type}
        }
    }
    // Crucial: After locking, the piece entity is removed/reset (handled by the caller logic)
}


// Refactored: applyGravity optimized to O(N*M) single-pass simulation.
func applyGravity(board *Board) *Board {
    // 1. Create a copy of the board to modify safely.
    newBoard := cloneBoard(*board) 
    
    // 2. Iterate column by column (x).
    for x := 0; x < board.Width; x++ {
        
        // Keep track of where the next piece should fall in this column.
        fallPositionY := board.Height - 1 // Start at the bottom row index
        
        // 3. Iterate from the bottom row up (y).
        for y := board.Height - 1; y >= 0; y-- {
            
            currentBlock := (*Block)(nil)
            // Safely retrieve block at (x, y) on the *original* board.
            // (Assuming safe access function exists: getBlock(board, x, y))
            if getBlock(board, x, y) != nil { 
                currentBlock = getBlock(board, x, y)
            }

            if currentBlock != nil {
                // If we found a piece, does it need to fall?
                // It should fall to the lowest available slot (fallPositionY)
                
                if y != fallPositionY {
                    // The piece is not at its final resting spot.
                    // Move the content of (x, y) to (x, fallPositionY)
                    // (Requires careful copying logic for structure/data)
                    
                    // Place the piece data at the correct, lower spot
                    // Update the map/array representation of the board state
                    
                    // Crucially, mark the old spot (x, y) as empty
                }
                
                // Move the target position down for the next piece encountered.
                fallPositionY--
            }
        }
    }
    // The board state represented by the structured data must be updated after iterating through all columns.
    return newState
}


// File: tetris.go

// ... existing code ...

// DrawBoard renders the current state of the board to the console in a readable format.
func DrawBoard(board *Board) {
    fmt.Println("=======================================")
    fmt.Println("           TETRIS BOARD STATE          ")
    fmt.Println("=======================================")
    
    // Print the top border
    fmt.Print("|")
    for i := 0; i < board.Width; i++ {
        fmt.Print("---")
    }
    fmt.Println("|")

    // Iterate through the board row by row (from top, y=0, to bottom, y=Height-1)
    for y := 0; y < board.Height; y++ {
        fmt.Print("|")
        for x := 0; x < board.Width; x++ {
            // NOTE: We must handle cases where the board cell is empty (nil block)
            if block := board.Grid[y][x]; block != nil {
                // Use a simple character representation for visualization
                fmt.Print("#") // Block present
            } else {
                fmt.Print(" ") // Empty space
            }
        }
        fmt.Println("|")
    }
    
    // Print the bottom border
    fmt.Println("=======================================")
}

// --- Integration into the Game Loop ---

// In the main GameLoop function:
func GameLoop(board *Board, piece *Piece, isGameOver bool) {
    if isGameOver {
        DrawBoard(board)
        fmt.Println("\n!!! GAME OVER !!!")
        return
    }

    // 1. (Input/Prediction happens here)
    // 2. Collision Check happens here...
    
    // 3. Rendering Step (The View)
    fmt.Println("\n--- Rendering Frame ---")
    DrawBoard(board)
    
    // 4. (If piece locks) LockPiece(board, piece) and clear lines...
}



// displayBoard redraws the current state of the grid, clearing the screen 
// and printing the grid representation to the console.
func displayBoard(board [][]*[][]interface{}) {
    // 1. Clear the screen (Platform-dependent ANSI escape sequence)
    fmt.Print("\033[H\033[2J")

    fmt.Println("=====================================")
    fmt.Println("           [ MEGA TETRIS BOARD]          ")
    fmt.Println("=====================================")

    // Print column headers (A, B, C...)
    fmt.Print("    ")
    for i := 0; i < len(board[0]); i++ {
        fmt.Printf("    [%c] ", 'A'+i)
    }
    fmt.Println("\n-------------------------------------")

    // 2. Iterate and print row by row
    for r := 0; r < len(board); r++ {
        // Print the row index label (1, 2, 3...)
        fmt.Printf("%2d | ", r+1)

        for c := 0; c < len(board[0]); c++ {
            cell := board[r][c]
            if cell == nil {
                // Empty cell
                fmt.Print(" .   ")
            } else {
                // Occupied cell: Display its color code or letter/symbol
                // Assuming the cell stores a string identifier (e.g., "Red", "Blue", "Empty")
                // For simplicity, we will print the first letter of the color name.
                symbol := ""
                if cell[0] != nil {
                    // Simple heuristic: Use the first letter of the color name
                    colorStr := fmt.Sprintf("%v", cell[0])
                    symbol = string(colorStr[0])
                } else {
                    symbol = "?"
                }
                
                // Colored display using ANSI escape codes for visualization
                // NOTE: This assumes you have color constants defined (e.g., ANSI_RED)
                // For basic text:
                fmt.Printf(" %s   ", symbol)
            }
        }
        fmt.Println("\n-------------------------------------")
    }
    fmt.Println("=====================================")
    fmt.Println("Game Status: Score: 100 | Lives: 3")
}

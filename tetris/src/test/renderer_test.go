package test

import (
	"testing"

	"github.com/aaron/tetris/src/renderer"
)

// TestColorMap tests all color mappings.
func TestColorMap(t *testing.T) {
	colors := []string{}

	// Check all piece colors exist
	for piece := 0; piece <= 6; piece++ {
		if color, ok := renderer.ColorMap[piece]; ok {
			colors = append(colors, color)
		} else {
			t.Errorf("Color mapping missing for piece type %d", piece)
		}
	}

	// Verify all pieces have non-empty color strings
	if len(colors) != 7 {
		t.Errorf("Expected 7 piece colors, got %d", len(colors))
	}
}

// TestColorMapEmpty tests empty slot color is Black (index 0).
func TestColorMapEmpty(t *testing.T) {
	if color, ok := renderer.ColorMap[0]; !ok {
		t.Error("Color mapping missing for piece type 0 (empty)")
	} else if color != renderer.Black {
		t.Errorf("Expected empty color to be Black, got %q", color)
	}
}

// TestColorMapPieceType tests all piece type colors.
func TestColorMapPieceType(t *testing.T) {
	expectedColors := []string{
		renderer.Black,    // 0: Empty
		renderer.Cyan,     // 1: Cyan (O/I)
		renderer.Yellow,   // 2: Yellow (T)
		renderer.Green,    // 3: Green (S)
		renderer.Magenta,  // 4: Magenta (Z)
		renderer.Blue,     // 5: Blue (J)
		renderer.Red,      // 6: Red (L)
	}

	for i, expected := range expectedColors {
		if actual, ok := renderer.ColorMap[i]; !ok {
			t.Errorf("Color mapping missing for piece type %d", i)
		} else if actual != expected {
			t.Errorf("Piece type %d: expected color %q, got %q", i, expected, actual)
		}
	}
}

// TestANSIColorCodes tests ANSI code format.
func TestANSIColorCodes(t *testing.T) {
	testCodes := map[string]string{
		"Black":   renderer.Black,
		"Red":     renderer.Red,
		"Green":   renderer.Green,
		"Yellow":  renderer.Yellow,
		"Blue":    renderer.Blue,
		"Magenta": renderer.Magenta,
		"Cyan":    renderer.Cyan,
		"White":   renderer.White,
	}

	for name, code := range testCodes {
		if len(code) == 0 {
			t.Errorf("%s color code is empty", name)
		}
		if len(code) < 3 {
			t.Errorf("%s color code appears incomplete: %q", name, code)
		}
	}
}

// TestColorFormat tests color codes start with ESC.
func TestColorFormat(t *testing.T) {
	testColors := map[string]string{
		"Black":   renderer.Black,
		"Red":     renderer.Red,
		"Green":   renderer.Green,
		"Yellow":  renderer.Yellow,
		"Blue":    renderer.Blue,
		"Magenta": renderer.Magenta,
		"Cyan":    renderer.Cyan,
	}

	for name, code := range testColors {
		// Check if code starts with escape character
		if len(code) < 1 || code[0] != '\033' {
			t.Errorf("%s should start with escape character", name)
		}
	}
}

// TestColorMapCoverage tests all game_state piece types are mapped.
func TestColorMapCoverage(t *testing.T) {
	// Game state uses 7 shapes (indices 0-6)
	// Color map should have entries for all of them
	for i := 0; i < 7; i++ {
		if _, ok := renderer.ColorMap[i]; !ok {
			t.Errorf("Color mapping missing for piece shape index %d", i)
		}
	}
}

// TestColorMapExtraIndices tests extra indices are handled.
func TestColorMapExtraIndices(t *testing.T) {
	// Test that index 7+ returns empty or doesn't panic
	_, ok1 := renderer.ColorMap[7]
	if ok1 {
		t.Log("Index 7 has color mapping (may be intentional)")
	}
}

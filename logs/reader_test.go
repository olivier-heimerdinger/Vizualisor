package logs

import (
	"testing"
)

func TestFilterLines(t *testing.T) {
	content := `2024-01-15 10:00:00 INFO Starting application
2024-01-15 10:00:01 ERROR Failed to connect to database
2024-01-15 10:00:02 INFO Connected to Redis
2024-01-15 10:00:03 WARN High memory usage detected
2024-01-15 10:00:04 ERROR Timeout on request /api/users
2024-01-15 10:00:05 INFO Request completed successfully`

	tests := []struct {
		name     string
		filter   string
		expected int // nombre de lignes attendues
	}{
		{"no filter", "", 6},
		{"filter ERROR", "ERROR", 2},
		{"filter INFO", "info", 3}, // case insensitive
		{"filter database", "database", 1},
		{"filter inexistant", "XXXXX", 0},
		{"filter empty", "", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterLines(content, tt.filter)
			if tt.filter == "" {
				if result != content {
					t.Errorf("Avec filtre vide, le contenu devrait être identique")
				}
				return
			}

			if tt.expected == 0 {
				if result != "" {
					t.Errorf("Filtre %q devrait retourner vide, got %q", tt.filter, result)
				}
				return
			}

			// Compter les lignes
			lines := 0
			for _, c := range result {
				if c == '\n' {
					lines++
				}
			}
			lines++ // dernière ligne sans \n

			if lines != tt.expected {
				t.Errorf("FilterLines avec %q: got %d lignes, want %d", tt.filter, lines, tt.expected)
			}
		})
	}
}

func TestFilterLinesEmpty(t *testing.T) {
	result := FilterLines("", "test")
	if result != "" {
		t.Errorf("FilterLines sur contenu vide devrait être vide, got %q", result)
	}
}

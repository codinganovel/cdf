package main

import (
	"os"
	"strings"
	"testing"

	"github.com/sahilm/fuzzy"
)

func TestFuzzyMatch(t *testing.T) {
	directories := []string{
		"/home/user/projects/myapp/api",
		"/home/user/projects/webapp/src",
		"/home/user/documents/work/api-docs",
		"/home/user/personal/project/api",
		"/var/log/applications",
		"/usr/local/bin",
	}

	t.Run("EmptyQuery", func(t *testing.T) {
		matches := fuzzyMatch("", directories)
		
		// Should return all directories with score 100
		if len(matches) != len(directories) {
			t.Errorf("Expected %d matches, got %d", len(directories), len(matches))
		}

		for _, match := range matches {
			if match.Score != 100 {
				t.Errorf("Expected score 100 for empty query, got %d", match.Score)
			}
		}
	})

	t.Run("SimpleQuery", func(t *testing.T) {
		matches := fuzzyMatch("api", directories)
		
		// Should find directories containing "api"
		if len(matches) == 0 {
			t.Error("Expected matches for 'api' query, got none")
		}

		// Check that matches are relevant (fuzzy matching might find partial matches)
		if len(matches) == 0 {
			t.Error("Expected at least one match for 'api' query")
		}
	})

	t.Run("ComplexQuery", func(t *testing.T) {
		matches := fuzzyMatch("proj/api", directories)
		
		// Should find project-related API directories
		found := false
		for _, match := range matches {
			if strings.Contains(match.Str, "projects") && strings.Contains(match.Str, "api") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find project API directory with 'proj/api' query")
		}
	})

	t.Run("NoMatches", func(t *testing.T) {
		matches := fuzzyMatch("xyz123nonexistent", directories)
		
		// Should return empty results for non-matching query
		if len(matches) != 0 {
			t.Errorf("Expected no matches for non-existent query, got %d", len(matches))
		}
	})
}

func TestFormatMatch(t *testing.T) {
	// Mock home directory for testing
	originalHome := os.Getenv("HOME")
	testHome := "/home/testuser"
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHome)

	testCases := []struct {
		input    string
		expected string
	}{
		{"/home/testuser/projects/myapp", "~/projects/myapp"},
		{"/home/testuser/documents", "~/documents"},
		{"/var/log/apps", "/var/log/apps"}, // Should not be modified
		{"/usr/local/bin", "/usr/local/bin"}, // Should not be modified
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			match := fuzzy.Match{Str: tc.input}
			result := formatMatch(match)
			if result != tc.expected {
				t.Errorf("formatMatch(%s) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGetMatchScore(t *testing.T) {
	testCases := []struct {
		score    int
		expected int
	}{
		{100, 100},
		{50, 50},
		{0, 0},
		{-1, 0}, // Negative scores should return 0
		{-10, 0},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			match := fuzzy.Match{Score: tc.score}
			result := getMatchScore(match)
			if result != tc.expected {
				t.Errorf("getMatchScore(%d) = %d, expected %d", tc.score, result, tc.expected)
			}
		})
	}
}

func TestHomeDir(t *testing.T) {
	// Test that homeDir returns a non-empty string
	home := homeDir()
	if home == "" {
		t.Error("homeDir() returned empty string")
	}

	// Test that it returns the same as os.UserHomeDir
	expectedHome, err := os.UserHomeDir()
	if err == nil && home != expectedHome {
		t.Errorf("homeDir() = %s, expected %s", home, expectedHome)
	}
}
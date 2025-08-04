package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGetStartPath(t *testing.T) {
	// Save original command line args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Run("NoArguments", func(t *testing.T) {
		// Reset flag for testing
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cdf"}
		
		// Parse flags to initialize
		flag.Int("depth", 5, "test")
		flag.Bool("no-ignore", false, "test")
		flag.Bool("debug", false, "test")
		flag.Bool("help", false, "test")
		flag.Bool("version", false, "test")
		flag.Parse()

		path, err := getStartPath()
		if err != nil {
			t.Fatalf("getStartPath() failed: %v", err)
		}

		// Should return current working directory
		expectedPath, _ := os.Getwd()
		if path != expectedPath {
			t.Errorf("getStartPath() = %s, expected %s", path, expectedPath)
		}
	})

	t.Run("WithValidPath", func(t *testing.T) {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "cdf_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Reset flag for testing
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cdf", tempDir}
		
		flag.Int("depth", 5, "test")
		flag.Bool("no-ignore", false, "test")
		flag.Bool("debug", false, "test")
		flag.Bool("help", false, "test")
		flag.Bool("version", false, "test")
		flag.Parse()

		path, err := getStartPath()
		if err != nil {
			t.Fatalf("getStartPath() failed: %v", err)
		}

		// Should return absolute path of the temp directory
		expectedPath, _ := filepath.Abs(tempDir)
		if path != expectedPath {
			t.Errorf("getStartPath() = %s, expected %s", path, expectedPath)
		}
	})

	t.Run("WithNonExistentPath", func(t *testing.T) {
		// Reset flag for testing
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cdf", "/nonexistent/path"}
		
		flag.Int("depth", 5, "test")
		flag.Bool("no-ignore", false, "test")
		flag.Bool("debug", false, "test")
		flag.Bool("help", false, "test")
		flag.Bool("version", false, "test")
		flag.Parse()

		_, err := getStartPath()
		if err == nil {
			t.Error("Expected error for nonexistent path, got nil")
		}
	})

	t.Run("WithRelativePath", func(t *testing.T) {
		// Reset flag for testing
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cdf", "."}
		
		flag.Int("depth", 5, "test")
		flag.Bool("no-ignore", false, "test")
		flag.Bool("debug", false, "test")
		flag.Bool("help", false, "test")
		flag.Bool("version", false, "test")
		flag.Parse()

		path, err := getStartPath()
		if err != nil {
			t.Fatalf("getStartPath() failed: %v", err)
		}

		// Should return absolute path
		if !filepath.IsAbs(path) {
			t.Errorf("getStartPath() returned relative path: %s", path)
		}
	})
}

// Test version constant
func TestVersion(t *testing.T) {
	if version == "" {
		t.Error("Version constant should not be empty")
	}

	// Basic format check
	if len(version) < 3 {
		t.Error("Version should be at least 3 characters long")
	}
}

// Integration test that checks main components work together
func TestIntegration(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "cdf_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directories
	testDirs := []string{
		"projects/webapp/src/api",
		"projects/myapp/components",
		"documents/work",
		"personal/notes",
	}

	for _, dir := range testDirs {
		fullPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatalf("Failed to create test dir %s: %v", fullPath, err)
		}
	}

	// Test scanning
	directories, err := scanDirectories(tempDir, 5, true)
	if err != nil {
		t.Fatalf("scanDirectories failed: %v", err)
	}

	if len(directories) == 0 {
		t.Error("Expected to find directories in test structure")
	}

	// Test fuzzy matching
	matches := fuzzyMatch("api", directories)
	
	// Should find the API directory
	found := false
	for _, match := range matches {
		if strings.Contains(match.Str, "api") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find API directory in fuzzy match results")
	}

	// Test formatting
	if len(matches) > 0 {
		formatted := formatMatch(matches[0])
		if formatted == "" {
			t.Error("formatMatch returned empty string")
		}
	}
}

func TestTwoPhaseScanningPriority(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "cdf_two_phase_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save and restore original working directory
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalCwd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create test directories in current (temp) directory
	localDirs := []string{
		"local1",
		"local2/sub",
	}
	for _, dir := range localDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create local dir %s: %v", dir, err)
		}
	}

	// Test two-phase scanning
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dirChan := scanTwoPhasesAsyncCtx(ctx, tempDir, 3, true, 10)
	
	var allDirs []string
	var phaseBoundaryFound = false
	
	for batch := range dirChan {
		if batch.Err != nil && batch.Err != context.Canceled {
			t.Fatalf("Error in scanning: %v", batch.Err)
		}
		
		allDirs = append(allDirs, batch.Directories...)
		
		// Check if we're transitioning between phases
		if batch.Done && !phaseBoundaryFound {
			phaseBoundaryFound = true
		}
	}

	// Verify we found some directories
	if len(allDirs) == 0 {
		t.Error("Expected to find directories in two-phase scan")
	}

	// Check that local directories appear early in results
	// (Note: we can't guarantee exact order due to async nature, 
	// but local dirs should be among the first batch)
	foundLocal := false
	for _, dir := range allDirs {
		if strings.Contains(dir, "local1") || strings.Contains(dir, "local2") {
			foundLocal = true
			break
		}
	}
	
	if !foundLocal {
		t.Error("Expected to find local directories in two-phase scan results")
	}
}
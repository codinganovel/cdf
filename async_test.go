package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAsyncScanning(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	
	// Create test directories
	testDirs := []string{
		"dir1",
		"dir2",
		"dir1/subdir1",
		"dir1/subdir2",
		"dir2/subdir1",
		"dir1/subdir1/deep",
	}
	
	for _, dir := range testDirs {
		if err := os.MkdirAll(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}
	
	// Test that async scanning sends batches progressively
	ch := scanDirectoriesAsync(tempDir, 3, true, 2) // Small batch size to test batching
	
	batches := 0
	totalDirs := 0
	startTime := time.Now()
	
	for batch := range ch {
		batches++
		
		// Count directories in this batch
		dirsInBatch := len(batch.Directories)
		if dirsInBatch > 0 {
			totalDirs += dirsInBatch
		}
		
		// Log timing for first batch (should be fast)
		if batches == 1 && dirsInBatch > 0 {
			elapsed := time.Since(startTime)
			t.Logf("First batch received after %v with %d directories", elapsed, dirsInBatch)
			if elapsed > 100*time.Millisecond {
				t.Error("First batch took too long, should be nearly instant")
			}
		}
		
		if batch.Done {
			t.Logf("Scanning complete: %d total directories in %d batches", totalDirs, batches)
			if batch.Err != nil {
				t.Errorf("Unexpected error: %v", batch.Err)
			}
		}
	}
	
	if batches < 2 {
		t.Errorf("Expected multiple batches with batch size 2, got %d", batches)
	}
	
	if totalDirs != len(testDirs) {
		t.Errorf("Expected %d directories, found %d", len(testDirs), totalDirs)
	}
}

func TestAsyncScanningWithContext(t *testing.T) {
	// Create test directory
	tempDir := t.TempDir()
	
	// Create many directories to ensure scanning takes some time
	for i := 0; i < 50; i++ {
		if err := os.MkdirAll(filepath.Join(tempDir, "dir", "sub", string(rune('a'+i%26))), 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}
	
	// Test cancellation
	ctx, cancel := context.WithCancel(context.Background())
	ch := scanDirectoriesAsyncCtx(ctx, tempDir, 5, true, 10)
	
	// Cancel after receiving first batch
	gotFirstBatch := false
	for batch := range ch {
		if !gotFirstBatch {
			gotFirstBatch = true
			cancel() // Cancel scanning
		}
		
		if batch.Done && batch.Err == context.Canceled {
			t.Log("Scanning was properly cancelled")
			return
		}
	}
	
	if !gotFirstBatch {
		t.Error("Never received first batch")
	}
}

func BenchmarkSyncVsAsync(b *testing.B) {
	b.Run("Sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dirs, _ := scanDirectories(".", 3, true)
			_ = dirs
		}
	})
	
	b.Run("AsyncFirstBatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ch := scanDirectoriesAsync(".", 3, true, 10)
			// Measure time to first batch
			<-ch
			// Drain channel
			for range ch {
			}
		}
	})
}
package main

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
)

var ignorePatterns = []string{
	".git", "node_modules", "target", ".cache", "vendor",
	"__pycache__", ".pytest_cache", "dist", "build",
	".terraform", ".vscode", ".idea",
}

func scanDirectories(root string, maxDepth int, useIgnorePatterns bool) ([]string, error) {
	var directories []string
	
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		
		if !d.IsDir() {
			return nil
		}
		
		if path == root {
			return nil
		}
		
		if !isWithinDepth(path, root, maxDepth) {
			return filepath.SkipDir
		}
		
		if useIgnorePatterns && shouldIgnore(d.Name()) {
			return filepath.SkipDir
		}
		
		directories = append(directories, path)
		return nil
	})
	
	return directories, err
}

// ScanConfig holds configuration for directory scanning
type ScanConfig struct {
	Root              string
	MaxDepth          int
	UseIgnorePatterns bool
	InitialBatchSize  int
	MaxBatchSize      int
}

// scanDirectoriesAsync scans directories and sends results through a channel in batches
func scanDirectoriesAsync(root string, maxDepth int, useIgnorePatterns bool, batchSize int) <-chan DirBatch {
	return scanDirectoriesAsyncCtx(context.Background(), root, maxDepth, useIgnorePatterns, batchSize)
}

// scanDirectoriesAsyncCtx scans directories with context support for cancellation
func scanDirectoriesAsyncCtx(ctx context.Context, root string, maxDepth int, useIgnorePatterns bool, batchSize int) <-chan DirBatch {
	config := ScanConfig{
		Root:              root,
		MaxDepth:          maxDepth,
		UseIgnorePatterns: useIgnorePatterns,
		InitialBatchSize:  batchSize,
		MaxBatchSize:      200, // Cap batch size for responsiveness
	}
	return scanWithConfigCtx(ctx, config)
}

// scanDirectoriesAsyncCtxExcluding scans directories while excluding a specific path
func scanDirectoriesAsyncCtxExcluding(ctx context.Context, root string, maxDepth int, useIgnorePatterns bool, batchSize int, excludePath string) <-chan DirBatch {
	config := ScanConfig{
		Root:              root,
		MaxDepth:          maxDepth,
		UseIgnorePatterns: useIgnorePatterns,
		InitialBatchSize:  batchSize,
		MaxBatchSize:      200,
	}
	return scanWithConfigCtxExcluding(ctx, config, excludePath)
}

func scanWithConfig(config ScanConfig) <-chan DirBatch {
	return scanWithConfigCtx(context.Background(), config)
}

func scanWithConfigCtxExcluding(ctx context.Context, config ScanConfig, excludePath string) <-chan DirBatch {
	// Use smaller buffer to provide backpressure
	ch := make(chan DirBatch, 2)
	
	go func() {
		defer close(ch)
		
		var batch []string
		batch = make([]string, 0, config.InitialBatchSize)
		currentBatchSize := config.InitialBatchSize
		dirCount := 0
		
		// Custom walk function that respects context cancellation and excludes a path
		err := walkDirContext(ctx, config.Root, func(path string, d fs.DirEntry, err error) error {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			
			if err != nil {
				// Log permission errors but continue scanning
				return nil
			}
			
			if !d.IsDir() {
				return nil
			}
			
			if path == config.Root {
				return nil
			}
			
			// Skip the excluded path and all its subdirectories
			if excludePath != "" && (path == excludePath || strings.HasPrefix(path, excludePath+string(filepath.Separator))) {
				return filepath.SkipDir
			}
			
			if !isWithinDepth(path, config.Root, config.MaxDepth) {
				return filepath.SkipDir
			}
			
			if config.UseIgnorePatterns && shouldIgnore(d.Name()) {
				return filepath.SkipDir
			}
			
			batch = append(batch, path)
			dirCount++
			
			// Send batch when it reaches the size limit
			if len(batch) >= currentBatchSize {
				// Create new slice to avoid data races
				sendBatch := make([]string, len(batch))
				copy(sendBatch, batch)
				
				select {
				case ch <- DirBatch{
					Directories: sendBatch,
					Done:        false,
				}:
				case <-ctx.Done():
					return ctx.Err()
				}
				
				// Adaptive batch sizing: increase batch size for large directories
				if dirCount > 500 && currentBatchSize < config.MaxBatchSize {
					currentBatchSize = min(currentBatchSize*2, config.MaxBatchSize)
				}
				
				batch = batch[:0] // Reset slice but keep capacity
			}
			
			return nil
		})
		
		// Handle context cancellation
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		
		// Send any remaining directories
		if len(batch) > 0 || err != nil {
			select {
			case ch <- DirBatch{
				Directories: batch,
				Done:        true,
				Err:         err,
			}:
			case <-ctx.Done():
			}
		} else {
			// Send done signal even if no remaining batch
			select {
			case ch <- DirBatch{
				Directories: nil,
				Done:        true,
				Err:         err,
			}:
			case <-ctx.Done():
			}
		}
	}()
	
	return ch
}

func scanWithConfigCtx(ctx context.Context, config ScanConfig) <-chan DirBatch {
	// Use smaller buffer to provide backpressure
	ch := make(chan DirBatch, 2)
	
	go func() {
		defer close(ch)
		
		var batch []string
		batch = make([]string, 0, config.InitialBatchSize)
		currentBatchSize := config.InitialBatchSize
		dirCount := 0
		
		// Custom walk function that respects context cancellation
		err := walkDirContext(ctx, config.Root, func(path string, d fs.DirEntry, err error) error {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			
			if err != nil {
				// Log permission errors but continue scanning
				return nil
			}
			
			if !d.IsDir() {
				return nil
			}
			
			if path == config.Root {
				return nil
			}
			
			if !isWithinDepth(path, config.Root, config.MaxDepth) {
				return filepath.SkipDir
			}
			
			if config.UseIgnorePatterns && shouldIgnore(d.Name()) {
				return filepath.SkipDir
			}
			
			batch = append(batch, path)
			dirCount++
			
			// Send batch when it reaches the size limit
			if len(batch) >= currentBatchSize {
				// Create new slice to avoid data races
				sendBatch := make([]string, len(batch))
				copy(sendBatch, batch)
				
				select {
				case ch <- DirBatch{
					Directories: sendBatch,
					Done:        false,
				}:
				case <-ctx.Done():
					return ctx.Err()
				}
				
				// Adaptive batch sizing: increase batch size for large directories
				if dirCount > 500 && currentBatchSize < config.MaxBatchSize {
					currentBatchSize = min(currentBatchSize*2, config.MaxBatchSize)
				}
				
				batch = batch[:0] // Reset slice but keep capacity
			}
			
			return nil
		})
		
		// Handle context cancellation
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		
		// Send any remaining directories
		if len(batch) > 0 || err != nil {
			select {
			case ch <- DirBatch{
				Directories: batch,
				Done:        true,
				Err:         err,
			}:
			case <-ctx.Done():
			}
		} else {
			// Send done signal even if no remaining batch
			select {
			case ch <- DirBatch{
				Directories: nil,
				Done:        true,
				Err:         err,
			}:
			case <-ctx.Done():
			}
		}
	}()
	
	return ch
}

// walkDirContext is a wrapper around filepath.WalkDir that respects context cancellation
func walkDirContext(ctx context.Context, root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Check context before processing
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return fn(path, d, err)
		}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DirBatch represents a batch of discovered directories
type DirBatch struct {
	Directories []string // New directories in this batch
	Done        bool     // Whether scanning is complete
	Err         error    // Any error that occurred
}

func shouldIgnore(name string) bool {
	for _, pattern := range ignorePatterns {
		if name == pattern {
			return true
		}
	}
	return false
}

func isWithinDepth(path, root string, maxDepth int) bool {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	
	depth := strings.Count(relPath, string(filepath.Separator))
	return depth <= maxDepth
}
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/codinganovel/autocd-go"
)

const version = "1.0.0"

func main() {
	var (
		depth     = flag.Int("depth", 5, "Maximum scan depth")
		noIgnore  = flag.Bool("no-ignore", false, "Disable ignore patterns")
		debug     = flag.Bool("debug", false, "Enable debug output")
		showHelp  = flag.Bool("help", false, "Show usage information")
		showVer   = flag.Bool("version", false, "Show version information")
	)
	
	flag.Parse()
	
	if *showHelp {
		showUsage()
		os.Exit(0)
	}
	
	if *showVer {
		fmt.Printf("cdf version %s\n", version)
		os.Exit(0)
	}
	
	startPath, err := getStartPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	if *debug {
		fmt.Fprintf(os.Stderr, "Scanning from: %s (depth: %d)\n", startPath, *depth)
	}
	
	// Create a context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle interrupt signals for clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if *debug {
			fmt.Fprintf(os.Stderr, "\nReceived interrupt, cleaning up...\n")
		}
		cancel()
	}()
	
	// Two-phase scanning for prioritized results
	dirChan := scanTwoPhasesAsyncCtx(ctx, startPath, *depth, !*noIgnore, 50)
	
	selectedPath, err := runTUIAsyncCtx(ctx, dirChan)
	if err != nil {
		if *debug {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		}
		// Check if it was a cancellation vs actual error
		if err == context.Canceled || err.Error() == "cancelled" {
			os.Exit(2)
		}
		os.Exit(1)
	}
	
	if *debug {
		fmt.Fprintf(os.Stderr, "Selected: %s\n", selectedPath)
	}
	
	if err := autocd.ExitWithDirectory(selectedPath); err != nil {
		fmt.Fprintf(os.Stderr, "autocd failed: %v\n", err)
		os.Exit(1)
	}
}

func getStartPath() (string, error) {
	args := flag.Args()
	if len(args) > 0 {
		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("invalid path %s: %v", path, err)
		}
		
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", absPath)
		}
		
		return absPath, nil
	}
	
	return os.Getwd()
}

// scanTwoPhasesAsyncCtx implements two-phase scanning:
// Phase 1: Current working directory (fast results)
// Phase 2: Root directory excluding current directory (broader coverage)
func scanTwoPhasesAsyncCtx(ctx context.Context, startPath string, maxDepth int, useIgnorePatterns bool, batchSize int) <-chan DirBatch {
	ch := make(chan DirBatch, 2)
	
	go func() {
		defer close(ch)
		
		cwd, err := os.Getwd()
		if err != nil {
			// Fallback to single-phase if we can't get CWD
			singlePhase := scanDirectoriesAsyncCtx(ctx, startPath, maxDepth, useIgnorePatterns, batchSize)
			for batch := range singlePhase {
				select {
				case ch <- batch:
				case <-ctx.Done():
					return
				}
			}
			return
		}
		
		// Phase 1: Scan current working directory first
		phase1Chan := scanDirectoriesAsyncCtx(ctx, cwd, maxDepth, useIgnorePatterns, batchSize)
		for batch := range phase1Chan {
			select {
			case ch <- DirBatch{
				Directories: batch.Directories,
				Done:        false, // Not done yet, phase 2 coming
				Err:         batch.Err,
			}:
			case <-ctx.Done():
				return
			}
			
			if batch.Err != nil {
				return
			}
		}
		
		// Phase 2: Scan from root, excluding current directory
		phase2Chan := scanDirectoriesAsyncCtxExcluding(ctx, "/", maxDepth, useIgnorePatterns, batchSize, cwd)
		for batch := range phase2Chan {
			select {
			case ch <- batch:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return ch
}

func showUsage() {
	fmt.Printf(`cdf - Directory Fuzzy Finder

Usage:
  cdf [options] [path]

Arguments:
  path              Starting directory for scan (default: current directory)

Options:
  --depth <n>       Maximum scan depth (default: 5)
  --no-ignore       Disable ignore patterns (scan all directories)
  --debug           Enable debug output to stderr
  --help            Show this help message
  --version         Show version information

Examples:
  cdf                    # Launch from current directory
  cdf /path/to/start     # Launch from specific directory
  cdf --depth 3          # Limit scan depth to 3 levels
  cdf --debug            # Enable debug output

Keyboard shortcuts:
  ↑↓ or j/k             Navigate results
  Type                  Filter results
  Enter                 Select directory
  Escape                Cancel

Exit codes:
  0                     Successful directory selection
  1                     Error in scanning or autocd failure
  2                     User cancelled (Escape pressed)
`)
}
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	
	directories, err := scanDirectories(startPath, *depth, !*noIgnore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directories: %v\n", err)
		os.Exit(1)
	}
	
	if len(directories) == 0 {
		fmt.Fprintf(os.Stderr, "No directories found\n")
		os.Exit(1)
	}
	
	if *debug {
		fmt.Fprintf(os.Stderr, "Found %d directories\n", len(directories))
	}
	
	selectedPath, err := runTUI(directories)
	if err != nil {
		if *debug {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		}
		os.Exit(2)
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
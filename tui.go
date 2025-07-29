package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/sahilm/fuzzy"
)

// uiState encapsulates all mutable state with proper synchronization
type uiState struct {
	mu           sync.RWMutex
	query        string
	selected     int
	scrollOffset int
	directories  []string
	matches      []fuzzy.Match
	scanComplete bool
}

func runTUI(directories []string) (string, error) {
	// For backward compatibility - convert to async version
	ch := make(chan DirBatch, 1)
	ch <- DirBatch{Directories: directories, Done: true}
	close(ch)
	return runTUIAsync(ch)
}

func runTUIAsync(dirChan <-chan DirBatch) (string, error) {
	return runTUIAsyncCtx(context.Background(), dirChan)
}

func runTUIAsyncCtx(ctx context.Context, dirChan <-chan DirBatch) (string, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return "", err
	}
	
	if err := screen.Init(); err != nil {
		return "", err
	}
	defer screen.Fini()
	
	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	screen.Clear()
	
	// UI state - encapsulated for thread safety
	state := &uiState{
		directories: make([]string, 0, 1000), // Pre-allocate for performance
		matches:     make([]fuzzy.Match, 0),
	}
	
	// Start a goroutine to receive directory updates
	updateChan := make(chan struct{}, 1)
	errorChan := make(chan error, 1)
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case batch, ok := <-dirChan:
				if !ok {
					close(updateChan)
					return
				}
				
				state.mu.Lock()
				// Check for errors
				if batch.Err != nil && batch.Err != context.Canceled {
					select {
					case errorChan <- batch.Err:
					default:
					}
				}
				
				// Append new directories
				if len(batch.Directories) > 0 {
					state.directories = append(state.directories, batch.Directories...)
					// Re-run fuzzy match on the updated list
					state.matches = fuzzyMatch(state.query, state.directories)
				}
				
				state.scanComplete = batch.Done
				state.mu.Unlock()
				
				// Non-blocking send to trigger refresh
				select {
				case updateChan <- struct{}{}:
				default:
				}
			}
		}
	}()
	
	// Main event loop
	eventCtx, cancelEvents := context.WithCancel(ctx)
	defer cancelEvents()
	
	// Separate goroutine for handling updates
	go func() {
		for {
			select {
			case <-eventCtx.Done():
				return
			case _, ok := <-updateChan:
				if !ok {
					return
				}
				screen.PostEvent(tcell.NewEventInterrupt(nil))
			}
		}
	}()
	
	for {
		// Check for scanning errors
		select {
		case err := <-errorChan:
			if err != nil && err != context.Canceled {
				return "", fmt.Errorf("scanning error: %w", err)
			}
		default:
		}
		// Render current state
		state.mu.RLock()
		updateDisplayAsync(screen, state.matches, state.query, state.selected, 
			state.scrollOffset, len(state.directories), state.scanComplete)
		state.mu.RUnlock()
		screen.Show()
		
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			result := handleKeyEventState(ev, state, screen)
			
			if result != 0 {
				state.mu.RLock()
				defer state.mu.RUnlock()
				
				if result == 1 && state.selected >= 0 && state.selected < len(state.matches) {
					return state.matches[state.selected].Str, nil
				}
				return "", fmt.Errorf("cancelled")
			}
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventInterrupt:
			// Directory update received - will refresh on next loop
		}
	}
}

// handleKeyEventState handles keyboard input with proper state management
func handleKeyEventState(event *tcell.EventKey, state *uiState, screen tcell.Screen) int {
	_, height := screen.Size()
	maxDisplay := height - 5
	
	state.mu.Lock()
	defer state.mu.Unlock()
	
	switch event.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlQ:
		return -1
	case tcell.KeyEnter:
		return 1
	case tcell.KeyUp:
		if state.selected > 0 {
			state.selected--
			if state.selected < state.scrollOffset {
				state.scrollOffset = state.selected
			}
		}
	case tcell.KeyDown:
		if state.selected < len(state.matches)-1 {
			state.selected++
			if state.selected >= state.scrollOffset+maxDisplay {
				state.scrollOffset = state.selected - maxDisplay + 1
			}
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(state.query) > 0 {
			state.query = state.query[:len(state.query)-1]
			state.matches = fuzzyMatch(state.query, state.directories)
			state.selected = 0
			state.scrollOffset = 0
		}
	case tcell.KeyRune:
		state.query += string(event.Rune())
		state.matches = fuzzyMatch(state.query, state.directories)
		state.selected = 0
		state.scrollOffset = 0
	}
	return 0
}

func handleKeyEvent(event *tcell.EventKey, query *string, selected *int, matches *[]fuzzy.Match, directories []string, screen tcell.Screen, scrollOffset int) (int, int) {
	_, height := screen.Size()
	maxDisplay := height - 5
	
	switch event.Key() {
	case tcell.KeyEscape:
		return -1, scrollOffset
	case tcell.KeyEnter:
		return 1, scrollOffset
	case tcell.KeyCtrlQ:
		return -1, scrollOffset
	case tcell.KeyUp:
		if *selected > 0 {
			*selected--
			if *selected < scrollOffset {
				scrollOffset = *selected
			}
		}
	case tcell.KeyDown:
		if *selected < len(*matches)-1 {
			*selected++
			if *selected >= scrollOffset+maxDisplay {
				scrollOffset = *selected - maxDisplay + 1
			}
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(*query) > 0 {
			*query = (*query)[:len(*query)-1]
			*matches = fuzzyMatch(*query, directories)
			if *selected >= len(*matches) {
				*selected = len(*matches) - 1
			}
			if *selected < 0 {
				*selected = 0
			}
			scrollOffset = 0
		}
	case tcell.KeyRune:
		*query += string(event.Rune())
		*matches = fuzzyMatch(*query, directories)
		*selected = 0
		scrollOffset = 0
	}
	return 0, scrollOffset
}

func updateDisplay(screen tcell.Screen, matches []fuzzy.Match, query string, selected int, scrollOffset int) {
	updateDisplayAsync(screen, matches, query, selected, scrollOffset, len(matches), true)
}

func updateDisplayAsync(screen tcell.Screen, matches []fuzzy.Match, query string, selected int, scrollOffset int, totalDirs int, scanComplete bool) {
	screen.Clear()
	
	width, height := screen.Size()
	
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	promptStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen)
	
	prompt := fmt.Sprintf("cdf > %s", query)
	drawText(screen, 0, 0, promptStyle, prompt)
	
	drawText(screen, 0, 1, style, "")
	
	startY := 3
	maxDisplay := height - 5
	
	endIndex := scrollOffset + maxDisplay
	if endIndex > len(matches) {
		endIndex = len(matches)
	}
	
	for i := scrollOffset; i < endIndex; i++ {
		if i >= len(matches) {
			break
		}
		
		match := matches[i]
		dir := formatMatch(match)
		score := getMatchScore(match)
		
		line := fmt.Sprintf("  📁 %-50s [%d%%]", dir, score)
		if len(line) > width-2 {
			line = line[:width-2]
		}
		
		displayIndex := i - scrollOffset
		if i == selected {
			selectedStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
			drawText(screen, 0, startY+displayIndex, selectedStyle, line)
		} else {
			drawText(screen, 0, startY+displayIndex, style, line)
		}
	}
	
	var status string
	scanStatus := ""
	if !scanComplete {
		scanStatus = fmt.Sprintf(" • Scanning... (%d dirs)", totalDirs)
	}
	
	if len(matches) > maxDisplay {
		start := scrollOffset + 1
		end := endIndex
		status = fmt.Sprintf("[%d matches, showing %d-%d]%s • ↑↓ navigate • Enter select • Esc/Ctrl+Q cancel", len(matches), start, end, scanStatus)
	} else {
		status = fmt.Sprintf("[%d matches]%s • ↑↓ navigate • Enter select • Esc/Ctrl+Q cancel", len(matches), scanStatus)
	}
	drawText(screen, 0, height-1, style, status)
}

func drawText(screen tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range text {
		screen.SetContent(x+i, y, r, nil, style)
	}
}
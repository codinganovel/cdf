package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/sahilm/fuzzy"
)

func runTUI(directories []string) (string, error) {
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
	
	query := ""
	selected := 0
	scrollOffset := 0
	matches := fuzzyMatch(query, directories)
	
	for {
		updateDisplay(screen, matches, query, selected, scrollOffset)
		screen.Show()
		
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			result, newScrollOffset := handleKeyEvent(ev, &query, &selected, &matches, directories, screen, scrollOffset)
			scrollOffset = newScrollOffset
			if result != 0 {
				if result == 1 && selected >= 0 && selected < len(matches) {
					return matches[selected].Str, nil
				}
				return "", fmt.Errorf("cancelled")
			}
		case *tcell.EventResize:
			screen.Sync()
		}
	}
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
	if len(matches) > maxDisplay {
		start := scrollOffset + 1
		end := endIndex
		status = fmt.Sprintf("[%d matches, showing %d-%d] • ↑↓ navigate • Enter select • Esc/Ctrl+Q cancel", len(matches), start, end)
	} else {
		status = fmt.Sprintf("[%d matches] • ↑↓ navigate • Enter select • Esc/Ctrl+Q cancel", len(matches))
	}
	drawText(screen, 0, height-1, style, status)
}

func drawText(screen tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range text {
		screen.SetContent(x+i, y, r, nil, style)
	}
}
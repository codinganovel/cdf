package main

import (
	"os"
	"strings"

	"github.com/sahilm/fuzzy"
)

func fuzzyMatch(query string, directories []string) []fuzzy.Match {
	if query == "" {
		var matches []fuzzy.Match
		for i, dir := range directories {
			matches = append(matches, fuzzy.Match{
				Str:            dir,
				Index:          i,
				Score:          100,
				MatchedIndexes: []int{},
			})
		}
		return matches
	}
	
	return fuzzy.Find(query, directories)
}

func formatMatch(match fuzzy.Match) string {
	dir := match.Str
	
	if strings.HasPrefix(dir, homeDir()) {
		dir = "~" + strings.TrimPrefix(dir, homeDir())
	}
	
	return dir
}

func getMatchScore(match fuzzy.Match) int {
	if match.Score < 0 {
		return 0
	}
	return match.Score
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}
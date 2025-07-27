package main

import (
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
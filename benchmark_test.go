package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sahilm/fuzzy"
)

func BenchmarkScanDirectories(b *testing.B) {
	// Create a realistic test directory structure
	tempDir, err := os.MkdirTemp("", "cdf_benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a moderately complex directory structure
	testDirs := []string{
		"projects/webapp/src/components",
		"projects/webapp/src/utils",
		"projects/webapp/src/pages",
		"projects/webapp/public/assets",
		"projects/api/controllers",
		"projects/api/models",
		"projects/api/routes",
		"projects/api/middleware",
		"documents/work/reports",
		"documents/work/presentations",
		"documents/personal/photos",
		"documents/personal/notes",
		"tools/scripts/deployment",
		"tools/scripts/testing",
		"tools/configs/nginx",
		"tools/configs/docker",
		"cache/build/temp", // will be ignored
		"node_modules/lib", // will be ignored
		".git/hooks",       // will be ignored
	}

	for _, dir := range testDirs {
		fullPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			b.Fatalf("Failed to create test dir %s: %v", fullPath, err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := scanDirectories(tempDir, 5, true)
		if err != nil {
			b.Fatalf("scanDirectories failed: %v", err)
		}
	}
}

func BenchmarkScanDirectoriesLarge(b *testing.B) {
	// Create a larger directory structure for stress testing
	tempDir, err := os.MkdirTemp("", "cdf_benchmark_large")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create many directories
	for i := 0; i < 50; i++ {
		for j := 0; j < 10; j++ {
			dir := filepath.Join(tempDir, "level1_"+string(rune(i+'a')), "level2_"+string(rune(j+'0')))
			if err := os.MkdirAll(dir, 0755); err != nil {
				b.Fatalf("Failed to create test dir %s: %v", dir, err)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := scanDirectories(tempDir, 5, true)
		if err != nil {
			b.Fatalf("scanDirectories failed: %v", err)
		}
	}
}

func BenchmarkFuzzyMatch(b *testing.B) {
	// Create a realistic set of directories
	directories := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		switch i % 10 {
		case 0:
			directories[i] = "/home/user/projects/webapp/src/components/Button"
		case 1:
			directories[i] = "/home/user/projects/api/controllers/UserController"
		case 2:
			directories[i] = "/home/user/documents/work/reports/quarterly"
		case 3:
			directories[i] = "/usr/local/lib/node_modules/react"
		case 4:
			directories[i] = "/var/log/applications/myapp"
		case 5:
			directories[i] = "/opt/tools/deployment/scripts"
		case 6:
			directories[i] = "/etc/nginx/sites-available"
		case 7:
			directories[i] = "/home/user/personal/photos/vacation"
		case 8:
			directories[i] = "/tmp/build/artifacts/release"
		case 9:
			directories[i] = "/home/user/workspace/backend/api/routes"
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = fuzzyMatch("api", directories)
	}
}

func BenchmarkFuzzyMatchEmpty(b *testing.B) {
	directories := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		directories[i] = "/some/path/number/" + string(rune(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = fuzzyMatch("", directories)
	}
}

func BenchmarkShouldIgnore(b *testing.B) {
	testNames := []string{
		".git", "node_modules", "target", "regular_dir",
		"src", "lib", "bin", "dist", "build", "cache",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, name := range testNames {
			shouldIgnore(name)
		}
	}
}

func BenchmarkIsWithinDepth(b *testing.B) {
	root := "/home/user"
	paths := []string{
		"/home/user/level1",
		"/home/user/level1/level2",
		"/home/user/level1/level2/level3",
		"/home/user/projects/webapp/src/components",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			isWithinDepth(path, root, 5)
		}
	}
}

func BenchmarkFormatMatch(b *testing.B) {
	matches := []string{
		"/home/user/projects/webapp",
		"/var/log/applications",
		"/usr/local/bin",
		"/home/user/documents/work",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, match := range matches {
			formatMatch(fuzzy.Match{Str: match})
		}
	}
}
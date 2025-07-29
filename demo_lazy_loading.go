// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Demo script to show lazy loading in action
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run demo_lazy_loading.go <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]
	fmt.Printf("Starting lazy scan of %s...\n\n", dir)

	ctx := context.Background()
	ch := scanDirectoriesAsyncCtx(ctx, dir, 5, true, 50)

	startTime := time.Now()
	batches := 0
	totalDirs := 0

	for batch := range ch {
		batches++
		dirsInBatch := len(batch.Directories)
		totalDirs += dirsInBatch

		elapsed := time.Since(startTime)
		fmt.Printf("Batch %d: %d dirs (total: %d) after %v\n", 
			batches, dirsInBatch, totalDirs, elapsed)

		if batch.Done {
			fmt.Printf("\nScanning complete!\n")
			fmt.Printf("Total: %d directories in %d batches\n", totalDirs, batches)
			fmt.Printf("Total time: %v\n", elapsed)
			if batch.Err != nil {
				fmt.Printf("Error: %v\n", batch.Err)
			}
		}

		// Simulate processing delay to see batching in action
		time.Sleep(10 * time.Millisecond)
	}
}
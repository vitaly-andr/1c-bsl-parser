package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LazarenkoA/1c-language-parser/ast"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: parseall <dir>")
		os.Exit(1)
	}

	dir := os.Args[1]
	var files []string

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == ".bsl" {
			files = append(files, path)
		}
		return nil
	})

	fmt.Printf("Found %d BSL files\n", len(files))

	var success, failed int64
	var failedFiles []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	start := time.Now()

	for _, f := range files {
		wg.Add(1)
		sem <- struct{}{}
		go func(path string) {
			defer wg.Done()
			defer func() { <-sem }()

			data, err := os.ReadFile(path)
			if err != nil {
				atomic.AddInt64(&failed, 1)
				return
			}

			a := ast.NewAST(string(data))
			if err := a.Parse(); err != nil {
				atomic.AddInt64(&failed, 1)
				mu.Lock()
				if len(failedFiles) < 500 {
					failedFiles = append(failedFiles, path+": "+err.Error())
				}
				mu.Unlock()
			} else {
				atomic.AddInt64(&success, 1)
			}
		}(f)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Success: %d\n", success)
	fmt.Printf("Failed:  %d\n", failed)
	fmt.Printf("Time:    %v\n", elapsed)
	fmt.Printf("Speed:   %.0f files/sec\n", float64(success+failed)/elapsed.Seconds())

	if len(failedFiles) > 0 {
		fmt.Println("\nFirst failed files:")
		for _, f := range failedFiles {
			fmt.Println("  ", f)
		}
	}
}
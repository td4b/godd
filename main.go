package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/schollz/progressbar/v3"
)

func main() {
	// Define flags for input, output, block size, and workers
	inputFile := flag.String("if", "", "Input file (required)")
	outputFile := flag.String("of", "", "Output file (required)")
	blockSize := flag.String("bs", "", "Block size in bytes (default: 512)")
	workers := flag.Int("workers", 4, "Number of concurrent workers (default: 4)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  go run dd.go -if=input.txt -of=output.txt -bs=1024 -workers=4\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Error: Both input file (-if) and output file (-of) are required.")
		flag.Usage()
		os.Exit(1)
	}

	// Convert block size to integer, use default if not set or invalid
	bs := 512 // Default block size
	if *blockSize != "" {
		var err error
		bs, err = strconv.Atoi(*blockSize)
		if err != nil || bs <= 0 {
			fmt.Printf("Warning: Invalid block size: %s. Using default block size of 512 bytes.\n", *blockSize)
			bs = 512
		}
	}

	// Open input file
	input, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error: Failed to open input file: %s\n", err)
		os.Exit(1)
	}
	defer input.Close()

	// Get input file size
	inputInfo, err := input.Stat()
	if err != nil {
		fmt.Printf("Error: Failed to get input file size: %s\n", err)
		os.Exit(1)
	}
	fileSize := inputInfo.Size()
	totalBlocks := int((fileSize + int64(bs) - 1) / int64(bs)) // Round up

	// Open output file (truncates by default)
	output, err := os.Create(*outputFile) // Truncates the file to zero length
	if err != nil {
		fmt.Printf("Error: Failed to create or clear output file: %s\n", err)
		os.Exit(1)
	}
	defer output.Close()

	// Display summary of operation
	fmt.Println("Starting copy operation:")
	fmt.Printf("  Input File: %s\n", *inputFile)
	fmt.Printf("  Output File: %s\n", *outputFile)
	fmt.Printf("  Block Size: %d bytes\n", bs)
	fmt.Printf("  Total Blocks: %d\n", totalBlocks)
	fmt.Printf("  Workers: %d\n", *workers)
	fmt.Println("=================================")

	// Initialize progress bar
	bar := progressbar.NewOptions(totalBlocks,
		progressbar.OptionSetDescription("Copying..."),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionClearOnFinish(),
	)

	// Worker pool for concurrent copying
	var wg sync.WaitGroup
	blockChan := make(chan int, *workers)

	// Worker function
	worker := func() {
		defer wg.Done()
		for block := range blockChan {
			offset := int64(block) * int64(bs)
			buffer := make([]byte, bs)

			// Read block
			n, err := input.ReadAt(buffer, offset)
			if err != nil && err.Error() != "EOF" {
				fmt.Printf("Error: Failed to read block %d: %s\n", block, err)
				return
			}

			// Write block
			_, err = output.WriteAt(buffer[:n], offset)
			if err != nil {
				fmt.Printf("Error: Failed to write block %d: %s\n", block, err)
				return
			}

			// Update progress bar
			_ = bar.Add(1)
		}
	}

	// Start workers
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker()
	}

	// Send blocks to workers
	for block := 0; block < totalBlocks; block++ {
		blockChan <- block
	}
	close(blockChan)

	// Wait for workers to finish
	wg.Wait()

	// Finish progress bar
	_ = bar.Finish()

	fmt.Printf("Copy operation completed successfully.\n")
	fmt.Printf("Copied %d blocks (%d bytes).\n", totalBlocks, fileSize)
}

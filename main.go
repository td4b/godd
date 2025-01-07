package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/klauspost/pgzip"
	"github.com/schollz/progressbar/v3"
)

func isGzip(file *os.File) (bool, error) {
	head := make([]byte, 2)
	if _, err := file.Read(head); err != nil {
		return false, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return false, err
	}
	return head[0] == 0x1f && head[1] == 0x8b, nil
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func main() {
	// Define flags for input, output, block size, and workers
	inputFile := flag.String("if", "", "Input file (required)")
	outputFile := flag.String("of", "", "Output file (required)")
	blockSize := flag.String("bs", "", "Block size in bytes (default: 512)")

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

	// Check if file is gzip
	gzipDetected, err := isGzip(input)
	if err != nil {
		fmt.Printf("Error: Failed to detect gzip format: %s\n", err)
		os.Exit(1)
	}

	// Create output file
	output, err := os.Create(*outputFile)
	if err != nil {
		fmt.Printf("Error: Failed to create or clear output file: %s\n", err)
		os.Exit(1)
	}
	defer output.Close()

	// Initialize progress bar
	fmt.Println("Starting operation...")
	var bar *progressbar.ProgressBar

	if gzipDetected {
		fmt.Println("Detected gzip file format. Processing with decompression...")
		gzipReader, err := pgzip.NewReader(input)
		if err != nil {
			fmt.Printf("Error: Failed to initialize gzip reader: %s\n", err)
			os.Exit(1)
		}
		defer gzipReader.Close()

		bar = progressbar.NewOptions64(-1,
			progressbar.OptionSetDescription("Decompressing"),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionShowBytes(true),
			progressbar.OptionOnCompletion(func() {
				fmt.Printf("\nDecompressed file size: Unknown (streaming).\n")
			}),
		)

		_, err = io.Copy(output, io.TeeReader(gzipReader, bar))
		if err != nil {
			fmt.Printf("Error: Failed to decompress file: %s\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Plain file detected. Processing without compression...")

		inputInfo, err := input.Stat()
		if err != nil {
			fmt.Printf("Error: Failed to get input file size: %s\n", err)
			os.Exit(1)
		}
		fileSize := inputInfo.Size()
		bar = progressbar.NewOptions64(fileSize,
			progressbar.OptionSetDescription("Copying"),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionShowBytes(true),
			progressbar.OptionOnCompletion(func() {
				fmt.Printf("\nCopied file size: %s\n", formatBytes(fileSize))
			}),
		)

		_, err = io.Copy(output, io.TeeReader(input, bar))
		if err != nil {
			fmt.Printf("Error: Failed to copy file: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Operation completed successfully.")
}

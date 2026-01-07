//go:build ignore

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Entry struct {
	UniqueID   string  `json:"unique_identifier"`
	NativeWord string  `json:"native word"`
	English    string  `json:"english word"`
	Source     string  `json:"source"`
	Score      float64 `json:"score"`
}

func convertFile(jsonPath string) error {
	// Open JSON file
	file, err := os.Open(jsonPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create CSV file
	csvPath := strings.TrimSuffix(jsonPath, ".json") + ".csv"
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"id", "native", "english", "source", "score"}); err != nil {
		return err
	}

	// Read and convert line by line
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024) // 1MB buffer
	scanner.Buffer(buf, 10*1024*1024) // Max 10MB per line

	count := 0
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		record := []string{
			entry.UniqueID,
			entry.NativeWord,
			entry.English,
			entry.Source,
			fmt.Sprintf("%.6f", entry.Score),
		}

		if err := writer.Write(record); err != nil {
			return err
		}
		count++
	}

	fmt.Printf("  %s: %d entries\n", filepath.Base(csvPath), count)
	return scanner.Err()
}

func main() {
	dir := "datasets/aksharantar"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// Find all JSON files
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Converting %d JSON files to CSV...\n\n", len(files))

	for _, f := range files {
		fmt.Printf("Converting %s...\n", filepath.Base(f))
		if err := convertFile(f); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		}
	}

	fmt.Println("\nDone!")
}

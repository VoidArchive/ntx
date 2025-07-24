package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

func ImportCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	fmt.Printf("Columns: %v\n", header)
	fmt.Printf("Total columns: %d\n", len(header))

	records := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read record %d: %w", records+1, err)
		}

		records++

		// Process each record here
		processRecord(header, record, records)
	}

	fmt.Printf("Successfully processed %d records\n", records)
	return nil
}

func processRecord(header []string, record []string, lineNum int) {
	// Basic validation
	if len(record) != len(header) {
		log.Printf("Warning: Record %d has %d fields, expected %d", lineNum, len(record), len(header))
	}

	// Example processing - modify based on your needs
	fmt.Printf("Record %d: ", lineNum)
	for i, field := range record {
		if i < len(header) {
			fmt.Printf("%s=%s ", header[i], field)
		}
	}
	fmt.Println()
}

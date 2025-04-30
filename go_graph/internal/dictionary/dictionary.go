// Package dictionary provides functionality for loading and manipulating dictionaries
package dictionary

import (
	"encoding/csv"
	"fmt"
	"os"
)

// LoadFromCSV loads a dictionary from a CSV file
// The CSV file should have two columns: word and definition
func LoadFromCSV(filepath string) (map[string]string, error) {
	// Open the CSV file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)
	reader.Comma = ','

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Create a dictionary
	dictionary := make(map[string]string)

	// Process records
	for i, record := range records {
		// Skip header if present
		if i == 0 && (len(record) > 0 && (record[0] == "word" || record[0] == "Word")) {
			continue
		}

		// Ensure record has at least 2 columns
		if len(record) < 2 {
			return nil, fmt.Errorf("invalid record at line %d: expected at least 2 columns", i+1)
		}

		word := record[0]
		definition := record[1]

		// Add to dictionary
		dictionary[word] = definition
	}

	return dictionary, nil
}

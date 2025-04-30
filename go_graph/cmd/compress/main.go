package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ekoch/explore_meaning/internal/dictionary"
	"github.com/ekoch/explore_meaning/internal/graph"
)

func main() {
	// Command line flags
	inputFile := flag.String("input", "", "Input dictionary CSV file path")
	outputFile := flag.String("output", "wordgraph.gz", "Output compressed graph file path")
	loadFile := flag.String("load", "", "Load compressed graph from file path")
	flag.Parse()

	// Either load an existing compressed graph or create a new one
	if *loadFile != "" {
		// Load existing compressed graph
		wg, err := graph.LoadCompressed(*loadFile)
		if err != nil {
			log.Fatalf("Failed to load compressed graph: %v", err)
		}

		// Print statistics
		stats := wg.GetStatistics()
		fmt.Printf("Loaded graph with %d lemmas and %d relationships\n",
			stats["TotalLemmas"], stats["TotalRelationships"])
	} else {
		// Create a new graph from dictionary
		if *inputFile == "" {
			log.Fatal("Input dictionary file is required when not loading from compressed graph")
		}

		// Load dictionary from CSV
		dict, err := dictionary.LoadFromCSV(*inputFile)
		if err != nil {
			log.Fatalf("Failed to load dictionary: %v", err)
		}

		fmt.Printf("Loaded dictionary with %d entries\n", len(dict))

		// Create and fill the word graph
		wg := graph.MakeNewWordGraph()
		wg.FillWith(dict)

		// Print original graph statistics
		stats := wg.GetStatistics()
		fmt.Printf("Created graph with %d lemmas and %d relationships\n",
			stats["TotalLemmas"], stats["TotalRelationships"])

		// Ensure output directory exists
		outputDir := filepath.Dir(*outputFile)
		if outputDir != "." && outputDir != "" {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("Failed to create output directory: %v", err)
			}
		}

		// Save compressed graph
		if err := wg.SaveCompressed(*outputFile); err != nil {
			log.Fatalf("Failed to save compressed graph: %v", err)
		}

		fmt.Printf("Compressed graph saved to %s\n", *outputFile)
	}
}

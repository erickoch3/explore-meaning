package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ekoch/explore_meaning/internal/analysis"
	"github.com/ekoch/explore_meaning/internal/graph"
	"github.com/ekoch/explore_meaning/internal/ingest"
)

func main() {
	// Define command-line flags
	inputFile := flag.String("input", "internal/data/dictionary_compact.json", "Input dictionary JSON file path")
	graphFile := flag.String("graph", "wordgraph.gz", "Path for saving/loading compressed graph")
	skipIngest := flag.Bool("skip-ingest", false, "Skip ingestion, use saved graph")
	skipAnalysis := flag.Bool("skip-analysis", false, "Skip analysis phase")
	saveGraph := flag.Bool("save", true, "Save graph after processing")

	flag.Parse()

	var wordGraph *graph.WordGraph

	// Check if we should load from a saved graph
	if *skipIngest {
		// Try to load existing compressed graph
		var err error
		wordGraph, err = graph.LoadCompressed(*graphFile)
		if err != nil {
			log.Fatalf("Failed to load compressed graph: %v", err)
		}

		// Print statistics
		stats := wordGraph.GetStatistics()
		fmt.Printf("Loaded graph with %d lemmas and %d relationships\n",
			stats["TotalLemmas"], stats["TotalRelationships"])
	} else {
		// Create a new graph
		wordGraph = graph.MakeNewWordGraph()

		// Load dictionary from JSON
		fmt.Printf("Loading dictionary from %s\n", *inputFile)
		rawDictionary := ingest.CreateDictionaryFromJSON(*inputFile)
		if rawDictionary == nil || len(rawDictionary) == 0 {
			log.Fatalf("Failed to load dictionary or dictionary is empty")
		}

		fmt.Printf("Loaded dictionary with %d entries\n", len(rawDictionary))

		// Fill the graph with dictionary data
		fmt.Println("Building word graph...")
		wordGraph.FillWith(rawDictionary)

		stats := wordGraph.GetStatistics()
		fmt.Printf("Created graph with %d lemmas and %d relationships\n",
			stats["TotalLemmas"], stats["TotalRelationships"])

		// Save the graph if requested
		if *saveGraph {
			// Ensure output directory exists
			outputDir := filepath.Dir(*graphFile)
			if outputDir != "." && outputDir != "" {
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					log.Fatalf("Failed to create output directory: %v", err)
				}
			}

			fmt.Printf("Saving graph to %s\n", *graphFile)
			if err := wordGraph.SaveCompressed(*graphFile); err != nil {
				log.Fatalf("Failed to save compressed graph: %v", err)
			}
		}
	}

	// Run analysis if not skipped
	if !*skipAnalysis {
		fmt.Println("Running analysis...")
		analysis.TrivialAnalysis(wordGraph)

		// If we've run analysis after loading and want to save the updated graph
		if *skipIngest && *saveGraph {
			fmt.Printf("Saving updated graph to %s\n", *graphFile)
			if err := wordGraph.SaveCompressed(*graphFile); err != nil {
				log.Fatalf("Failed to save compressed graph after analysis: %v", err)
			}
		}
	}
}

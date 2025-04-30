package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ekoch/explore_meaning/internal/graph"
	"github.com/ekoch/explore_meaning/internal/semantics"
)

// LemmaRatio represents a lemma and its ratio of defined-by to defines
type LemmaRatio struct {
	Lemma     semantics.Lemma
	DefinedBy int
	Defines   int
	Ratio     float64
	WordCount int
	Category  string // "producer", "consumer", or "balanced"
	IsExtreme bool   // Whether it's an extreme case (only defines/defined-by)
}

func main() {
	// Parse command line arguments
	graphFile := flag.String("graph", "./data/wordgraph.gz", "Path to the compressed graph file")
	limit := flag.Int("limit", 100, "Number of top producers/consumers to show")
	minWords := flag.Int("min-words", 5, "Minimum number of relationships required")
	flag.Parse()

	// Load the word graph
	wordGraph, err := graph.LoadCompressed(*graphFile)
	if err != nil {
		fmt.Printf("Error loading graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded graph with %d lemmas\n", len(wordGraph.Nodes))

	// Calculate ratios for all lemmas
	fmt.Println("Analyzing lemma relationships...")
	lemmaRatios := make([]LemmaRatio, 0, len(wordGraph.Nodes))

	for _, node := range wordGraph.Nodes {
		definedByCount := len(node.DefinedByLemmas)
		definesCount := len(node.DefinesLemmas)
		totalCount := definedByCount + definesCount

		// Skip lemmas with too few relationships
		if totalCount < *minWords {
			continue
		}

		// Calculate ratio and determine category
		ratio := 0.0
		category := "balanced"
		isExtreme := false

		if definesCount == 0 {
			ratio = float64(totalCount) // Max possible ratio
			category = "consumer"
			isExtreme = true
		} else if definedByCount == 0 {
			ratio = 0.0
			category = "producer"
			isExtreme = true
		} else {
			ratio = float64(definedByCount) / float64(definesCount)
			if ratio > 2.0 {
				category = "consumer"
			} else if ratio < 0.5 {
				category = "producer"
			}
		}

		lemmaRatios = append(lemmaRatios, LemmaRatio{
			Lemma:     node.Lemma,
			DefinedBy: definedByCount,
			Defines:   definesCount,
			Ratio:     ratio,
			WordCount: totalCount,
			Category:  category,
			IsExtreme: isExtreme,
		})
	}

	fmt.Printf("Found %d lemmas with at least %d relationships\n", len(lemmaRatios), *minWords)

	// Sort producers (first by ratio, then by number of words defined)
	sort.Slice(lemmaRatios, func(i, j int) bool {
		// First compare by ratio (lower is better for producers)
		if lemmaRatios[i].Ratio != lemmaRatios[j].Ratio {
			return lemmaRatios[i].Ratio < lemmaRatios[j].Ratio
		}
		// Within same ratio, sub-rank by number of words defined (higher is better)
		return lemmaRatios[i].Defines > lemmaRatios[j].Defines
	})

	// Display top producers
	fmt.Printf("\n=== Top %d Producers (ranked by in:out ratio, sub-ranked by words defined) ===\n", *limit)
	fmt.Println("These words have the lowest in:out ratios, with ties broken by number of words defined")
	printLemmaRatios(lemmaRatios, *limit, "producer")

	// Sort consumers (descending ratio - highest ratio first)
	sort.Slice(lemmaRatios, func(i, j int) bool {
		return lemmaRatios[i].Ratio > lemmaRatios[j].Ratio
	})

	// Display top consumers
	fmt.Printf("\n=== Top %d Consumers (highest in:out ratio) ===\n", *limit)
	fmt.Println("These words are frequently used in definitions but rarely define other words")
	printLemmaRatios(lemmaRatios, *limit, "consumer")

	// Display interesting stats
	extremeProducers := 0
	extremeConsumers := 0
	for _, lr := range lemmaRatios {
		if lr.IsExtreme {
			if lr.Category == "producer" {
				extremeProducers++
			} else if lr.Category == "consumer" {
				extremeConsumers++
			}
		}
	}

	fmt.Printf("\n=== Additional Statistics ===\n")
	fmt.Printf("Words that only define others (pure producers): %d\n", extremeProducers)
	fmt.Printf("Words that are only defined by others (pure consumers): %d\n", extremeConsumers)
}

// printLemmaRatios prints a list of lemma ratios
func printLemmaRatios(lemmaRatios []LemmaRatio, limit int, category string) {
	count := 0
	fmt.Printf("%-20s %-10s %-10s %-15s %-15s\n", "LEMMA", "DEFINED-BY", "DEFINES", "RATIO", "TOTAL RELS")
	fmt.Println(strings.Repeat("-", 75))

	for _, lr := range lemmaRatios {
		if count >= limit {
			break
		}

		if category == "" || lr.Category == category {
			ratioStr := fmt.Sprintf("%.2f:1", lr.Ratio)
			if lr.IsExtreme {
				if lr.Category == "producer" {
					ratioStr = "pure producer"
				} else if lr.Category == "consumer" {
					ratioStr = "pure consumer"
				}
			}

			fmt.Printf("%-20s %-10d %-10d %-15s %-15d\n",
				lr.Lemma, lr.DefinedBy, lr.Defines, ratioStr, lr.WordCount)
			count++
		}
	}
}

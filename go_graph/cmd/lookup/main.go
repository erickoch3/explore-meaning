package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ekoch/explore_meaning/internal/graph"
	"github.com/ekoch/explore_meaning/internal/semantics"
)

func main() {
	// Parse command line arguments
	graphFile := flag.String("graph", "./data/wordgraph.gz", "Path to the compressed graph file")
	interactive := flag.Bool("interactive", false, "Run in interactive mode")
	lemmaQuery := flag.String("lemma", "", "Lemma to look up")
	flag.Parse()

	// Load the word graph
	wordGraph, err := graph.LoadCompressed(*graphFile)
	if err != nil {
		fmt.Printf("Error loading graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded graph with %d lemmas\n", len(wordGraph.Nodes))

	// If a lemma was specified directly, look it up
	if *lemmaQuery != "" {
		lookupLemma(wordGraph, *lemmaQuery)
		return
	}

	// If interactive mode, prompt for lemmas until user exits
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("\nEnter a word to look up (or 'exit' to quit): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input == "exit" || input == "quit" {
				break
			}

			if input == "" {
				continue
			}

			lookupLemma(wordGraph, input)
		}
		return
	}

	// If neither interactive nor lemma query, print usage
	fmt.Println("Please use --lemma to specify a word or --interactive for interactive mode")
}

// lookupLemma finds and displays information about a lemma in the graph
func lookupLemma(wordGraph *graph.WordGraph, word string) {
	// Lemmatize the input
	lemma := semantics.Lemmatize(word)

	// Get the lemma node from the graph
	node := wordGraph.GetLemmaNode(lemma)

	// Check if this is a known lemma with relationships
	if node == nil || (len(node.DefinedByLemmas) == 0 && len(node.DefinesLemmas) == 0 && len(node.Definitions) == 0) {
		fmt.Printf("No data found for '%s' (lemma: '%s')\n", word, lemma)
		return
	}

	fmt.Printf("\n=== Information for '%s' (lemma: '%s') ===\n", word, lemma)

	// Show definitions with their analysis
	fmt.Printf("\nDefinitions (%d):\n", len(node.Definitions))
	for i, def := range node.Definitions {
		fmt.Printf("\n%d. %s\n", i+1, def.Text)

		// Count words used in this definition
		wordCount := len(def.UsesWords)

		if wordCount > 0 {
			fmt.Printf("  - Uses %d words\n", wordCount)

			// Calculate ratio for this specific definition
			// (total words defining this lemma / words used in this definition)
			ratio := float64(len(node.DefinedByLemmas)) / float64(wordCount)
			fmt.Printf("  - Definition ratio (in:out): %.2f:1\n", ratio)

			// Provide analysis for this definition
			if ratio > 2.0 {
				fmt.Println("  - This definition is primarily using common defining words (abstract)")
			} else if ratio < 0.5 {
				fmt.Println("  - This definition uses many specific words (concrete)")
			} else {
				fmt.Println("  - This definition has a balanced vocabulary")
			}
		}
	}

	// Show words that define this lemma
	fmt.Printf("\nDefined by %d lemmas:\n", len(node.DefinedByLemmas))
	for definedBy := range node.DefinedByLemmas {
		fmt.Printf("  - %s\n", definedBy)
	}

	// Show words that this lemma defines
	fmt.Printf("\nDefines %d lemmas:\n", len(node.DefinesLemmas))
	for defines := range node.DefinesLemmas {
		fmt.Printf("  - %s\n", defines)
	}

	// Display mathematical highlights and interesting metrics
	definedByCount := len(node.DefinedByLemmas)
	definesCount := len(node.DefinesLemmas)
	totalRelations := definedByCount + definesCount

	fmt.Printf("\n=== Overall Highlights ===\n")
	fmt.Printf("Total relationships: %d\n", totalRelations)

	if totalRelations > 0 {
		// Calculate and display the ratio of defined-by to defines
		ratio := float64(definedByCount) / float64(definesCount)
		if definesCount == 0 {
			fmt.Println("This word only appears in definitions, never defines other words")
		} else if definedByCount == 0 {
			fmt.Println("This word only defines other words, never appears in definitions")
		} else {
			fmt.Printf("Overall ratio (defined-by:defines): %.2f:1\n", ratio)

			if ratio > 2.0 {
				fmt.Println("This word is primarily used in definitions (consumer)")
			} else if ratio < 0.5 {
				fmt.Println("This word primarily defines others (producer)")
			} else {
				fmt.Println("This word has a balanced relationship in the lexicon")
			}
		}
	}
}

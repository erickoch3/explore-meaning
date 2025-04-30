// The purpose of this package is to create an in-memory graph structure that maps lemmas to each other based
// on their definitions and what other words they define.
package graph

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ekoch/explore_meaning/internal/semantics"
	"github.com/ekoch/explore_meaning/internal/storage"
)

type WordGraph struct {
	lemmaMap map[semantics.Lemma]*semantics.LemmaNode
	Nodes    []*semantics.LemmaNode
}

// Compiled regex for matching words (avoids recompiling the regex for each word)
var wordRegex = regexp.MustCompile(`[a-zA-Z]+`)

// MakeNewWordGraph creates a new empty word graph
func MakeNewWordGraph() *WordGraph {
	return &WordGraph{
		lemmaMap: make(map[semantics.Lemma]*semantics.LemmaNode),
		Nodes:    []*semantics.LemmaNode{},
	}
}

// cleanDefinition takes a definition string and returns a cleaned slice of words
// by removing punctuation and splitting properly
func cleanDefinition(definition string) []string {
	// Convert to lowercase
	definition = strings.ToLower(definition)

	// Extract all words
	words := wordRegex.FindAllString(definition, -1)

	// Filter out common stop words that don't add meaning
	filteredWords := make([]string, 0, len(words))
	for _, word := range words {
		// Skip very short words and common stop words
		if len(word) <= 1 || isStopWord(word) {
			continue
		}
		filteredWords = append(filteredWords, word)
	}

	return filteredWords
}

// isStopWord checks if a word is a common English stop word
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "and": true, "or": true,
		"of": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "with": true, "by": true, "as": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"being": true, "this": true, "that": true, "these": true, "those": true,
	}

	return stopWords[word]
}

// FillWith populates the word graph with data from a dictionary
func (wg *WordGraph) FillWith(dict map[string]string) {
	// Regular expression to match numbered definitions
	// Matches patterns like "1. ", "2. ", etc. at the beginning of a line or after a space
	numDefRegex := regexp.MustCompile(`(?:^|\s)(\d+\.)\s`)

	// Iterate over each entry in our dictionary
	for word, definitionText := range dict {
		// Create / acquire the word's lemma and corresponding node
		lemma := semantics.Lemmatize(word)
		lemmaNode := wg.GetLemmaNode(lemma)

		// Check if the definition contains numbered sections
		if numDefRegex.MatchString(definitionText) {
			// Find all instances of numbered definitions
			indices := numDefRegex.FindAllStringIndex(definitionText, -1)

			// If we found numbered definitions, split the text
			if len(indices) > 0 {
				// Extract individual definitions
				for i := 0; i < len(indices); i++ {
					startIdx := indices[i][0]
					// If there's whitespace before the number and it's not the first match, adjust the start index
					if startIdx > 0 && definitionText[startIdx] == ' ' {
						startIdx++
					}

					endIdx := len(definitionText)
					if i < len(indices)-1 {
						endIdx = indices[i+1][0]
					}

					// Get the individual definition
					singleDef := definitionText[startIdx:endIdx]

					// Process the individual definition
					definition := lemmaNode.AddDefinition(singleDef)
					processSingleDefinition(wg, lemma, lemmaNode, definition, singleDef)
				}
				continue
			}
		}

		// If no numbered sections or failed to parse, treat it as a single definition
		definition := lemmaNode.AddDefinition(definitionText)
		processSingleDefinition(wg, lemma, lemmaNode, definition, definitionText)
	}
}

// processSingleDefinition handles the processing of a single definition text
func processSingleDefinition(wg *WordGraph, lemma semantics.Lemma, lemmaNode *semantics.LemmaNode, definition *semantics.Definition, definitionText string) {
	// Process the definition words and add their lemmas to the node's definition
	for _, definitionWord := range cleanDefinition(definitionText) {
		definitionLemma := semantics.Lemmatize(definitionWord)

		// Skip empty lemmas or self-references
		if definitionLemma == "" || definitionLemma == lemma {
			continue
		}

		// Track which words are used in this specific definition
		definition.UsesWords[definitionLemma] = true

		// The definition lemma will be noted to define this lemma
		definitionLemmaNode := wg.GetLemmaNode(definitionLemma)
		definitionLemmaNode.Defines(lemma)

		// Add definition lemmas to this lemma's node
		lemmaNode.IsDefinedBy(definitionLemma)
	}
}

// Simple function to put a given lemma into the word graph by
// creating a corresponding node and adding the node to its map
func (wg *WordGraph) AddLemma(lemma semantics.Lemma) {
	if _, lemmaExists := wg.lemmaMap[lemma]; lemmaExists {
		return
	} else {
		lemmaNode := semantics.MakeLemmaNode(lemma)
		wg.lemmaMap[lemma] = lemmaNode
		wg.Nodes = append(wg.Nodes, lemmaNode)
	}
}

// Acquires the lemma node for a lemma from our graph, creating the node
// if it doesn't exist
func (wg *WordGraph) GetLemmaNode(lemma semantics.Lemma) *semantics.LemmaNode {
	if _, lemmaExists := wg.lemmaMap[lemma]; !lemmaExists {
		wg.AddLemma(lemma)
	}
	return wg.lemmaMap[lemma]
}

// SaveCompressed compresses the graph and saves it to a file
func (wg *WordGraph) SaveCompressed(filename string) error {
	// Compress the graph
	cg := storage.CompressGraph(wg.Nodes)

	// Print some stats
	stats := cg.Statistics()
	fmt.Printf("Compressing graph with %d lemmas and %d relationships\n",
		stats["TotalLemmas"], stats["TotalRelationships"])

	// Save to file
	if err := cg.SaveToFile(filename); err != nil {
		return fmt.Errorf("failed to save compressed graph: %w", err)
	}

	fmt.Printf("Graph successfully saved to %s\n", filename)
	return nil
}

// LoadCompressed loads a compressed graph from a file
func LoadCompressed(filename string) (*WordGraph, error) {
	// Load the compressed graph
	cg, err := storage.LoadFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load compressed graph: %w", err)
	}

	// Convert back to WordGraph
	nodes := cg.ConvertToLemmaNodes()

	// Create a new graph
	wg := MakeNewWordGraph()

	// Add all nodes to the graph
	for _, node := range nodes {
		wg.lemmaMap[node.Lemma] = node
		wg.Nodes = append(wg.Nodes, node)
	}

	// Print some stats
	stats := cg.Statistics()
	fmt.Printf("Loaded graph with %d lemmas and %d relationships\n",
		stats["TotalLemmas"], stats["TotalRelationships"])

	return wg, nil
}

// GetStatistics returns statistics about the graph
func (wg *WordGraph) GetStatistics() map[string]int {
	stats := make(map[string]int)

	stats["TotalLemmas"] = len(wg.Nodes)

	totalDefinitions := 0
	for _, node := range wg.Nodes {
		totalDefinitions += len(node.DefinedByLemmas)
	}

	stats["TotalRelationships"] = totalDefinitions

	return stats
}

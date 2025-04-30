package tests

import (
	"os"
	"testing"

	"github.com/ekoch/explore_meaning/internal/graph"
	"github.com/ekoch/explore_meaning/internal/ingest"
	"github.com/ekoch/explore_meaning/internal/semantics"
	"github.com/ekoch/explore_meaning/tests/testutil"
)

// TestWordGraphCreation tests the creation of a new, empty word graph
func TestWordGraphCreation(t *testing.T) {
	wg := graph.MakeNewWordGraph()

	if wg == nil {
		t.Fatal("Failed to create word graph")
	}

	if len(wg.Nodes) != 0 {
		t.Errorf("Expected empty nodes list, got %d nodes", len(wg.Nodes))
	}
}

// TestDictionaryFromJSON tests loading a dictionary from a JSON file
func TestDictionaryFromJSON(t *testing.T) {
	testFilePath := testutil.GetTestDictionaryPath()
	dict := ingest.CreateDictionaryFromJSON(testFilePath)

	if dict == nil {
		t.Fatal("Failed to load dictionary from JSON")
	}

	// Verify that the dictionary contains the expected entries
	expectedWords := []string{"red", "blue", "green", "color", "light", "eye"}
	for _, word := range expectedWords {
		if _, exists := dict[word]; !exists {
			t.Errorf("Expected dictionary to contain word '%s', but it was not found", word)
		}
	}

	// Verify that the definitions are correct
	if def, exists := dict["color"]; !exists || def != "a property possessed by an object producing different sensations on the eye as a result of the way it reflects or emits light" {
		t.Errorf("Definition for 'color' is incorrect or missing")
	}
}

// TestGraphFilling tests filling a graph with dictionary data
func TestGraphFilling(t *testing.T) {
	wg := graph.MakeNewWordGraph()
	testFilePath := testutil.GetTestDictionaryPath()
	dict := ingest.CreateDictionaryFromJSON(testFilePath)

	if dict == nil {
		t.Fatal("Failed to load dictionary from JSON")
	}

	// Fill the graph with the dictionary
	wg.FillWith(dict)

	// Check that nodes were created for all lemmas
	expectedLemmas := []string{"red", "blue", "green", "color", "light", "eye"}
	for _, lemmaStr := range expectedLemmas {
		lemma := semantics.Lemmatize(lemmaStr)
		found := false

		for _, node := range wg.Nodes {
			if node.Lemma == lemma {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected to find lemma '%s' in the graph, but it was not found", lemmaStr)
		}
	}

	// Check some specific relationships
	// 'color' should be defined by 'eye' and 'light'
	colorLemma := semantics.Lemmatize("color")
	eyeLemma := semantics.Lemmatize("eye")
	lightLemma := semantics.Lemmatize("light")

	colorNode := findNode(wg, colorLemma)
	if colorNode == nil {
		t.Fatal("Could not find 'color' node in the graph")
	}

	if !colorNode.DefinedByLemmas[eyeLemma] {
		t.Errorf("Expected 'color' to be defined by 'eye'")
	}

	if !colorNode.DefinedByLemmas[lightLemma] {
		t.Errorf("Expected 'color' to be defined by 'light'")
	}

	// 'red', 'blue', and 'green' should all be defined by 'color'
	redLemma := semantics.Lemmatize("red")
	blueLemma := semantics.Lemmatize("blue")
	greenLemma := semantics.Lemmatize("green")

	redNode := findNode(wg, redLemma)
	blueNode := findNode(wg, blueLemma)
	greenNode := findNode(wg, greenLemma)

	if redNode == nil || blueNode == nil || greenNode == nil {
		t.Fatal("Could not find all color nodes in the graph")
	}

	if !redNode.DefinedByLemmas[colorLemma] {
		t.Errorf("Expected 'red' to be defined by 'color'")
	}

	if !blueNode.DefinedByLemmas[colorLemma] {
		t.Errorf("Expected 'blue' to be defined by 'color'")
	}

	if !greenNode.DefinedByLemmas[colorLemma] {
		t.Errorf("Expected 'green' to be defined by 'color'")
	}
}

// Helper function to find a node in the graph by lemma
func findNode(wg *graph.WordGraph, lemma semantics.Lemma) *semantics.LemmaNode {
	for _, node := range wg.Nodes {
		if node.Lemma == lemma {
			return node
		}
	}
	return nil
}

func TestStopWordFiltering(t *testing.T) {
	// Create a graph with definitions containing stop words
	wg := graph.MakeNewWordGraph()

	// Dictionary with definitions containing stop words
	dictionary := map[string]string{
		"computer": "a machine that is used for storing and processing data",
		"internet": "a global network connecting millions of computers",
		"software": "the programs and other operating information used by a computer",
	}

	wg.FillWith(dictionary)

	// Get the nodes
	computerNode := findNode(wg, semantics.Lemmatize("computer"))

	if computerNode == nil {
		t.Fatal("Could not find computer node in the graph")
	}

	// Check that stop words aren't included in the definitions
	stopWords := []string{"a", "the", "is", "are", "for", "of", "by", "and", "that"}

	for _, stopWord := range stopWords {
		stopLemma := semantics.Lemmatize(stopWord)
		// Computer should not be defined by stop words
		if computerNode.DefinedByLemmas[stopLemma] {
			t.Errorf("Stop word '%s' was not filtered out of definitions", stopWord)
		}
	}

	// Check that meaningful words ARE included
	meaningfulWords := []string{"machine", "store", "process", "data"}

	for _, meaningfulWord := range meaningfulWords {
		meaningfulLemma := semantics.Lemmatize(meaningfulWord)
		// Computer should be defined by these meaningful words
		if !computerNode.DefinedByLemmas[meaningfulLemma] {
			t.Logf("Meaningful word '%s' (lemma: '%s') not found in definition",
				meaningfulWord, meaningfulLemma)
		}
	}
}

func TestMain(m *testing.M) {
	// Setup code
	exitCode := m.Run() // Run tests
	// Teardown code
	os.Exit(exitCode)
}

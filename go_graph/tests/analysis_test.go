package tests

import (
	"math"
	"testing"

	"github.com/ekoch/explore_meaning/internal/analysis"
	"github.com/ekoch/explore_meaning/internal/graph"
	"github.com/ekoch/explore_meaning/internal/ingest"
	"github.com/ekoch/explore_meaning/internal/semantics"
	"github.com/ekoch/explore_meaning/tests/testutil"
)

// TestCalculateInOutRatio tests the calculation of in-out ratios for lemma nodes
func TestCalculateInOutRatio(t *testing.T) {
	// Test empty node
	emptyNode := semantics.MakeLemmaNode("empty")
	ratio := analysis.CalculateInOutRatio(emptyNode)
	if ratio != 0 {
		t.Errorf("Expected ratio of 0 for empty node, got %f", ratio)
	}

	// Test node with only incoming links
	inOnlyNode := semantics.MakeLemmaNode("inOnly")
	inOnlyNode.IsDefinedBy("word1")
	inOnlyNode.IsDefinedBy("word2")
	ratio = analysis.CalculateInOutRatio(inOnlyNode)
	if ratio != math.MaxFloat32 {
		t.Errorf("Expected max ratio for node with only in connections, got %f", ratio)
	}

	// Test node with only outgoing links
	outOnlyNode := semantics.MakeLemmaNode("outOnly")
	outOnlyNode.Defines("word1")
	outOnlyNode.Defines("word2")
	ratio = analysis.CalculateInOutRatio(outOnlyNode)
	if ratio != 0 {
		t.Errorf("Expected ratio of 0 for node with only out connections, got %f", ratio)
	}

	// Test balanced node
	balancedNode := semantics.MakeLemmaNode("balanced")
	balancedNode.IsDefinedBy("inWord1")
	balancedNode.IsDefinedBy("inWord2")
	balancedNode.Defines("outWord1")
	balancedNode.Defines("outWord2")
	ratio = analysis.CalculateInOutRatio(balancedNode)
	if ratio != 1.0 {
		t.Errorf("Expected ratio of 1.0 for balanced node, got %f", ratio)
	}
}

// TestAnalysisOnSampleDictionary tests the analysis functions on a sample dictionary
func TestAnalysisOnSampleDictionary(t *testing.T) {
	wg := graph.MakeNewWordGraph()
	testFilePath := testutil.GetTestDictionaryPath()
	dict := ingest.CreateDictionaryFromJSON(testFilePath)

	if dict == nil {
		t.Fatal("Failed to load dictionary from JSON")
	}

	// Fill the graph with the dictionary
	wg.FillWith(dict)

	// Verify nodes exist
	if len(wg.Nodes) == 0 {
		t.Fatal("Graph has no nodes after filling")
	}

	// Test analysis calculations
	for _, node := range wg.Nodes {
		ratio := analysis.CalculateInOutRatio(node)

		// Ensure ratio is valid
		if ratio < 0 || (ratio > 0 && math.IsInf(float64(ratio), 1)) {
			t.Errorf("Invalid ratio %f for lemma %s", ratio, node.Lemma)
		}

		// Verify ratio is consistent with node data
		in := len(node.DefinedByLemmas)
		out := len(node.DefinesLemmas)

		if out == 0 {
			if in == 0 {
				if ratio != 0 {
					t.Errorf("Expected ratio of 0 for node with no connections, got %f", ratio)
				}
			} else {
				if ratio != math.MaxFloat32 {
					t.Errorf("Expected max ratio for node with only in connections, got %f", ratio)
				}
			}
		} else {
			expectedRatio := float32(in) / float32(out)
			if ratio != expectedRatio {
				t.Errorf("Expected ratio of %f for node, got %f", expectedRatio, ratio)
			}
		}
	}

	// This won't actually validate output, but will check that the function runs without error
	analysis.TrivialAnalysis(wg)
}

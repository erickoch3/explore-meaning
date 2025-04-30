// The purpose of this package is to analyze our word graph to glean insights into the meanings of the words
package analysis

import (
	"fmt"
	"math"
	"sort"

	"github.com/ekoch/explore_meaning/internal/graph"
	"github.com/ekoch/explore_meaning/internal/semantics"
)

// CalculateInOutRatio computes the ratio of "in" words to "out" words for a given node
// and handles division by zero cases
func CalculateInOutRatio(node *semantics.LemmaNode) float32 {
	in := len(node.DefinedByLemmas)
	out := len(node.DefinesLemmas)

	if out == 0 {
		if in == 0 {
			return 0 // Both in and out are zero
		}
		return math.MaxFloat32 // Only in has values, out is zero
	}

	return float32(in) / float32(out)
}

// TrivialAnalysis analyzes the word graph to find insights about the abstraction level of words
func TrivialAnalysis(wg *graph.WordGraph) {
	// This map will hold the ratios of "in" words (words that define the lemma) to
	// "out" words (words the lemma defines)
	// in naive theory, the words with high ratios are higher abstraction (edge of graph)
	// while words with low ratios are toward the center of the graph...
	inOutRatio := make(map[semantics.Lemma]float32)

	// Calculate ratio for each lemma node
	for _, lemmaNode := range wg.Nodes {
		inOutRatio[lemmaNode.Lemma] = CalculateInOutRatio(lemmaNode)
	}

	// Sort the map in ascending order of ratio
	sortedLemmas := make([]semantics.Lemma, len(wg.Nodes))
	for i, node := range wg.Nodes {
		sortedLemmas[i] = node.Lemma
	}

	sort.Slice(sortedLemmas, func(i, j int) bool {
		return inOutRatio[sortedLemmas[i]] < inOutRatio[sortedLemmas[j]]
	})

	// Print the top 10 lemmas by in-out-ratio
	limit := 10
	if len(sortedLemmas) < limit {
		limit = len(sortedLemmas)
	}

	for i, lemma := range sortedLemmas {
		if i >= limit {
			break
		}
		fmt.Printf("%v Lemma `%v` has in-out ratio of %v\n", i, lemma, inOutRatio[lemma])
	}
}

package semantics

import (
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
)

// A lemma is an irreducable English word construct
// For example, running and runs
// both reduce to the lemma "run"
type Lemma string

// Definition represents a single definition of a lemma
type Definition struct {
	Text      string         // The full text of the definition
	DefinesID string         // A unique identifier for this definition
	Lemma     Lemma          // The lemma this definition belongs to
	UsesWords map[Lemma]bool // Words used in this definition
}

// A lemma node is a unique node in our word graph
// that holds the associated lemma, words that the lemma defines,
// and words by which the lemma is defined.
type LemmaNode struct {
	Lemma           Lemma
	DefinesLemmas   map[Lemma]bool
	DefinedByLemmas map[Lemma]bool
	Definitions     []*Definition // Stores the definitions for this lemma
}

// A singleton instance of the lemmatizer
var lemmatizer *golem.Lemmatizer

// init initializes the lemmatizer once when the package is loaded
func init() {
	var err error
	lemmatizer, err = golem.New(en.New())
	if err != nil {
		panic(err)
	}
}

// Converts a string into a Lemma
func Lemmatize(word string) Lemma {
	if word == "" {
		return ""
	}
	return Lemma(lemmatizer.Lemma(word))
}

// Function to store another lemma in the node's DefinedBy map
func (ln *LemmaNode) IsDefinedBy(lemma Lemma) {
	ln.DefinedByLemmas[lemma] = true
}

// Function to store another lemma in the node's Defines map
func (ln *LemmaNode) Defines(lemma Lemma) {
	ln.DefinesLemmas[lemma] = true
}

// Function to add a definition to the lemma node
func (ln *LemmaNode) AddDefinition(text string) *Definition {
	// Create a unique ID for the definition using lemma name and position
	defID := string(ln.Lemma) + "_def_" + string(len(ln.Definitions)+1)

	def := &Definition{
		Text:      text,
		DefinesID: defID,
		Lemma:     ln.Lemma,
		UsesWords: make(map[Lemma]bool),
	}

	ln.Definitions = append(ln.Definitions, def)
	return def
}

// Function to build a new empty Lemma node for the graph
func MakeLemmaNode(lemma Lemma) *LemmaNode {
	return &LemmaNode{
		Lemma:           lemma,
		DefinesLemmas:   make(map[Lemma]bool),
		DefinedByLemmas: make(map[Lemma]bool),
		Definitions:     make([]*Definition, 0),
	}
}

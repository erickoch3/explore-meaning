// Package storage provides functionality for compressing and persisting the word graph
package storage

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"os"

	"github.com/ekoch/explore_meaning/internal/semantics"
)

// CompressedGraph represents a memory-efficient version of the word graph
type CompressedGraph struct {
	// Maps string lemma to its numeric ID
	LemmaToID map[string]uint32
	// Maps numeric ID back to string lemma
	IDToLemma map[uint32]string
	// Store relationships using numeric IDs instead of strings
	// For each lemma ID, store the IDs of lemmas that define it
	DefinedByMap map[uint32][]uint32
	// For each lemma ID, store the IDs of lemmas it defines
	DefinesMap map[uint32][]uint32
	// Store definitions for each lemma ID
	Definitions map[uint32][]Definition
	// Next available ID for new lemmas
	NextID uint32
}

// Definition structure for compressed storage
type Definition struct {
	Text      string   // The full text of the definition
	DefinesID string   // A unique identifier for this definition
	UsesWords []uint32 // IDs of words used in this definition
}

// NewCompressedGraph creates a new empty compressed graph
func NewCompressedGraph() *CompressedGraph {
	return &CompressedGraph{
		LemmaToID:    make(map[string]uint32),
		IDToLemma:    make(map[uint32]string),
		DefinedByMap: make(map[uint32][]uint32),
		DefinesMap:   make(map[uint32][]uint32),
		Definitions:  make(map[uint32][]Definition),
		NextID:       1, // Start from 1, reserve 0 as invalid
	}
}

// GetOrCreateLemmaID gets an existing ID for a lemma or creates a new one
func (cg *CompressedGraph) GetOrCreateLemmaID(lemma string) uint32 {
	if id, exists := cg.LemmaToID[lemma]; exists {
		return id
	}

	// Create new ID
	id := cg.NextID
	cg.NextID++

	// Store mappings
	cg.LemmaToID[lemma] = id
	cg.IDToLemma[id] = lemma

	// Initialize empty relationship arrays
	cg.DefinedByMap[id] = []uint32{}
	cg.DefinesMap[id] = []uint32{}
	cg.Definitions[id] = []Definition{}

	return id
}

// AddDefinition adds a relationship where definedByLemma defines targetLemma
func (cg *CompressedGraph) AddDefinition(targetLemma, definedByLemma string) {
	targetID := cg.GetOrCreateLemmaID(string(targetLemma))
	definerID := cg.GetOrCreateLemmaID(string(definedByLemma))

	// Add to defined-by relationship
	cg.DefinedByMap[targetID] = appendIfNotExists(cg.DefinedByMap[targetID], definerID)

	// Add to defines relationship
	cg.DefinesMap[definerID] = appendIfNotExists(cg.DefinesMap[definerID], targetID)
}

// AddLemmaDefinition adds a definition for a lemma
func (cg *CompressedGraph) AddLemmaDefinition(lemmaID uint32, def *semantics.Definition) {
	// Convert used words to IDs
	usesWordIDs := make([]uint32, 0, len(def.UsesWords))
	for word := range def.UsesWords {
		wordID := cg.GetOrCreateLemmaID(string(word))
		usesWordIDs = append(usesWordIDs, wordID)
	}

	// Create compressed definition
	compressedDef := Definition{
		Text:      def.Text,
		DefinesID: def.DefinesID,
		UsesWords: usesWordIDs,
	}

	// Add to definitions map
	cg.Definitions[lemmaID] = append(cg.Definitions[lemmaID], compressedDef)
}

// Helper function to append value to slice if it doesn't already exist
func appendIfNotExists(slice []uint32, value uint32) []uint32 {
	for _, item := range slice {
		if item == value {
			return slice // Already exists, don't add
		}
	}
	return append(slice, value)
}

// CompressGraph converts a normal word graph to a compressed graph
func CompressGraph(nodes []*semantics.LemmaNode) *CompressedGraph {
	cg := NewCompressedGraph()

	// First pass: create IDs for all lemmas
	for _, node := range nodes {
		cg.GetOrCreateLemmaID(string(node.Lemma))
	}

	// Second pass: add all relationships and definitions
	for _, node := range nodes {
		targetLemma := string(node.Lemma)
		targetID := cg.LemmaToID[targetLemma]

		// Add "defined by" relationships
		for definer := range node.DefinedByLemmas {
			cg.AddDefinition(targetLemma, string(definer))
		}

		// Add definitions
		for _, def := range node.Definitions {
			cg.AddLemmaDefinition(targetID, def)
		}
	}

	return cg
}

// ConvertToLemmaNodes converts the compressed graph back to lemma nodes
func (cg *CompressedGraph) ConvertToLemmaNodes() []*semantics.LemmaNode {
	nodes := make([]*semantics.LemmaNode, 0, len(cg.LemmaToID))

	for lemmaStr, id := range cg.LemmaToID {
		lemma := semantics.Lemma(lemmaStr)
		node := semantics.MakeLemmaNode(lemma)

		// Add "defined by" relationships
		for _, definerID := range cg.DefinedByMap[id] {
			definerLemma := semantics.Lemma(cg.IDToLemma[definerID])
			node.IsDefinedBy(definerLemma)
		}

		// Add "defines" relationships
		for _, definedID := range cg.DefinesMap[id] {
			definedLemma := semantics.Lemma(cg.IDToLemma[definedID])
			node.Defines(definedLemma)
		}

		// Add definitions
		if defs, ok := cg.Definitions[id]; ok {
			for _, compressedDef := range defs {
				def := node.AddDefinition(compressedDef.Text)
				def.DefinesID = compressedDef.DefinesID

				// Restore used words
				for _, wordID := range compressedDef.UsesWords {
					wordLemma := semantics.Lemma(cg.IDToLemma[wordID])
					def.UsesWords[wordLemma] = true
				}
			}
		}

		nodes = append(nodes, node)
	}

	return nodes
}

// SaveToFile saves the compressed graph to a file with gzip compression
func (cg *CompressedGraph) SaveToFile(filename string) error {
	// Create or open the file
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(f)
	defer gzipWriter.Close()

	// Create encoder
	encoder := gob.NewEncoder(gzipWriter)

	// Encode and write the data
	if err := encoder.Encode(cg); err != nil {
		return fmt.Errorf("failed to encode graph: %w", err)
	}

	return nil
}

// LoadFromFile loads a compressed graph from a file
func LoadFromFile(filename string) (*CompressedGraph, error) {
	// Open the file
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create decoder
	decoder := gob.NewDecoder(gzipReader)

	// Decode the data
	var cg CompressedGraph
	if err := decoder.Decode(&cg); err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("failed to decode graph: %w", err)
		}
	}

	return &cg, nil
}

// Statistics returns statistics about the compressed graph size
func (cg *CompressedGraph) Statistics() map[string]int {
	stats := make(map[string]int)

	stats["TotalLemmas"] = len(cg.LemmaToID)
	stats["TotalRelationships"] = 0

	for _, definers := range cg.DefinedByMap {
		stats["TotalRelationships"] += len(definers)
	}

	return stats
}

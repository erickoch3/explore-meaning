// The purpose of this package is to create an in-memory struct that will hold our dictionary after ingestion
package store

type WordGraph struct {
	Words []semantics.Word
}

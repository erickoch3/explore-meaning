# Output file for the compressed graph
GRAPH_FILE = ./data/wordgraph.gz
INPUT_FILE = ./internal/data/dictionary_compact.json

# Build the word map service
build:
	go build -o bin/word_map_service cmd/word_map_service/main.go

# Run the full pipeline (ingest and analyze)
run: build
	./bin/word_map_service --input $(INPUT_FILE) --graph $(GRAPH_FILE)

# Run only the ingest phase and save the graph
ingest: build
	./bin/word_map_service --input $(INPUT_FILE) --graph $(GRAPH_FILE) --skip-analysis

# Run only the analysis phase using a saved graph
analyze: build
	./bin/word_map_service --skip-ingest --graph $(GRAPH_FILE) --skip-analysis=false

# Build the web UI
webui:
	go build -o bin/webui cmd/webui/main.go

# Run the web UI
run-webui: webui
	./bin/webui --graph $(GRAPH_FILE)

# Define LEMMA with a default value that can be overridden from command line
# Usage: make lookup LEMMA=dog
LEMMA ?= love
# Run the lookup tool with a specific lemma
lookup:
	go run cmd/lookup/main.go --graph $(GRAPH_FILE) --lemma $(LEMMA)

# Run the lookup tool in interactive mode
ui:
	go run cmd/lookup/main.go --graph $(GRAPH_FILE) --interactive

# Build the analysis tool
analysis:
	go build -o bin/analysis cmd/analysis/main.go

# Run the analysis tool with configurable parameters
# Usage: make run-analysis LIMIT=50 MIN_WORDS=10
LIMIT ?= 100
MIN_WORDS ?= 5
run-analysis: analysis
	./bin/analysis --graph $(GRAPH_FILE) --limit $(LIMIT) --min-words $(MIN_WORDS)

# Run tests
test:
	go test -v ./tests/

# Clean up built files
clean:
	rm -f bin/word_map_service

# Create necessary directories
init:
	mkdir -p bin data

.PHONY: build run ingest analyze test clean init
# Explore Meaning

The purpose of this project is to map a dictionary to determine which root words are used to define other words.

The objective is to discover what core concepts are the building blocks for "meaning."

Humans learn language through first connecting sensory input signals to "nouns" within the world.

They then learn to qualify these nouns or relate them through adjectives and verbs. 

There should be some core set of building blocks that we use to define the world around us.

Ultimately, our goal should be to find the absolute language of the universe (to describe meaning as directly as possible). Human language is just a translation of this base level meaning to enable us to pass meaning from person to person.

Now in reality, language is a means of representing the world around us. However, you might imagine components of the world to be representations (or configurations) of some numenal essence described by a word as well. For instance, say you see a circular rock. You might describe it as a rock. You might also describe it as a circle. Your use of the word `Circle` relates the enity you describe to the numenal perfect `Circle` suggesting that it has similar characteristics, even though the rock is not a perfect circle. Thereby, your words are incapable of modeling perfect the world around you. You can only capture the relationship of objects in the world to numenal essences (core meaning) to convey and retranslate.

## Usage

The application supports incremental processing of the word graph, allowing you to save and load the state between runs:

### Commands

```bash
# Run the full pipeline (ingest and analyze)
make run

# Run only the ingest phase and save the graph
make ingest

# Run only the analysis phase using a saved graph
make analyze

# Build the word map service
make build

# Clean up built files
make clean

# Run tests
make test

# Create necessary directories
make init
```

### Command-line Options

When running the binary directly, you can use these options:

```bash
./bin/word_map_service [options]

Options:
  --input string    Input dictionary JSON file path (default "internal/data/dictionary_compact.json")
  --graph string    Path for saving/loading compressed graph (default "wordgraph.gz")
  --skip-ingest     Skip ingestion, use saved graph
  --skip-analysis   Skip analysis phase
  --save            Save graph after processing (default true)
```

## Reference

Dictionary from:
`https://github.com/matthewreagan/WebstersEnglishDictionary`
Saved locally to prevent loss in the future.
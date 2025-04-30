// This package serves to pull a dictionary from the internet or json file and put it in our store.
package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func CreateDictionaryFromJSON(filepath string) map[string]string {
	// Open the JSON file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Failed to read data from JSON:", err)
		return nil
	}
	defer file.Close()

	// Read the JSON content
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading the file:", err)
		return nil
	}

	// Construct the empty dictionary to hold the data
	dictionary := make(map[string]string)

	// Unmarshal the JSON data into the map
	err = json.Unmarshal(data, &dictionary)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	return dictionary
}

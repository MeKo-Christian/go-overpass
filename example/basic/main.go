package main

import (
	"log"

	"github.com/MeKo-Christian/go-overpass"
)

func main() {
	// Create a new client with default settings
	client := overpass.New()

	// Query for a specific node (Berlin TV Tower)
	query := `
		[out:json];
		node(125448864);
		out;
	`

	result, err := client.Query(query)
	if err != nil {
		log.Fatalf("Error querying Overpass API: %v", err)
	}

	// Print the results
	log.Printf("Found %d nodes\n", len(result.Nodes))

	for _, node := range result.Nodes {
		log.Printf("Node ID: %d\n", node.ID)
		log.Printf("  Location: %.6f, %.6f\n", node.Lat, node.Lon)

		if name, ok := node.Tags["name"]; ok {
			log.Printf("  Name: %s\n", name)
		}

		if node.Tags != nil {
			log.Printf("  Tags: %v\n", node.Tags)
		}
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MeKo-Christian/go-overpass"
)

func main() {
	// Create a custom HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create a custom Overpass client with:
	// - Custom HTTP client
	// - Custom endpoint (you can use alternative Overpass API servers)
	// - Custom max parallel requests (3 concurrent requests)
	client := overpass.NewWithSettings(
		httpClient,
		"https://overpass-api.de/api/interpreter",
		3, // maxParallel
	)

	// Query for cafes in a small area
	query := `
		[out:json];
		node["amenity"="cafe"](52.5,13.4,52.51,13.41);
		out;
	`

	result, err := client.Query(query)
	if err != nil {
		log.Fatalf("Error querying Overpass API: %v", err)
	}

	// Print the results
	fmt.Printf("Found %d cafes\n", len(result.Nodes))
	for _, node := range result.Nodes {
		name := node.Meta.Tags["name"]
		if name == "" {
			name = "Unnamed"
		}
		fmt.Printf("- %s (ID: %d) at %.6f, %.6f\n",
			name, node.Meta.ID, node.Lat, node.Lon)
	}
}

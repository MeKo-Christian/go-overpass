package main

import (
	"fmt"
	"log"

	"github.com/MeKo-Christian/go-overpass"
)

func main() {
	client := overpass.New()

	// Query for all restaurants in a bounding box (small area in Berlin)
	// Bounding box format: (south, west, north, east)
	query := `
		[out:json][bbox:52.5,13.4,52.51,13.41];
		(
		  node["amenity"="restaurant"];
		  way["amenity"="restaurant"];
		);
		out center;
	`

	result, err := client.Query(query)
	if err != nil {
		log.Fatalf("Error querying Overpass API: %v", err)
	}

	// Check if bounds were returned
	if result.Bounds != nil {
		fmt.Printf("Query bounds: %.6f,%.6f to %.6f,%.6f\n\n",
			result.Bounds.MinLat, result.Bounds.MinLon,
			result.Bounds.MaxLat, result.Bounds.MaxLon)
	}

	// Process restaurant nodes
	fmt.Printf("Found %d restaurant nodes\n", len(result.Nodes))
	for _, node := range result.Nodes {
		name := node.Meta.Tags["name"]
		if name == "" {
			name = "Unnamed restaurant"
		}
		cuisine := node.Meta.Tags["cuisine"]
		if cuisine == "" {
			cuisine = "unspecified"
		}

		fmt.Printf("- %s\n", name)
		fmt.Printf("  Cuisine: %s\n", cuisine)
		fmt.Printf("  Location: %.6f, %.6f\n", node.Lat, node.Lon)
		if phone, ok := node.Meta.Tags["phone"]; ok {
			fmt.Printf("  Phone: %s\n", phone)
		}
		fmt.Println()
	}

	// Process restaurant ways (buildings)
	fmt.Printf("Found %d restaurant ways\n", len(result.Ways))
	for _, way := range result.Ways {
		name := way.Meta.Tags["name"]
		if name == "" {
			name = "Unnamed restaurant"
		}
		fmt.Printf("- %s (Way ID: %d)\n", name, way.Meta.ID)

		// Ways might have center coordinates
		if way.Center != nil {
			fmt.Printf("  Center: %.6f, %.6f\n", way.Center.Lat, way.Center.Lon)
		}
		fmt.Println()
	}

	// Summary
	total := len(result.Nodes) + len(result.Ways) + len(result.Relations)
	fmt.Printf("Total elements found: %d\n", total)
}

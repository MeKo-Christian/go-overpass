package main

import (
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

	// Note: Result doesn't have bounds - individual elements may have bounds

	// Process restaurant nodes
	log.Printf("Found %d restaurant nodes\n", len(result.Nodes))

	for _, node := range result.Nodes {
		name := node.Tags["name"]
		if name == "" {
			name = "Unnamed restaurant"
		}

		cuisine := node.Tags["cuisine"]
		if cuisine == "" {
			cuisine = "unspecified"
		}

		log.Printf("- %s\n", name)
		log.Printf("  Cuisine: %s\n", cuisine)
		log.Printf("  Location: %.6f, %.6f\n", node.Lat, node.Lon)

		if phone, ok := node.Tags["phone"]; ok {
			log.Printf("  Phone: %s\n", phone)
		}

		log.Printf("")
	}

	// Process restaurant ways (buildings)
	log.Printf("Found %d restaurant ways\n", len(result.Ways))

	for _, way := range result.Ways {
		name := way.Tags["name"]
		if name == "" {
			name = "Unnamed restaurant"
		}

		log.Printf("- %s (Way ID: %d)\n", name, way.ID)

		// Ways might have geometry from 'out center' or 'out geom'
		if len(way.Geometry) > 0 {
			// Use first point as approximate center (or calculate actual center)
			log.Printf("  Geometry available (%d points)\n", len(way.Geometry))
		}

		log.Printf("")
	}

	// Summary
	total := len(result.Nodes) + len(result.Ways) + len(result.Relations)
	log.Printf("Total elements found: %d\n", total)
}

package main

import (
	"fmt"
	"log"

	"github.com/MeKo-Christian/go-overpass"
)

func main() {
	client := overpass.New()

	// Query for a way (street) with its nodes and geometry
	// This example queries for a specific street in Berlin
	query := `
		[out:json];
		way(4382581);
		out geom;
		out;
	`

	result, err := client.Query(query)
	if err != nil {
		log.Fatalf("Error querying Overpass API: %v", err)
	}

	// Print ways
	fmt.Printf("Found %d ways\n", len(result.Ways))
	for _, way := range result.Ways {
		name := way.Meta.Tags["name"]
		if name == "" {
			name = "Unnamed way"
		}
		fmt.Printf("\nWay: %s (ID: %d)\n", name, way.Meta.ID)
		fmt.Printf("  Number of nodes: %d\n", len(way.Nodes))

		// Print some tags
		if highway, ok := way.Meta.Tags["highway"]; ok {
			fmt.Printf("  Highway type: %s\n", highway)
		}

		// If geometry is available, print first few coordinates
		if way.Geometry != nil && len(way.Geometry) > 0 {
			fmt.Printf("  Geometry (first 3 points):\n")
			for i, point := range way.Geometry {
				if i >= 3 {
					break
				}
				fmt.Printf("    %.6f, %.6f\n", point.Lat, point.Lon)
			}
		}

		// Print referenced nodes
		if len(way.Nodes) > 0 {
			fmt.Printf("  Node IDs (first 5): ")
			for i, node := range way.Nodes {
				if i >= 5 {
					break
				}
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%d", node.Meta.ID)
			}
			fmt.Println()
		}
	}

	// Example query for a relation (e.g., a bus route)
	relationQuery := `
		[out:json];
		relation(1234567);
		out;
		>;
		out;
	`

	fmt.Println("\n--- Querying for relation ---")
	relResult, err := client.Query(relationQuery)
	if err != nil {
		// This might fail if the relation doesn't exist, which is fine for an example
		fmt.Printf("Note: Relation query failed (example ID may not exist): %v\n", err)
		return
	}

	fmt.Printf("Found %d relations\n", len(relResult.Relations))
	for _, relation := range relResult.Relations {
		name := relation.Meta.Tags["name"]
		if name == "" {
			name = "Unnamed relation"
		}
		fmt.Printf("\nRelation: %s (ID: %d)\n", name, relation.Meta.ID)
		fmt.Printf("  Type: %s\n", relation.Meta.Tags["type"])
		fmt.Printf("  Number of members: %d\n", len(relation.Members))

		// Print first few members
		if len(relation.Members) > 0 {
			fmt.Printf("  Members (first 3):\n")
			for i, member := range relation.Members {
				if i >= 3 {
					break
				}
				// Get the ID from the appropriate member type
				var memberID int64
				switch member.Type {
				case overpass.ElementTypeNode:
					if member.Node != nil {
						memberID = member.Node.Meta.ID
					}
				case overpass.ElementTypeWay:
					if member.Way != nil {
						memberID = member.Way.Meta.ID
					}
				case overpass.ElementTypeRelation:
					if member.Relation != nil {
						memberID = member.Relation.Meta.ID
					}
				}
				fmt.Printf("    %s (role: %s, id: %d)\n",
					member.Type, member.Role, memberID)
			}
		}
	}
}

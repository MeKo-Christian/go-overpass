package main

import (
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

	printWays(&result)

	// Example query for a relation (e.g., a bus route)
	relationQuery := `
		[out:json];
		relation(1234567);
		out;
		>;
		out;
	`

	log.Printf("\n--- Querying for relation ---")

	relResult, err := client.Query(relationQuery)
	if err != nil {
		// This might fail if the relation doesn't exist, which is fine for an example
		log.Printf("Note: Relation query failed (example ID may not exist): %v\n", err)
		return
	}

	printRelations(&relResult)
}

func printWays(result *overpass.Result) {
	log.Printf("Found %d ways\n", len(result.Ways))

	for _, way := range result.Ways {
		printWay(way)
	}
}

func printWay(way *overpass.Way) {
	name := way.Tags["name"]
	if name == "" {
		name = "Unnamed way"
	}

	log.Printf("\nWay: %s (ID: %d)\n", name, way.ID)
	log.Printf("  Number of nodes: %d\n", len(way.Nodes))

	// Print some tags
	if highway, ok := way.Tags["highway"]; ok {
		log.Printf("  Highway type: %s\n", highway)
	}

	// If geometry is available, print first few coordinates
	if len(way.Geometry) > 0 {
		log.Printf("  Geometry (first 3 points):\n")

		for i, point := range way.Geometry {
			if i >= 3 {
				break
			}

			log.Printf("    %.6f, %.6f\n", point.Lat, point.Lon)
		}
	}

	// Print referenced nodes
	if len(way.Nodes) > 0 {
		log.Printf("  Node IDs (first 5): ")

		for i, node := range way.Nodes {
			if i >= 5 {
				break
			}

			if i > 0 {
				log.Printf(", ")
			}

			log.Printf("%d", node.ID)
		}

		log.Printf("")
	}
}

func printRelations(result *overpass.Result) {
	log.Printf("Found %d relations\n", len(result.Relations))

	for _, relation := range result.Relations {
		name := relation.Tags["name"]
		if name == "" {
			name = "Unnamed relation"
		}

		log.Printf("\nRelation: %s (ID: %d)\n", name, relation.ID)
		log.Printf("  Type: %s\n", relation.Tags["type"])
		log.Printf("  Number of members: %d\n", len(relation.Members))

		// Print first few members
		if len(relation.Members) > 0 {
			log.Printf("  Members (first 3):\n")

			for i, member := range relation.Members {
				if i >= 3 {
					break
				}

				printMember(member)
			}
		}
	}
}

func printMember(member overpass.RelationMember) {
	// Get the ID from the appropriate member type
	var memberID int64

	switch member.Type {
	case overpass.ElementTypeNode:
		if member.Node != nil {
			memberID = member.Node.ID
		}
	case overpass.ElementTypeWay:
		if member.Way != nil {
			memberID = member.Way.ID
		}
	case overpass.ElementTypeRelation:
		if member.Relation != nil {
			memberID = member.Relation.ID
		}
	}

	log.Printf("    %s (role: %s, id: %d)\n",
		member.Type, member.Role, memberID)
}

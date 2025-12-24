package overpass

import (
	"fmt"
	"strings"
)

// QueryBuilder provides fluent API for building Overpass QL queries.
type QueryBuilder struct {
	elements   []string     // element type filters
	bbox       *BoundingBox // bounding box constraint
	filters    []TagFilter  // tag filters
	outputMode string       // output mode
	settings   []string     // query settings like [out:json]
}

// BoundingBox represents geographic bounds (south, west, north, east).
type BoundingBox struct {
	South, West, North, East float64
}

// TagFilter represents OSM tag filtering.
type TagFilter struct {
	Key      string
	Value    string
	Operator string // "=", "!=", "~", "exists"
}

// NewQueryBuilder creates new query builder with [out:json] default.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		elements:   []string{},
		filters:    []TagFilter{},
		outputMode: "out body",
		settings:   []string{"out:json"},
	}
}

// Node adds node element type to query.
func (qb *QueryBuilder) Node() *QueryBuilder {
	qb.elements = append(qb.elements, "node")
	return qb
}

// Way adds way element type to query.
func (qb *QueryBuilder) Way() *QueryBuilder {
	qb.elements = append(qb.elements, "way")
	return qb
}

// Relation adds relation element type to query.
func (qb *QueryBuilder) Relation() *QueryBuilder {
	qb.elements = append(qb.elements, "relation")
	return qb
}

// BBox sets bounding box constraint.
func (qb *QueryBuilder) BBox(south, west, north, east float64) *QueryBuilder {
	qb.bbox = &BoundingBox{
		South: south,
		West:  west,
		North: north,
		East:  east,
	}

	return qb
}

// Tag adds exact tag match filter.
func (qb *QueryBuilder) Tag(key, value string) *QueryBuilder {
	qb.filters = append(qb.filters, TagFilter{
		Key:      key,
		Value:    value,
		Operator: "=",
	})

	return qb
}

// TagExists adds filter for tag existence (any value).
func (qb *QueryBuilder) TagExists(key string) *QueryBuilder {
	qb.filters = append(qb.filters, TagFilter{
		Key:      key,
		Operator: "exists",
	})

	return qb
}

// TagNot adds negative tag match filter.
func (qb *QueryBuilder) TagNot(key, value string) *QueryBuilder {
	qb.filters = append(qb.filters, TagFilter{
		Key:      key,
		Value:    value,
		Operator: "!=",
	})

	return qb
}

// TagRegex adds regex tag value filter.
func (qb *QueryBuilder) TagRegex(key, pattern string) *QueryBuilder {
	qb.filters = append(qb.filters, TagFilter{
		Key:      key,
		Value:    pattern,
		Operator: "~",
	})

	return qb
}

// Output sets output mode (body, skel, ids, tags, meta, center, geom, bb).
func (qb *QueryBuilder) Output(mode string) *QueryBuilder {
	qb.outputMode = "out " + mode
	return qb
}

// OutputBody outputs all information (default).
func (qb *QueryBuilder) OutputBody() *QueryBuilder {
	qb.outputMode = "out body"
	return qb
}

// OutputGeom outputs with geometry (for ways/relations).
func (qb *QueryBuilder) OutputGeom() *QueryBuilder {
	qb.outputMode = "out geom"
	return qb
}

// OutputCenter outputs center point only.
func (qb *QueryBuilder) OutputCenter() *QueryBuilder {
	qb.outputMode = "out center"
	return qb
}

// OutputMeta outputs with metadata.
func (qb *QueryBuilder) OutputMeta() *QueryBuilder {
	qb.outputMode = "out meta"
	return qb
}

// Timeout sets query timeout in seconds.
func (qb *QueryBuilder) Timeout(seconds int) *QueryBuilder {
	// Remove existing timeout if any
	for i, s := range qb.settings {
		if strings.HasPrefix(s, "timeout:") {
			qb.settings = append(qb.settings[:i], qb.settings[i+1:]...)
			break
		}
	}

	qb.settings = append(qb.settings, fmt.Sprintf("timeout:%d", seconds))

	return qb
}

// Build constructs the Overpass QL query string.
func (qb *QueryBuilder) Build() string {
	parts := make([]string, 0, 10)

	// Settings
	if len(qb.settings) > 0 {
		parts = append(parts, "["+strings.Join(qb.settings, "][")+"]")
	}

	// If no element types specified, use all
	elements := qb.elements
	if len(elements) == 0 {
		elements = []string{"node", "way", "relation"}
	}

	// Union of element queries
	if len(elements) > 1 {
		parts = append(parts, "(")
	}

	filterSuffix := qb.buildFilterString()
	bboxSuffix := qb.buildBboxString()

	for i, elemType := range elements {
		if i > 0 {
			parts = append(parts, " ")
		}

		query := elemType + filterSuffix + bboxSuffix + ";"
		parts = append(parts, query)
	}

	if len(elements) > 1 {
		parts = append(parts, ");")
	}

	// Output
	parts = append(parts, qb.outputMode+";")

	return strings.Join(parts, "")
}

// String implements Stringer interface.
func (qb *QueryBuilder) String() string {
	return qb.Build()
}

// buildFilterString creates the filter suffix for an element query.
func (qb *QueryBuilder) buildFilterString() string {
	var filters string
	for _, filter := range qb.filters {
		switch filter.Operator {
		case "=":
			filters += fmt.Sprintf(`["%s"="%s"]`, filter.Key, filter.Value)
		case "!=":
			filters += fmt.Sprintf(`["%s"!="%s"]`, filter.Key, filter.Value)
		case "~":
			filters += fmt.Sprintf(`["%s"~"%s"]`, filter.Key, filter.Value)
		case "exists":
			filters += fmt.Sprintf(`["%s"]`, filter.Key)
		}
	}

	return filters
}

// buildBboxString creates the bounding box suffix if set.
func (qb *QueryBuilder) buildBboxString() string {
	if qb.bbox == nil {
		return ""
	}

	return fmt.Sprintf("(%.6f,%.6f,%.6f,%.6f)",
		qb.bbox.South, qb.bbox.West, qb.bbox.North, qb.bbox.East)
}

// Helper functions for common queries

// FindRestaurants creates query for restaurants in bounding box.
func FindRestaurants(south, west, north, east float64) *QueryBuilder {
	return NewQueryBuilder().
		Node().
		Way().
		BBox(south, west, north, east).
		Tag("amenity", "restaurant").
		OutputCenter()
}

// FindHighways creates query for highways in bounding box.
func FindHighways(south, west, north, east float64, highwayType string) *QueryBuilder {
	return NewQueryBuilder().
		Way().
		BBox(south, west, north, east).
		Tag("highway", highwayType).
		OutputGeom()
}

// FindAmenity creates query for amenity type in bounding box.
func FindAmenity(south, west, north, east float64, amenityType string) *QueryBuilder {
	return NewQueryBuilder().
		Node().
		Way().
		BBox(south, west, north, east).
		Tag("amenity", amenityType).
		OutputCenter()
}

// FindByTag creates query for elements with specific tag in bounding box.
func FindByTag(south, west, north, east float64, key, value string) *QueryBuilder {
	return NewQueryBuilder().
		Node().
		Way().
		Relation().
		BBox(south, west, north, east).
		Tag(key, value).
		OutputBody()
}

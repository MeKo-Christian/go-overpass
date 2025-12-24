package overpass

import (
	"strings"
	"testing"
)

func TestNewQueryBuilder(t *testing.T) {
	t.Parallel()

	queryBuilder := NewQueryBuilder()

	if queryBuilder == nil {
		t.Fatal("expected non-nil builder")
	}

	query := queryBuilder.Build()
	if !strings.Contains(query, "[out:json]") {
		t.Error("expected [out:json] in default query")
	}
}

func TestBuilderSingleNode(t *testing.T) {
	t.Parallel()

	query := NewQueryBuilder().
		Node().
		Tag("amenity", "restaurant").
		Build()

	expected := `[out:json]node["amenity"="restaurant"];out body;`
	if query != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, query)
	}
}

func TestBuilderBoundingBox(t *testing.T) {
	t.Parallel()

	query := NewQueryBuilder().
		Node().
		BBox(52.5, 13.4, 52.51, 13.41).
		Tag("amenity", "cafe").
		Build()

	if !strings.Contains(query, "(52.500000,13.400000,52.510000,13.410000)") {
		t.Errorf("bounding box not formatted correctly: %s", query)
	}
}

func TestBuilderMultipleElements(t *testing.T) {
	t.Parallel()

	query := NewQueryBuilder().
		Node().
		Way().
		Tag("amenity", "school").
		Build()

	if !strings.Contains(query, "(") {
		t.Error("expected union syntax for multiple elements")
	}

	if !strings.Contains(query, "node") || !strings.Contains(query, "way") {
		t.Error("missing element types")
	}
}

func TestBuilderTagFilters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		builder  *QueryBuilder
		expected string
	}{
		{
			"exact match",
			NewQueryBuilder().Node().Tag("highway", "primary"),
			`["highway"="primary"]`,
		},
		{
			"not equal",
			NewQueryBuilder().Node().TagNot("highway", "footway"),
			`["highway"!="footway"]`,
		},
		{
			"exists",
			NewQueryBuilder().Node().TagExists("name"),
			`["name"]`,
		},
		{
			"regex",
			NewQueryBuilder().Node().TagRegex("name", ".*Street"),
			`["name"~".*Street"]`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			query := testCase.builder.Build()
			if !strings.Contains(query, testCase.expected) {
				t.Errorf("expected %s in query:\n%s", testCase.expected, query)
			}
		})
	}
}

func TestBuilderMultipleFilters(t *testing.T) {
	t.Parallel()

	query := NewQueryBuilder().
		Node().
		Tag("amenity", "restaurant").
		Tag("cuisine", "italian").
		TagExists("phone").
		Build()

	if !strings.Contains(query, `["amenity"="restaurant"]`) {
		t.Error("missing amenity filter")
	}

	if !strings.Contains(query, `["cuisine"="italian"]`) {
		t.Error("missing cuisine filter")
	}

	if !strings.Contains(query, `["phone"]`) {
		t.Error("missing phone exists filter")
	}
}

func TestBuilderOutputModes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		builder  *QueryBuilder
		expected string
	}{
		{
			"body (default)",
			NewQueryBuilder().Node().OutputBody(),
			"out body;",
		},
		{
			"geom",
			NewQueryBuilder().Way().OutputGeom(),
			"out geom;",
		},
		{
			"center",
			NewQueryBuilder().Node().OutputCenter(),
			"out center;",
		},
		{
			"meta",
			NewQueryBuilder().Node().OutputMeta(),
			"out meta;",
		},
		{
			"custom",
			NewQueryBuilder().Node().Output("skel"),
			"out skel;",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			query := testCase.builder.Build()
			if !strings.Contains(query, testCase.expected) {
				t.Errorf("expected %s in query:\n%s", testCase.expected, query)
			}
		})
	}
}

func TestBuilderTimeout(t *testing.T) {
	t.Parallel()

	query := NewQueryBuilder().
		Node().
		Timeout(60).
		Build()

	if !strings.Contains(query, "[timeout:60]") {
		t.Errorf("expected timeout in query: %s", query)
	}
}

func TestBuilderTimeoutReplacement(t *testing.T) {
	t.Parallel()

	query := NewQueryBuilder().
		Node().
		Timeout(30).
		Timeout(60). // Should replace first timeout
		Build()

	if !strings.Contains(query, "[timeout:60]") {
		t.Errorf("expected timeout:60 in query: %s", query)
	}

	if strings.Contains(query, "[timeout:30]") {
		t.Errorf("old timeout should be replaced: %s", query)
	}
}

func TestBuilderChaining(t *testing.T) {
	t.Parallel()

	query := NewQueryBuilder().
		Node().
		Way().
		BBox(52.5, 13.4, 52.51, 13.41).
		Tag("amenity", "restaurant").
		Tag("cuisine", "italian").
		OutputCenter().
		Timeout(30).
		Build()

	// Verify all components present
	checks := []string{
		"[out:json]",
		"[timeout:30]",
		"node",
		"way",
		`["amenity"="restaurant"]`,
		`["cuisine"="italian"]`,
		"(52.500000,13.400000,52.510000,13.410000)",
		"out center;",
	}

	for _, check := range checks {
		if !strings.Contains(query, check) {
			t.Errorf("expected %s in query:\n%s", check, query)
		}
	}
}

func TestBuilderNoElements(t *testing.T) {
	t.Parallel()

	query := NewQueryBuilder().
		Tag("amenity", "restaurant").
		Build()

	// Should default to all element types
	if !strings.Contains(query, "node") {
		t.Error("expected node when no elements specified")
	}

	if !strings.Contains(query, "way") {
		t.Error("expected way when no elements specified")
	}

	if !strings.Contains(query, "relation") {
		t.Error("expected relation when no elements specified")
	}
}

func TestBuilderStringer(t *testing.T) {
	t.Parallel()

	qb := NewQueryBuilder().Node().Tag("name", "Berlin")

	str := qb.String()
	build := qb.Build()

	if str != build {
		t.Error("String() and Build() should return same result")
	}
}

func TestHelperFindRestaurants(t *testing.T) {
	t.Parallel()

	query := FindRestaurants(52.5, 13.4, 52.51, 13.41).Build()

	if !strings.Contains(query, `["amenity"="restaurant"]`) {
		t.Error("missing restaurant filter")
	}

	if !strings.Contains(query, "out center;") {
		t.Error("missing center output")
	}

	if !strings.Contains(query, "node") {
		t.Error("missing node element type")
	}

	if !strings.Contains(query, "way") {
		t.Error("missing way element type")
	}
}

func TestHelperFindHighways(t *testing.T) {
	t.Parallel()

	query := FindHighways(52.5, 13.4, 52.51, 13.41, "primary").Build()

	if !strings.Contains(query, `["highway"="primary"]`) {
		t.Error("missing highway filter")
	}

	if !strings.Contains(query, "way") {
		t.Error("missing way element type")
	}

	if !strings.Contains(query, "out geom;") {
		t.Error("missing geom output")
	}
}

func TestHelperFindAmenity(t *testing.T) {
	t.Parallel()

	query := FindAmenity(52.5, 13.4, 52.51, 13.41, "cafe").Build()

	if !strings.Contains(query, `["amenity"="cafe"]`) {
		t.Error("missing amenity filter")
	}

	if !strings.Contains(query, "node") {
		t.Error("missing node element type")
	}

	if !strings.Contains(query, "way") {
		t.Error("missing way element type")
	}
}

func TestHelperFindByTag(t *testing.T) {
	t.Parallel()

	query := FindByTag(52.5, 13.4, 52.51, 13.41, "leisure", "park").Build()

	if !strings.Contains(query, `["leisure"="park"]`) {
		t.Error("missing tag filter")
	}

	if !strings.Contains(query, "node") || !strings.Contains(query, "way") {
		t.Error("missing element types")
	}

	if !strings.Contains(query, "relation") {
		t.Error("missing relation element type")
	}
}

func TestBuilderComplexQuery(t *testing.T) {
	t.Parallel()

	// Test realistic complex query
	query := NewQueryBuilder().
		Node().
		Way().
		BBox(52.5, 13.4, 52.51, 13.41).
		Tag("amenity", "restaurant").
		TagNot("diet:vegan", "only").
		TagExists("wheelchair").
		TagRegex("cuisine", "italian|pizza").
		OutputCenter().
		Timeout(45).
		Build()

	// Just verify it builds without error and contains key parts
	if query == "" {
		t.Error("builder produced empty query")
	}

	// Verify it's valid-ish OverpassQL
	if !strings.HasPrefix(query, "[") {
		t.Error("query should start with settings")
	}

	if !strings.HasSuffix(query, ";") {
		t.Error("query should end with semicolon")
	}

	// Check specific filters are present
	expectedParts := []string{
		`["amenity"="restaurant"]`,
		`["diet:vegan"!="only"]`,
		`["wheelchair"]`,
		`["cuisine"~"italian|pizza"]`,
	}

	for _, part := range expectedParts {
		if !strings.Contains(query, part) {
			t.Errorf("missing expected part %s in query:\n%s", part, query)
		}
	}
}

func TestBuilderSingleElement(t *testing.T) {
	t.Parallel()

	// Single element should not have union syntax
	query := NewQueryBuilder().
		Node().
		Tag("amenity", "restaurant").
		Build()

	// Should NOT have "(" for union when only one element type
	if strings.Contains(query, "(") && strings.Contains(query, ");") {
		t.Error("single element should not use union syntax")
	}
}

func TestBuilderThreeElements(t *testing.T) {
	t.Parallel()

	query := NewQueryBuilder().
		Node().
		Way().
		Relation().
		Tag("name", "Test").
		Build()

	// Should have all three elements
	if !strings.Contains(query, "node") {
		t.Error("missing node")
	}

	if !strings.Contains(query, "way") {
		t.Error("missing way")
	}

	if !strings.Contains(query, "relation") {
		t.Error("missing relation")
	}

	// Should use union syntax
	if !strings.Contains(query, "(") || !strings.Contains(query, ");") {
		t.Error("three elements should use union syntax")
	}
}

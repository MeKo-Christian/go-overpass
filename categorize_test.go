package overpass

import (
	"testing"
)

func TestGetCategory(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tags     map[string]string
		expected Category
	}{
		{
			"highway",
			map[string]string{"highway": "primary"},
			CategoryTransportation,
		},
		{
			"railway",
			map[string]string{"railway": "station"},
			CategoryTransportation,
		},
		{
			"aeroway",
			map[string]string{"aeroway": "aerodrome"},
			CategoryTransportation,
		},
		{
			"amenity",
			map[string]string{"amenity": "restaurant"},
			CategoryAmenity,
		},
		{
			"natural",
			map[string]string{"natural": "tree"},
			CategoryNatural,
		},
		{
			"waterway",
			map[string]string{"waterway": "river"},
			CategoryWater,
		},
		{
			"building",
			map[string]string{"building": "yes"},
			CategoryBuilding,
		},
		{
			"leisure",
			map[string]string{"leisure": "park"},
			CategoryLeisure,
		},
		{
			"landuse",
			map[string]string{"landuse": "forest"},
			CategoryLanduse,
		},
		{
			"boundary",
			map[string]string{"boundary": "administrative"},
			CategoryBoundary,
		},
		{
			"place",
			map[string]string{"place": "city"},
			CategoryPlace,
		},
		{
			"shop",
			map[string]string{"shop": "supermarket"},
			CategoryShop,
		},
		{
			"tourism",
			map[string]string{"tourism": "hotel"},
			CategoryTourism,
		},
		{
			"unknown",
			map[string]string{"foo": "bar"},
			CategoryUnknown,
		},
		{
			"priority: highway over building",
			map[string]string{"highway": "residential", "building": "yes"},
			CategoryTransportation,
		},
		{
			"priority: highway over amenity",
			map[string]string{"highway": "service", "amenity": "parking"},
			CategoryTransportation,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: testCase.tags}

			got := meta.GetCategory()
			if got != testCase.expected {
				t.Errorf("expected %s, got %s", testCase.expected, got)
			}
		})
	}
}

func TestGetSubcategory(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tags     map[string]string
		expected string
	}{
		{
			"highway primary",
			map[string]string{"highway": "primary"},
			"primary",
		},
		{
			"amenity restaurant",
			map[string]string{"amenity": "restaurant"},
			"restaurant",
		},
		{
			"natural tree",
			map[string]string{"natural": "tree"},
			"tree",
		},
		{
			"waterway river",
			map[string]string{"waterway": "river"},
			"river",
		},
		{
			"building yes",
			map[string]string{"building": "yes"},
			"yes",
		},
		{
			"leisure park",
			map[string]string{"leisure": "park"},
			"park",
		},
		{
			"landuse forest",
			map[string]string{"landuse": "forest"},
			"forest",
		},
		{
			"boundary administrative",
			map[string]string{"boundary": "administrative"},
			"administrative",
		},
		{
			"place city",
			map[string]string{"place": "city"},
			"city",
		},
		{
			"shop supermarket",
			map[string]string{"shop": "supermarket"},
			"supermarket",
		},
		{
			"tourism hotel",
			map[string]string{"tourism": "hotel"},
			"hotel",
		},
		{
			"unknown - empty",
			map[string]string{"foo": "bar"},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: tc.tags}

			got := meta.GetSubcategory()
			if got != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, got)
			}
		})
	}
}

func TestCategoryHelpers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		tags   map[string]string
		method func(*Meta) bool
		expect bool
	}{
		{
			"IsTransportation - highway",
			map[string]string{"highway": "primary"},
			(*Meta).IsTransportation,
			true,
		},
		{
			"IsTransportation - not",
			map[string]string{"amenity": "restaurant"},
			(*Meta).IsTransportation,
			false,
		},
		{
			"IsAmenity - restaurant",
			map[string]string{"amenity": "restaurant"},
			(*Meta).IsAmenity,
			true,
		},
		{
			"IsAmenity - not",
			map[string]string{"highway": "primary"},
			(*Meta).IsAmenity,
			false,
		},
		{
			"IsNatural - tree",
			map[string]string{"natural": "tree"},
			(*Meta).IsNatural,
			true,
		},
		{
			"IsNatural - not",
			map[string]string{"amenity": "restaurant"},
			(*Meta).IsNatural,
			false,
		},
		{
			"IsWater - river",
			map[string]string{"waterway": "river"},
			(*Meta).IsWater,
			true,
		},
		{
			"IsWater - not",
			map[string]string{"highway": "primary"},
			(*Meta).IsWater,
			false,
		},
		{
			"IsBuilding - yes",
			map[string]string{"building": "yes"},
			(*Meta).IsBuilding,
			true,
		},
		{
			"IsBuilding - not",
			map[string]string{"amenity": "restaurant"},
			(*Meta).IsBuilding,
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: tc.tags}

			got := tc.method(&meta)
			if got != tc.expect {
				t.Errorf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestGetName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tags     map[string]string
		expected string
	}{
		{
			"has name",
			map[string]string{"name": "Berlin"},
			"Berlin",
		},
		{
			"no name",
			map[string]string{"amenity": "restaurant"},
			"",
		},
		{
			"empty name",
			map[string]string{"name": ""},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: tc.tags}

			got := meta.GetName()
			if got != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, got)
			}
		})
	}
}

func TestHasTag(t *testing.T) {
	t.Parallel()

	meta := Meta{
		Tags: map[string]string{
			"amenity": "restaurant",
			"name":    "Test",
		},
	}

	if !meta.HasTag("amenity") {
		t.Error("expected amenity tag to exist")
	}

	if !meta.HasTag("name") {
		t.Error("expected name tag to exist")
	}

	if meta.HasTag("missing") {
		t.Error("expected missing tag to not exist")
	}
}

func TestGetTag(t *testing.T) {
	t.Parallel()

	meta := Meta{
		Tags: map[string]string{
			"amenity": "restaurant",
		},
	}

	if got := meta.GetTag("amenity", "default"); got != "restaurant" {
		t.Errorf("expected restaurant, got %s", got)
	}

	if got := meta.GetTag("missing", "default"); got != "default" {
		t.Errorf("expected default, got %s", got)
	}

	if got := meta.GetTag("missing", ""); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

func TestMatchesFilter(t *testing.T) {
	t.Parallel()

	meta := Meta{
		Tags: map[string]string{
			"amenity": "restaurant",
			"cuisine": "italian",
		},
	}

	if !meta.MatchesFilter("amenity", "restaurant") {
		t.Error("expected match on amenity=restaurant")
	}

	if meta.MatchesFilter("amenity", "cafe") {
		t.Error("expected no match on amenity=cafe")
	}

	if meta.MatchesFilter("missing", "value") {
		t.Error("expected no match on missing tag")
	}

	if !meta.MatchesFilter("cuisine", "italian") {
		t.Error("expected match on cuisine=italian")
	}
}

func TestIsRoad(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tags     map[string]string
		expected bool
	}{
		{
			"primary highway",
			map[string]string{"highway": "primary"},
			true,
		},
		{
			"residential highway",
			map[string]string{"highway": "residential"},
			true,
		},
		{
			"footway",
			map[string]string{"highway": "footway"},
			true,
		},
		{
			"not road",
			map[string]string{"railway": "station"},
			false,
		},
		{
			"not road - amenity",
			map[string]string{"amenity": "parking"},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: tc.tags}

			got := meta.IsRoad()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestIsRailway(t *testing.T) {
	t.Parallel()

	meta1 := Meta{Tags: map[string]string{"railway": "station"}}
	if !meta1.IsRailway() {
		t.Error("expected railway")
	}

	meta2 := Meta{Tags: map[string]string{"highway": "primary"}}
	if meta2.IsRailway() {
		t.Error("expected not railway")
	}

	meta3 := Meta{Tags: map[string]string{"railway": "tram"}}
	if !meta3.IsRailway() {
		t.Error("expected railway")
	}
}

func TestIsFoodRelated(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tags     map[string]string
		expected bool
	}{
		{"restaurant", map[string]string{"amenity": "restaurant"}, true},
		{"cafe", map[string]string{"amenity": "cafe"}, true},
		{"fast_food", map[string]string{"amenity": "fast_food"}, true},
		{"bar", map[string]string{"amenity": "bar"}, true},
		{"pub", map[string]string{"amenity": "pub"}, true},
		{"food_court", map[string]string{"amenity": "food_court"}, true},
		{"biergarten", map[string]string{"amenity": "biergarten"}, true},
		{"school", map[string]string{"amenity": "school"}, false},
		{"no amenity", map[string]string{"building": "yes"}, false},
		{"hospital", map[string]string{"amenity": "hospital"}, false},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: tc.tags}

			got := meta.IsFoodRelated()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func testBoolMethod(t *testing.T, testCases []struct {
	name     string
	tags     map[string]string
	expected bool
}, checkFunc func(Meta) bool,
) {
	t.Helper()

	for _, testCase := range testCases {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: testCase.tags}

			got := checkFunc(meta)
			if got != testCase.expected {
				t.Errorf("expected %v, got %v", testCase.expected, got)
			}
		})
	}
}

func TestIsEducation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tags     map[string]string
		expected bool
	}{
		{"school", map[string]string{"amenity": "school"}, true},
		{"university", map[string]string{"amenity": "university"}, true},
		{"college", map[string]string{"amenity": "college"}, true},
		{"library", map[string]string{"amenity": "library"}, true},
		{"kindergarten", map[string]string{"amenity": "kindergarten"}, true},
		{"restaurant", map[string]string{"amenity": "restaurant"}, false},
		{"hospital", map[string]string{"amenity": "hospital"}, false},
		{"no amenity", map[string]string{"building": "yes"}, false},
	}

	testBoolMethod(t, testCases, func(m Meta) bool { return m.IsEducation() })
}

func TestIsHealthcare(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tags     map[string]string
		expected bool
	}{
		{"hospital", map[string]string{"amenity": "hospital"}, true},
		{"clinic", map[string]string{"amenity": "clinic"}, true},
		{"doctors", map[string]string{"amenity": "doctors"}, true},
		{"dentist", map[string]string{"amenity": "dentist"}, true},
		{"pharmacy", map[string]string{"amenity": "pharmacy"}, true},
		{"restaurant", map[string]string{"amenity": "restaurant"}, false},
		{"school", map[string]string{"amenity": "school"}, false},
		{"no amenity", map[string]string{"building": "yes"}, false},
	}

	testBoolMethod(t, testCases, func(m Meta) bool { return m.IsHealthcare() })
}

func TestCategoryPriority(t *testing.T) {
	t.Parallel()

	// Element with multiple category tags - highway should win
	meta := Meta{
		Tags: map[string]string{
			"highway":  "residential",
			"building": "yes",
			"amenity":  "parking",
		},
	}

	category := meta.GetCategory()
	if category != CategoryTransportation {
		t.Errorf("expected transportation category due to priority, got %s", category)
	}

	subcategory := meta.GetSubcategory()
	if subcategory != "residential" {
		t.Errorf("expected residential subcategory, got %s", subcategory)
	}
}

func TestEmptyTags(t *testing.T) {
	t.Parallel()

	meta := Meta{Tags: map[string]string{}}

	if meta.GetCategory() != CategoryUnknown {
		t.Error("expected unknown category for empty tags")
	}

	if meta.GetSubcategory() != "" {
		t.Error("expected empty subcategory for empty tags")
	}

	if meta.GetName() != "" {
		t.Error("expected empty name for empty tags")
	}

	if meta.HasTag("anything") {
		t.Error("expected no tags")
	}
}

func TestNilTags(t *testing.T) {
	t.Parallel()

	meta := Meta{Tags: nil}

	// Should not panic
	category := meta.GetCategory()
	if category != CategoryUnknown {
		t.Errorf("expected unknown category for nil tags, got %s", category)
	}

	subcategory := meta.GetSubcategory()
	if subcategory != "" {
		t.Errorf("expected empty subcategory for nil tags, got %s", subcategory)
	}

	if meta.GetName() != "" {
		t.Error("expected empty name for nil tags")
	}

	if meta.HasTag("anything") {
		t.Error("expected no tags for nil Tags map")
	}
}

func TestMultipleTransportationTypes(t *testing.T) {
	t.Parallel()

	// Should prioritize highway over railway
	meta := Meta{
		Tags: map[string]string{
			"highway": "primary",
			"railway": "tram",
		},
	}

	if !meta.IsTransportation() {
		t.Error("expected transportation")
	}

	// Highway has priority
	subcategory := meta.GetSubcategory()
	if subcategory != "primary" {
		t.Errorf("expected highway subcategory to have priority, got %s", subcategory)
	}
}

func TestRealWorldExample(t *testing.T) {
	t.Parallel()

	// Restaurant with full details
	restaurant := Meta{
		Tags: map[string]string{
			"amenity":    "restaurant",
			"name":       "La Bella Vita",
			"cuisine":    "italian",
			"wheelchair": "yes",
		},
	}

	if !restaurant.IsAmenity() {
		t.Error("expected amenity")
	}

	if !restaurant.IsFoodRelated() {
		t.Error("expected food related")
	}

	if restaurant.GetName() != "La Bella Vita" {
		t.Errorf("expected name La Bella Vita, got %s", restaurant.GetName())
	}

	if !restaurant.HasTag("wheelchair") {
		t.Error("expected wheelchair tag")
	}

	// Road with name
	road := Meta{
		Tags: map[string]string{
			"highway": "primary",
			"name":    "Main Street",
			"lanes":   "2",
		},
	}

	if !road.IsTransportation() {
		t.Error("expected transportation")
	}

	if !road.IsRoad() {
		t.Error("expected road")
	}

	if road.GetSubcategory() != "primary" {
		t.Errorf("expected primary subcategory, got %s", road.GetSubcategory())
	}

	if road.GetName() != "Main Street" {
		t.Errorf("expected Main Street, got %s", road.GetName())
	}
}

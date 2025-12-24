package overpass

import (
	"testing"
)

func testGetCategoryHelper(t *testing.T, tags map[string]string, expected Category) {
	t.Helper()

	meta := Meta{Tags: tags}

	got := meta.GetCategory()
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestGetCategory(t *testing.T) { //nolint:funlen // many test cases for comprehensive coverage
	t.Parallel()

	t.Run("highway", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"highway": "primary"}, CategoryTransportation)
	})

	t.Run("railway", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"railway": "station"}, CategoryTransportation)
	})

	t.Run("aeroway", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"aeroway": "aerodrome"}, CategoryTransportation)
	})

	t.Run("amenity", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"amenity": "restaurant"}, CategoryAmenity)
	})

	t.Run("natural", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"natural": "tree"}, CategoryNatural)
	})

	t.Run("waterway", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"waterway": "river"}, CategoryWater)
	})

	t.Run("building", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"building": "yes"}, CategoryBuilding)
	})

	t.Run("leisure", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"leisure": "park"}, CategoryLeisure)
	})

	t.Run("landuse", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"landuse": "forest"}, CategoryLanduse)
	})

	t.Run("boundary", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"boundary": "administrative"}, CategoryBoundary)
	})

	t.Run("place", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"place": "city"}, CategoryPlace)
	})

	t.Run("shop", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"shop": "supermarket"}, CategoryShop)
	})

	t.Run("tourism", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"tourism": "hotel"}, CategoryTourism)
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"foo": "bar"}, CategoryUnknown)
	})

	t.Run("priority: highway over building", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"highway": "residential", "building": "yes"}, CategoryTransportation)
	})

	t.Run("priority: highway over amenity", func(t *testing.T) {
		t.Parallel()
		testGetCategoryHelper(t, map[string]string{"highway": "service", "amenity": "parking"}, CategoryTransportation)
	})
}

func testGetSubcategoryHelper(t *testing.T, tags map[string]string, expected string) {
	t.Helper()

	meta := Meta{Tags: tags}

	got := meta.GetSubcategory()
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestGetSubcategory(t *testing.T) { //nolint:funlen // many test cases for comprehensive coverage
	t.Parallel()

	t.Run("highway primary", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"highway": "primary"}, "primary")
	})

	t.Run("amenity restaurant", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"amenity": "restaurant"}, "restaurant")
	})

	t.Run("natural tree", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"natural": "tree"}, "tree")
	})

	t.Run("waterway river", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"waterway": "river"}, "river")
	})

	t.Run("building yes", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"building": "yes"}, "yes")
	})

	t.Run("leisure park", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"leisure": "park"}, "park")
	})

	t.Run("landuse forest", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"landuse": "forest"}, "forest")
	})

	t.Run("boundary administrative", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"boundary": "administrative"}, "administrative")
	})

	t.Run("place city", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"place": "city"}, "city")
	})

	t.Run("shop supermarket", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"shop": "supermarket"}, "supermarket")
	})

	t.Run("tourism hotel", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"tourism": "hotel"}, "hotel")
	})

	t.Run("unknown - empty", func(t *testing.T) {
		t.Parallel()
		testGetSubcategoryHelper(t, map[string]string{"foo": "bar"}, "")
	})
}

func testCategoryHelperMethod(t *testing.T, tags map[string]string, method func(*Meta) bool, expect bool) {
	t.Helper()

	meta := Meta{Tags: tags}

	got := method(&meta)
	if got != expect {
		t.Errorf("expected %v, got %v", expect, got)
	}
}

func TestCategoryHelpers(t *testing.T) {
	t.Parallel()

	t.Run("IsTransportation - highway", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"highway": "primary"}, (*Meta).IsTransportation, true)
	})

	t.Run("IsTransportation - not", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"amenity": "restaurant"}, (*Meta).IsTransportation, false)
	})

	t.Run("IsAmenity - restaurant", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"amenity": "restaurant"}, (*Meta).IsAmenity, true)
	})

	t.Run("IsAmenity - not", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"highway": "primary"}, (*Meta).IsAmenity, false)
	})

	t.Run("IsNatural - tree", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"natural": "tree"}, (*Meta).IsNatural, true)
	})

	t.Run("IsNatural - not", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"amenity": "restaurant"}, (*Meta).IsNatural, false)
	})

	t.Run("IsWater - river", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"waterway": "river"}, (*Meta).IsWater, true)
	})

	t.Run("IsWater - not", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"highway": "primary"}, (*Meta).IsWater, false)
	})

	t.Run("IsBuilding - yes", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"building": "yes"}, (*Meta).IsBuilding, true)
	})

	t.Run("IsBuilding - not", func(t *testing.T) {
		t.Parallel()
		testCategoryHelperMethod(t, map[string]string{"amenity": "restaurant"}, (*Meta).IsBuilding, false)
	})
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

	for _, testCase := range testCases {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: testCase.tags}

			got := meta.GetName()
			if got != testCase.expected {
				t.Errorf("expected %s, got %s", testCase.expected, got)
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

	for _, testCase := range testCases {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: testCase.tags}

			got := meta.IsRoad()
			if got != testCase.expected {
				t.Errorf("expected %v, got %v", testCase.expected, got)
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

	for _, testCase := range testCases {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			meta := Meta{Tags: testCase.tags}

			got := meta.IsFoodRelated()
			if got != testCase.expected {
				t.Errorf("expected %v, got %v", testCase.expected, got)
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

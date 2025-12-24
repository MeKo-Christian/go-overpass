package overpass

// Category represents high-level OSM feature category.
type Category string

const (
	CategoryTransportation Category = "transportation"
	CategoryAmenity        Category = "amenity"
	CategoryNatural        Category = "natural"
	CategoryWater          Category = "water"
	CategoryBuilding       Category = "building"
	CategoryLeisure        Category = "leisure"
	CategoryLanduse        Category = "landuse"
	CategoryBoundary       Category = "boundary"
	CategoryPlace          Category = "place"
	CategoryShop           Category = "shop"
	CategoryTourism        Category = "tourism"
	CategoryUnknown        Category = "unknown"
)

var tagToCategoryMap = map[string]Category{ //nolint:gochecknoglobals // lookup table for category detection
	"highway":  CategoryTransportation,
	"railway":  CategoryTransportation,
	"aeroway":  CategoryTransportation,
	"amenity":  CategoryAmenity,
	"natural":  CategoryNatural,
	"waterway": CategoryWater,
	"building": CategoryBuilding,
	"leisure":  CategoryLeisure,
	"landuse":  CategoryLanduse,
	"boundary": CategoryBoundary,
	"place":    CategoryPlace,
	"shop":     CategoryShop,
	"tourism":  CategoryTourism,
}

var categoryPriorityOrder = []string{ //nolint:gochecknoglobals // defines priority order for category detection
	"highway", "railway", "aeroway", "amenity", "natural", "waterway",
	"building", "leisure", "landuse", "boundary", "place", "shop", "tourism",
}

// GetCategory returns high-level category based on OSM tags.
func (m *Meta) GetCategory() Category {
	for _, tag := range categoryPriorityOrder {
		if _, ok := m.Tags[tag]; ok {
			return tagToCategoryMap[tag]
		}
	}

	return CategoryUnknown
}

// lookup table for subcategory detection
//
//nolint:gochecknoglobals
var categoryToSubcategoryTags = map[Category][]string{
	CategoryTransportation: {"highway", "railway", "aeroway"},
	CategoryAmenity:        {"amenity"},
	CategoryNatural:        {"natural"},
	CategoryWater:          {"waterway"},
	CategoryBuilding:       {"building"},
	CategoryLeisure:        {"leisure"},
	CategoryLanduse:        {"landuse"},
	CategoryBoundary:       {"boundary"},
	CategoryPlace:          {"place"},
	CategoryShop:           {"shop"},
	CategoryTourism:        {"tourism"},
}

// GetSubcategory returns detailed subcategory (tag value).
func (m *Meta) GetSubcategory() string {
	category := m.GetCategory()

	// Look for subcategory tags in the order defined for this category
	for _, tag := range categoryToSubcategoryTags[category] {
		if v, ok := m.Tags[tag]; ok {
			return v
		}
	}

	return ""
}

// IsTransportation checks if element is transportation-related.
func (m *Meta) IsTransportation() bool {
	return m.GetCategory() == CategoryTransportation
}

// IsAmenity checks if element is an amenity.
func (m *Meta) IsAmenity() bool {
	return m.GetCategory() == CategoryAmenity
}

// IsNatural checks if element is natural feature.
func (m *Meta) IsNatural() bool {
	return m.GetCategory() == CategoryNatural
}

// IsWater checks if element is water-related.
func (m *Meta) IsWater() bool {
	return m.GetCategory() == CategoryWater
}

// IsBuilding checks if element is a building.
func (m *Meta) IsBuilding() bool {
	return m.GetCategory() == CategoryBuilding
}

// GetName returns the name tag value if present.
func (m *Meta) GetName() string {
	if name, ok := m.Tags["name"]; ok {
		return name
	}

	return ""
}

// HasTag checks if specific tag exists.
func (m *Meta) HasTag(key string) bool {
	_, ok := m.Tags[key]
	return ok
}

// GetTag returns tag value with default fallback.
func (m *Meta) GetTag(key, defaultValue string) string {
	if value, ok := m.Tags[key]; ok {
		return value
	}

	return defaultValue
}

// MatchesFilter checks if element matches tag filter.
func (m *Meta) MatchesFilter(key, value string) bool {
	if tagValue, ok := m.Tags[key]; ok {
		return tagValue == value
	}

	return false
}

// Transportation subcategory helpers

// IsRoad checks if element is a road (highway).
func (m *Meta) IsRoad() bool {
	_, ok := m.Tags["highway"]
	return ok
}

// IsRailway checks if element is railway.
func (m *Meta) IsRailway() bool {
	_, ok := m.Tags["railway"]
	return ok
}

// Amenity subcategory helpers

// IsFoodRelated checks if amenity is food/drink related.
func (m *Meta) IsFoodRelated() bool {
	if amenity, ok := m.Tags["amenity"]; ok {
		return amenity == "restaurant" ||
			amenity == "cafe" ||
			amenity == "fast_food" ||
			amenity == "bar" ||
			amenity == "pub" ||
			amenity == "food_court" ||
			amenity == "biergarten"
	}

	return false
}

// IsEducation checks if amenity is education-related.
func (m *Meta) IsEducation() bool {
	if amenity, ok := m.Tags["amenity"]; ok {
		return amenity == "school" ||
			amenity == "university" ||
			amenity == "college" ||
			amenity == "library" ||
			amenity == "kindergarten"
	}

	return false
}

// IsHealthcare checks if amenity is healthcare-related.
func (m *Meta) IsHealthcare() bool {
	if amenity, ok := m.Tags["amenity"]; ok {
		return amenity == "hospital" ||
			amenity == "clinic" ||
			amenity == "doctors" ||
			amenity == "dentist" ||
			amenity == "pharmacy"
	}

	return false
}

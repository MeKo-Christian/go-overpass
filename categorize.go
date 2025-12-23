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

// GetCategory returns high-level category based on OSM tags.
func (m *Meta) GetCategory() Category {
	// Check tags in priority order
	if _, ok := m.Tags["highway"]; ok {
		return CategoryTransportation
	}

	if _, ok := m.Tags["railway"]; ok {
		return CategoryTransportation
	}

	if _, ok := m.Tags["aeroway"]; ok {
		return CategoryTransportation
	}

	if _, ok := m.Tags["amenity"]; ok {
		return CategoryAmenity
	}

	if _, ok := m.Tags["natural"]; ok {
		return CategoryNatural
	}

	if _, ok := m.Tags["waterway"]; ok {
		return CategoryWater
	}

	if _, ok := m.Tags["building"]; ok {
		return CategoryBuilding
	}

	if _, ok := m.Tags["leisure"]; ok {
		return CategoryLeisure
	}

	if _, ok := m.Tags["landuse"]; ok {
		return CategoryLanduse
	}

	if _, ok := m.Tags["boundary"]; ok {
		return CategoryBoundary
	}

	if _, ok := m.Tags["place"]; ok {
		return CategoryPlace
	}

	if _, ok := m.Tags["shop"]; ok {
		return CategoryShop
	}

	if _, ok := m.Tags["tourism"]; ok {
		return CategoryTourism
	}

	return CategoryUnknown
}

// GetSubcategory returns detailed subcategory (tag value).
func (m *Meta) GetSubcategory() string {
	category := m.GetCategory()

	switch category {
	case CategoryTransportation:
		if v, ok := m.Tags["highway"]; ok {
			return v
		}

		if v, ok := m.Tags["railway"]; ok {
			return v
		}

		if v, ok := m.Tags["aeroway"]; ok {
			return v
		}
	case CategoryAmenity:
		if v, ok := m.Tags["amenity"]; ok {
			return v
		}
	case CategoryNatural:
		if v, ok := m.Tags["natural"]; ok {
			return v
		}
	case CategoryWater:
		if v, ok := m.Tags["waterway"]; ok {
			return v
		}
	case CategoryBuilding:
		if v, ok := m.Tags["building"]; ok {
			// "yes" is generic building, return as-is
			return v
		}
	case CategoryLeisure:
		if v, ok := m.Tags["leisure"]; ok {
			return v
		}
	case CategoryLanduse:
		if v, ok := m.Tags["landuse"]; ok {
			return v
		}
	case CategoryBoundary:
		if v, ok := m.Tags["boundary"]; ok {
			return v
		}
	case CategoryPlace:
		if v, ok := m.Tags["place"]; ok {
			return v
		}
	case CategoryShop:
		if v, ok := m.Tags["shop"]; ok {
			return v
		}
	case CategoryTourism:
		if v, ok := m.Tags["tourism"]; ok {
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

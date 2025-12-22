package turbo

import (
	"fmt"
	"strings"
)

// Geocoder resolves a free-form query into OSM identifiers and geometry.
// Implementations can call external services (e.g., Nominatim).
type Geocoder interface {
	Geocode(query string) (GeocodeResult, error)
}

// GeocodeResult describes the first geocoding match.
type GeocodeResult struct {
	OSMType string
	OSMID   int64
	AreaID  int64
	BBox    *BBox
	Center  *Center
}

func expandGeocode(content string, opts Options) (string, error) {
	if opts.Geocoder == nil {
		return "", ErrMissingGeocoder
	}

	kind, query, ok := parseGeocodeMacro(content)
	if !ok {
		return "", ErrBadMacro
	}

	result, err := opts.Geocoder.Geocode(query)
	if err != nil {
		return "", err
	}

	switch kind {
	case "geocodeId":
		t, ok := normalizeOSMType(result.OSMType)
		if !ok || result.OSMID <= 0 {
			return "", ErrGeocodeData
		}
		return fmt.Sprintf("%s(%d)", t, result.OSMID), nil
	case "geocodeArea":
		areaID := result.AreaID
		if areaID == 0 {
			var err error
			areaID, err = deriveAreaID(result)
			if err != nil {
				return "", err
			}
		}
		return fmt.Sprintf("area(%d)", areaID), nil
	case "geocodeBbox":
		if result.BBox == nil {
			return "", ErrGeocodeData
		}
		return formatBBox(*result.BBox), nil
	case "geocodeCoords":
		if result.Center == nil {
			return "", ErrGeocodeData
		}
		return formatCenter(*result.Center), nil
	default:
		return "", ErrBadMacro
	}
}

func parseGeocodeMacro(content string) (string, string, bool) {
	parts := strings.SplitN(content, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	kind := strings.TrimSpace(parts[0])
	query := strings.TrimSpace(parts[1])
	if kind == "" || query == "" {
		return "", "", false
	}

	switch kind {
	case "geocodeId", "geocodeArea", "geocodeBbox", "geocodeCoords":
		return kind, query, true
	default:
		return "", "", false
	}
}

func normalizeOSMType(t string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "node":
		return "node", true
	case "way":
		return "way", true
	case "relation":
		return "relation", true
	default:
		return "", false
	}
}

func deriveAreaID(result GeocodeResult) (int64, error) {
	t, ok := normalizeOSMType(result.OSMType)
	if !ok || result.OSMID <= 0 {
		return 0, ErrGeocodeData
	}

	switch t {
	case "relation":
		return 3600000000 + result.OSMID, nil
	case "way":
		return 2400000000 + result.OSMID, nil
	default:
		return 0, ErrGeocodeData
	}
}

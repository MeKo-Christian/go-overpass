package turbo

import (
	"fmt"
	"strings"
)

const (
	osmTypeNode     = "node"
	osmTypeWay      = "way"
	osmTypeRelation = "relation"
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

func expandGeocodeID(result GeocodeResult, format QueryFormat) (string, error) {
	typeStr, ok := normalizeOSMType(result.OSMType)
	if !ok || result.OSMID <= 0 {
		return "", ErrGeocodeData
	}

	if format == FormatXML {
		return fmt.Sprintf(`type="%s" ref="%d"`, typeStr, result.OSMID), nil
	}

	return fmt.Sprintf("%s(%d)", typeStr, result.OSMID), nil
}

func expandGeocodeArea(result GeocodeResult, format QueryFormat) (string, error) {
	areaID := result.AreaID
	if areaID == 0 {
		var err error

		areaID, err = deriveAreaID(result)
		if err != nil {
			return "", err
		}
	}

	if format == FormatXML {
		return fmt.Sprintf(`type="area" ref="%d"`, areaID), nil
	}

	return fmt.Sprintf("area(%d)", areaID), nil
}

func expandGeocodeBbox(result GeocodeResult, format QueryFormat) (string, error) {
	if result.BBox == nil {
		return "", ErrGeocodeData
	}

	return formatBBox(*result.BBox, format), nil
}

func expandGeocodeCoords(result GeocodeResult, format QueryFormat) (string, error) {
	if result.Center == nil {
		return "", ErrGeocodeData
	}

	return formatCenter(*result.Center, format), nil
}

func expandGeocode(content string, opts Options, format QueryFormat) (string, error) {
	if opts.Geocoder == nil {
		return "", ErrMissingGeocoder
	}

	kind, query, ok := parseGeocodeMacro(content)
	if !ok {
		return "", ErrBadMacro
	}

	result, err := opts.Geocoder.Geocode(query)
	if err != nil {
		return "", fmt.Errorf("geocoding failed: %w", err)
	}

	switch kind {
	case "geocodeId":
		return expandGeocodeID(result, format)
	case "geocodeArea":
		return expandGeocodeArea(result, format)
	case "geocodeBbox":
		return expandGeocodeBbox(result, format)
	case "geocodeCoords":
		return expandGeocodeCoords(result, format)
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
	case osmTypeNode:
		return osmTypeNode, true
	case osmTypeWay:
		return osmTypeWay, true
	case osmTypeRelation:
		return osmTypeRelation, true
	default:
		return "", false
	}
}

func deriveAreaID(result GeocodeResult) (int64, error) {
	typeStr, ok := normalizeOSMType(result.OSMType)
	if !ok || result.OSMID <= 0 {
		return 0, ErrGeocodeData
	}

	switch typeStr {
	case osmTypeRelation:
		return 3600000000 + result.OSMID, nil
	case osmTypeWay:
		return 2400000000 + result.OSMID, nil
	default:
		return 0, ErrGeocodeData
	}
}

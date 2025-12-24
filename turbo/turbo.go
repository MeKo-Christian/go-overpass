package turbo

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// BBox represents a bounding box in south, west, north, east order.
type BBox struct {
	South float64
	West  float64
	North float64
	East  float64
}

// Center represents a latitude/longitude pair.
type Center struct {
	Lat float64
	Lon float64
}

// DataSource describes a {{data:...}} macro.
type DataSource struct {
	Mode    string
	Options map[string]string
	Parsed  DataOptions
}

// DataOptions exposes typed access to common {{data:...}} parameters.
type DataOptions struct {
	Server string
	Params map[string]string
}

// Options control macro expansion.
type Options struct {
	BBox      *BBox
	Center    *Center
	Now       time.Time
	Shortcuts map[string]string
	Geocoder  Geocoder
	Format    QueryFormat
}

// Result holds the expanded query and any extracted metadata.
type Result struct {
	Query  string
	Style  string
	Styles []string
	// ParsedStyles contains the parsed MapCSS stylesheets from {{style:...}} macros.
	// Each element corresponds to the same index in Styles.
	ParsedStyles []*Stylesheet
	Data         *DataSource
	// EndpointOverride suggests an Overpass API endpoint derived from {{data:overpass,server=...}}.
	EndpointOverride string
	// DataServer suggests the backend server from {{data:...,server=...}} (overpass or sql).
	DataServer string
}

// QueryFormat controls how macros are expanded.
type QueryFormat int

const (
	FormatAuto QueryFormat = iota
	FormatQL
	FormatXML
)

// Time unit constants for date macro parsing.
const (
	unitSecond = "second"
	unitMinute = "minute"
	unitHour   = "hour"
	unitDay    = "day"
	unitWeek   = "week"
	unitMonth  = "month"
	unitYear   = "year"
)

var (
	ErrMissingBBox     = errors.New("turbo: bbox not provided")
	ErrMissingCenter   = errors.New("turbo: center not provided")
	ErrMissingGeocoder = errors.New("turbo: geocoder not provided")
	ErrGeocodeData     = errors.New("turbo: geocoder result missing data")
	ErrBadMacro        = errors.New("turbo: unsupported or malformed macro")
)

// Expand replaces a subset of Overpass Turbo macros with Overpass QL compatible text.
//
// Supported macros:
//   - {{bbox}} and {{center}} using Options.BBox/Options.Center
//   - {{date}} and {{date:<n unit>}} using Options.Now (UTC if set, else time.Now)
//   - Custom shortcuts: {{key=value}} defines {{key}}
//   - {{style:...}} and {{data:...}} are removed from output and returned in Result
//
// Unsupported geocode macros return an error for now.
func Expand(query string, opts Options) (Result, error) { //nolint:gocognit // complex macro expansion logic
	format := detectFormat(query, opts.Format)

	shortcuts := map[string]string{}
	for k, v := range opts.Shortcuts {
		shortcuts[k] = v
	}

	err := scanMacros(query, func(_ int, _ int, content string) error {
		name, value, ok := parseShortcutDefinition(content)
		if ok {
			shortcuts[name] = value
		}

		return nil
	})
	if err != nil {
		return Result{}, err
	}

	var res Result

	expanded, err := replaceMacros(query, func(content string) (string, error) {
		content = strings.TrimSpace(content)
		if content == "" {
			return "", ErrBadMacro
		}

		if _, _, ok := parseShortcutDefinition(content); ok {
			return "", nil
		}

		if strings.HasPrefix(content, "style:") {
			style := strings.TrimSpace(strings.TrimPrefix(content, "style:"))

			res.Style = style
			if style != "" {
				res.Styles = append(res.Styles, style)
				// Parse the MapCSS stylesheet
				parsed, err := ParseMapCSS(style)
				if err != nil {
					// Store nil for unparseable styles, don't fail the expansion
					res.ParsedStyles = append(res.ParsedStyles, nil)
				} else {
					res.ParsedStyles = append(res.ParsedStyles, parsed)
				}
			}

			return "", nil
		}

		if strings.HasPrefix(content, "data:") {
			dataSrc, err := parseDataSource(strings.TrimPrefix(content, "data:"))
			if err != nil {
				return "", err
			}

			res.Data = dataSrc
			if server, ok := dataSrc.Options["server"]; ok {
				res.DataServer = server
			}

			if strings.EqualFold(dataSrc.Mode, "overpass") {
				if server, ok := dataSrc.Options["server"]; ok {
					res.EndpointOverride = normalizeEndpoint(server)
				}
			}

			return "", nil
		}

		if strings.HasPrefix(content, "geocode") {
			return expandGeocode(content, opts, format)
		}

		if content == "bbox" {
			if opts.BBox == nil {
				return "", ErrMissingBBox
			}

			return formatBBox(*opts.BBox, format), nil
		}

		if content == "center" {
			if opts.Center == nil {
				return "", ErrMissingCenter
			}

			return formatCenter(*opts.Center, format), nil
		}

		if strings.HasPrefix(content, "date") {
			t, err := expandDate(content, opts.Now)
			if err != nil {
				return "", err
			}

			return t, nil
		}

		if value, ok := shortcuts[content]; ok {
			return value, nil
		}

		return "", ErrBadMacro
	})
	if err != nil {
		return Result{}, err
	}

	res.Query = expanded

	return res, nil
}

// ApplyEndpointOverride returns the endpoint to use based on Result.EndpointOverride.
// If no override is present, the fallback endpoint is returned.
func ApplyEndpointOverride(fallback string, res Result) string {
	if res.EndpointOverride == "" {
		return fallback
	}

	return res.EndpointOverride
}

func normalizeEndpoint(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	trimmed = strings.TrimSuffix(trimmed, "/")

	if strings.HasSuffix(trimmed, "/api") {
		return trimmed + "/interpreter"
	}

	if strings.HasSuffix(trimmed, "/api/interpreter") {
		return trimmed
	}

	if strings.Contains(trimmed, "/api/interpreter") {
		return trimmed
	}

	if strings.HasSuffix(trimmed, "/api/") {
		return strings.TrimSuffix(trimmed, "/") + "/interpreter"
	}

	return trimmed
}

func detectFormat(query string, format QueryFormat) QueryFormat {
	if format != FormatAuto {
		return format
	}

	if strings.Contains(query, "<osm-script") || strings.Contains(query, "<query") {
		return FormatXML
	}

	return FormatQL
}

func formatBBox(bbox BBox, format QueryFormat) string {
	switch format {
	case FormatXML:
		return fmt.Sprintf(`s="%s" w="%s" n="%s" e="%s"`,
			formatFloat(bbox.South),
			formatFloat(bbox.West),
			formatFloat(bbox.North),
			formatFloat(bbox.East),
		)
	default:
		return fmt.Sprintf("%s,%s,%s,%s",
			formatFloat(bbox.South),
			formatFloat(bbox.West),
			formatFloat(bbox.North),
			formatFloat(bbox.East),
		)
	}
}

func formatCenter(center Center, format QueryFormat) string {
	switch format {
	case FormatXML:
		return fmt.Sprintf(`lat="%s" lon="%s"`,
			formatFloat(center.Lat),
			formatFloat(center.Lon),
		)
	default:
		return fmt.Sprintf("%s,%s",
			formatFloat(center.Lat),
			formatFloat(center.Lon),
		)
	}
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func parseShortcutDefinition(content string) (string, string, bool) {
	if strings.HasPrefix(strings.TrimSpace(content), "data:") || strings.HasPrefix(strings.TrimSpace(content), "style:") {
		return "", "", false
	}

	parts := strings.SplitN(content, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	name := strings.TrimSpace(parts[0])
	if name == "" || strings.Contains(name, ":") {
		return "", "", false
	}

	return name, strings.TrimSpace(parts[1]), true
}

func parseDataSource(raw string) (*DataSource, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrBadMacro
	}

	parts := strings.Split(raw, ",")

	mode := strings.TrimSpace(parts[0])
	if mode == "" {
		return nil, ErrBadMacro
	}

	options := map[string]string{}
	parsed := DataOptions{}

	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			return nil, ErrBadMacro
		}

		key := strings.TrimSpace(keyValue[0])
		if key == "" {
			return nil, ErrBadMacro
		}

		value := strings.TrimSpace(keyValue[1])

		options[key] = value
		if key == "server" {
			parsed.Server = value
		} else {
			if parsed.Params == nil {
				parsed.Params = map[string]string{}
			}

			parsed.Params[key] = value
		}
	}

	return &DataSource{
		Mode:    mode,
		Options: options,
		Parsed:  parsed,
	}, nil
}

func applyDateOffset(base time.Time, value int, unit string) time.Time {
	switch unit {
	case unitSecond:
		return base.Add(-time.Duration(value) * time.Second)
	case unitMinute:
		return base.Add(-time.Duration(value) * time.Minute)
	case unitHour:
		return base.Add(-time.Duration(value) * time.Hour)
	case unitDay:
		return base.Add(-time.Duration(value) * 24 * time.Hour)
	case unitWeek:
		return base.Add(-time.Duration(value) * 7 * 24 * time.Hour)
	case unitMonth:
		return base.AddDate(0, -value, 0)
	case unitYear:
		return base.AddDate(-value, 0, 0)
	}

	return base
}

func expandDate(content string, now time.Time) (string, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	raw := strings.TrimSpace(strings.TrimPrefix(content, "date"))
	if raw == "" {
		return now.Format(time.RFC3339Nano), nil
	}

	if !strings.HasPrefix(raw, ":") {
		return "", ErrBadMacro
	}

	raw = strings.TrimSpace(strings.TrimPrefix(raw, ":"))
	if raw == "" {
		return "", ErrBadMacro
	}

	value, unit, err := parseRelativeDuration(raw)
	if err != nil {
		return "", err
	}

	if !isValidUnit(unit) {
		return "", ErrBadMacro
	}

	now = applyDateOffset(now, value, unit)

	return now.Format(time.RFC3339Nano), nil
}

func isValidUnit(unit string) bool {
	switch unit {
	case unitSecond, unitMinute, unitHour, unitDay, unitWeek, unitMonth, unitYear:
		return true
	default:
		return false
	}
}

var unitMap = map[string]string{ //nolint:gochecknoglobals // lookup table for unit parsing
	"second": unitSecond, "seconds": unitSecond,
	"minute": unitMinute, "minutes": unitMinute,
	"hour": unitHour, "hours": unitHour,
	"day": unitDay, "days": unitDay,
	"week": unitWeek, "weeks": unitWeek,
	"month": unitMonth, "months": unitMonth,
	"year": unitYear, "years": unitYear,
}

func parseRelativeDuration(raw string) (int, string, error) {
	fields := strings.Fields(raw)
	if len(fields) != 2 {
		return 0, "", ErrBadMacro
	}

	value, err := strconv.Atoi(fields[0])
	if err != nil || value < 0 {
		return 0, "", ErrBadMacro
	}

	unit, ok := unitMap[strings.ToLower(fields[1])]
	if !ok {
		return 0, "", ErrBadMacro
	}

	return value, unit, nil
}

func scanMacros(query string, callback func(start int, end int, content string) error) error {
	for pos := 0; pos < len(query); {
		openIdx := strings.Index(query[pos:], "{{")
		if openIdx == -1 {
			return nil
		}

		openIdx += pos

		closeIdx := strings.Index(query[openIdx+2:], "}}")
		if closeIdx == -1 {
			return ErrBadMacro
		}

		closeIdx = closeIdx + openIdx + 2

		content := query[openIdx+2 : closeIdx]

		err := callback(openIdx, closeIdx+2, content)
		if err != nil {
			return err
		}

		pos = closeIdx + 2
	}

	return nil
}

func replaceMacros(query string, replace func(content string) (string, error)) (string, error) {
	var out bytes.Buffer
	last := 0

	err := scanMacros(query, func(start int, end int, content string) error {
		out.WriteString(query[last:start])

		value, err := replace(content)
		if err != nil {
			return err
		}

		out.WriteString(value)

		last = end

		return nil
	})
	if err != nil {
		return "", err
	}

	out.WriteString(query[last:])

	return out.String(), nil
}

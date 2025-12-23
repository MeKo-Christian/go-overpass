package turbo

import (
	"strings"
	"testing"
	"time"
)

func TestExpandBBoxAndCenter(t *testing.T) {
	query := "node({{bbox}});out;{{center}}"
	res, err := Expand(query, Options{
		BBox:   &BBox{South: 1.1, West: 2.2, North: 3.3, East: 4.4},
		Center: &Center{Lat: 5.5, Lon: 6.6},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(res.Query, "1.1,2.2,3.3,4.4") {
		t.Fatalf("bbox not expanded: %s", res.Query)
	}
	if !strings.Contains(res.Query, "5.5,6.6") {
		t.Fatalf("center not expanded: %s", res.Query)
	}
}

func TestExpandDate(t *testing.T) {
	now := time.Date(2024, 2, 10, 12, 30, 0, 0, time.UTC)
	res, err := Expand(`node["check_date" > "{{date:1 day}}"];out;`, Options{
		Now: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(res.Query, "2024-02-09T12:30:00Z") {
		t.Fatalf("date not expanded: %s", res.Query)
	}
}

func TestCustomShortcuts(t *testing.T) {
	query := "{{foo=bar}}node({{foo}});out;"
	res, err := Expand(query, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(res.Query, "{{foo=bar}}") {
		t.Fatalf("shortcut definition not removed: %s", res.Query)
	}
	if !strings.Contains(res.Query, "node(bar)") {
		t.Fatalf("shortcut not expanded: %s", res.Query)
	}
}

func TestStyleAndDataExtraction(t *testing.T) {
	query := `{{style:line[highway=path]{color:red;}}}
{{data:overpass,server=https://overpass-api.de/api/}}
node(1);out;`
	res, err := Expand(query, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Style == "" || res.Data == nil {
		t.Fatalf("expected style and data to be extracted")
	}
	if len(res.Styles) != 1 {
		t.Fatalf("expected 1 style, got %d", len(res.Styles))
	}
	if res.DataServer != "https://overpass-api.de/api/" {
		t.Fatalf("expected data server to be captured, got %q", res.DataServer)
	}
	if res.EndpointOverride != "https://overpass-api.de/api/interpreter" {
		t.Fatalf("expected endpoint override to be captured, got %q", res.EndpointOverride)
	}
	if res.Data.Parsed.Server != "https://overpass-api.de/api/" {
		t.Fatalf("expected parsed server to be captured, got %q", res.Data.Parsed.Server)
	}
	if strings.Contains(res.Query, "style:") || strings.Contains(res.Query, "data:") {
		t.Fatalf("style/data macros should be removed: %s", res.Query)
	}
}

func TestMultipleStyles(t *testing.T) {
	query := `{{style:a}}node(1);out;{{style:b}}`
	res, err := Expand(query, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Style != "b" {
		t.Fatalf("expected last style in Style, got %q", res.Style)
	}
	if len(res.Styles) != 2 {
		t.Fatalf("expected 2 styles, got %d", len(res.Styles))
	}
}

func TestApplyEndpointOverride(t *testing.T) {
	res := Result{EndpointOverride: "https://example.com/api/interpreter"}
	endpoint := ApplyEndpointOverride("https://default/api/interpreter", res)
	if endpoint != "https://example.com/api/interpreter" {
		t.Fatalf("unexpected endpoint override: %s", endpoint)
	}

	endpoint = ApplyEndpointOverride("https://default/api/interpreter", Result{})
	if endpoint != "https://default/api/interpreter" {
		t.Fatalf("unexpected fallback endpoint: %s", endpoint)
	}
}

func TestSQLDataSource(t *testing.T) {
	query := `{{data:sql,server=https://postpass.example/api/0.2/,token=abc}}
node(1);out;`
	res, err := Expand(query, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Data == nil || res.Data.Mode != "sql" {
		t.Fatalf("expected sql data mode")
	}
	if res.DataServer != "https://postpass.example/api/0.2/" {
		t.Fatalf("expected data server to be captured, got %q", res.DataServer)
	}
	if res.Data.Parsed.Server != "https://postpass.example/api/0.2/" {
		t.Fatalf("expected parsed server to be captured, got %q", res.Data.Parsed.Server)
	}
	if res.Data.Parsed.Params["token"] != "abc" {
		t.Fatalf("expected parsed token param, got %q", res.Data.Parsed.Params["token"])
	}
	if res.EndpointOverride != "" {
		t.Fatalf("did not expect endpoint override for sql mode")
	}
}

func TestSQLDataConfigFromResult(t *testing.T) {
	query := `{{data:sql,server=https://postpass.example/api/0.2/,token=abc,foo=bar}}
node(1);out;`
	res, err := Expand(query, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := SQLDataConfigFromResult(res)
	if cfg == nil {
		t.Fatalf("expected sql config")
	}
	if cfg.Server != "https://postpass.example/api/0.2/" {
		t.Fatalf("unexpected server: %s", cfg.Server)
	}
	if cfg.Params["token"] != "abc" || cfg.Params["foo"] != "bar" {
		t.Fatalf("unexpected params: %#v", cfg.Params)
	}
	if _, ok := cfg.Params["server"]; ok {
		t.Fatalf("server should not be in params")
	}
}

func TestUnsupportedMacro(t *testing.T) {
	_, err := Expand("node({{geocodeArea:Vienna}});out;", Options{})
	if err == nil {
		t.Fatalf("expected error for unsupported macro")
	}
}

func TestMissingGeocoder(t *testing.T) {
	_, err := Expand("node({{geocodeId:Vienna}});out;", Options{})
	if err == nil {
		t.Fatalf("expected error for missing geocoder")
	}
	if err != ErrMissingGeocoder {
		t.Fatalf("expected ErrMissingGeocoder, got %v", err)
	}
}

type fakeGeocoder struct {
	result GeocodeResult
	err    error
}

func (f fakeGeocoder) Geocode(query string) (GeocodeResult, error) {
	return f.result, f.err
}

func TestGeocodeMacros(t *testing.T) {
	geocoder := fakeGeocoder{
		result: GeocodeResult{
			OSMType: "relation",
			OSMID:   1645,
			BBox:    &BBox{South: 1, West: 2, North: 3, East: 4},
			Center:  &Center{Lat: 5, Lon: 6},
		},
	}

	res, err := Expand("{{geocodeId:Vienna}};{{geocodeArea:Vienna}};{{geocodeBbox:Vienna}};{{geocodeCoords:Vienna}}", Options{
		Geocoder: geocoder,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(res.Query, "relation(1645)") {
		t.Fatalf("geocodeId not expanded: %s", res.Query)
	}
	if !strings.Contains(res.Query, "area(3600001645)") {
		t.Fatalf("geocodeArea not expanded: %s", res.Query)
	}
	if !strings.Contains(res.Query, "1,2,3,4") {
		t.Fatalf("geocodeBbox not expanded: %s", res.Query)
	}
	if !strings.Contains(res.Query, "5,6") {
		t.Fatalf("geocodeCoords not expanded: %s", res.Query)
	}
}

func TestXMLMacroExpansion(t *testing.T) {
	query := `<osm-script><query {{bbox}}/><center {{center}}/></osm-script>`
	res, err := Expand(query, Options{
		BBox:   &BBox{South: 1, West: 2, North: 3, East: 4},
		Center: &Center{Lat: 5, Lon: 6},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(res.Query, `s="1" w="2" n="3" e="4"`) {
		t.Fatalf("xml bbox not expanded: %s", res.Query)
	}
	if !strings.Contains(res.Query, `lat="5" lon="6"`) {
		t.Fatalf("xml center not expanded: %s", res.Query)
	}
}

func TestXMLGeocodeExpansion(t *testing.T) {
	geocoder := fakeGeocoder{
		result: GeocodeResult{
			OSMType: "relation",
			OSMID:   1645,
		},
	}
	query := `<osm-script><query {{geocodeId:Vienna}} {{geocodeArea:Vienna}}/></osm-script>`
	res, err := Expand(query, Options{
		Geocoder: geocoder,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(res.Query, `type="relation" ref="1645"`) {
		t.Fatalf("xml geocodeId not expanded: %s", res.Query)
	}
	if !strings.Contains(res.Query, `type="area" ref="3600001645"`) {
		t.Fatalf("xml geocodeArea not expanded: %s", res.Query)
	}
}

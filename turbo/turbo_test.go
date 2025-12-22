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
	if res.EndpointOverride != "https://overpass-api.de/api/interpreter" {
		t.Fatalf("expected endpoint override to be captured, got %q", res.EndpointOverride)
	}
	if strings.Contains(res.Query, "style:") || strings.Contains(res.Query, "data:") {
		t.Fatalf("style/data macros should be removed: %s", res.Query)
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

func TestUnsupportedMacro(t *testing.T) {
	_, err := Expand("node({{geocodeArea:Vienna}});out;", Options{})
	if err == nil {
		t.Fatalf("expected error for unsupported macro")
	}
}

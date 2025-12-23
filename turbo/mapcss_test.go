package turbo

import (
	"testing"
)

func TestParseMapCSSBasicSelector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantType string
		wantLen  int
	}{
		{
			name:     "node selector",
			input:    "node { color: red; }",
			wantType: "node",
			wantLen:  1,
		},
		{
			name:     "way selector",
			input:    "way { color: blue; }",
			wantType: "way",
			wantLen:  1,
		},
		{
			name:     "relation selector",
			input:    "relation { color: green; }",
			wantType: "relation",
			wantLen:  1,
		},
		{
			name:     "area selector",
			input:    "area { fill-color: yellow; }",
			wantType: "area",
			wantLen:  1,
		},
		{
			name:     "line selector",
			input:    "line { width: 2; }",
			wantType: "line",
			wantLen:  1,
		},
		{
			name:     "wildcard selector",
			input:    "* { opacity: 0.5; }",
			wantType: "*",
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss, err := ParseMapCSS(tt.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(ss.Rules) != tt.wantLen {
				t.Errorf("got %d rules, want %d", len(ss.Rules), tt.wantLen)
			}

			if len(ss.Rules) > 0 && len(ss.Rules[0].Selectors) > 0 {
				if ss.Rules[0].Selectors[0].Type != tt.wantType {
					t.Errorf("got type %q, want %q", ss.Rules[0].Selectors[0].Type, tt.wantType)
				}
			}
		})
	}
}

func TestParseMapCSSConditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantKey string
		wantOp  string
		wantVal string
	}{
		{
			name:    "equals",
			input:   "way[highway=primary] { color: red; }",
			wantKey: "highway",
			wantOp:  "=",
			wantVal: "primary",
		},
		{
			name:    "not equals",
			input:   "way[highway!=motorway] { color: blue; }",
			wantKey: "highway",
			wantOp:  "!=",
			wantVal: "motorway",
		},
		{
			name:    "exists",
			input:   "way[name] { color: green; }",
			wantKey: "name",
			wantOp:  "",
			wantVal: "",
		},
		{
			name:    "not exists",
			input:   "way[!name] { color: gray; }",
			wantKey: "name",
			wantOp:  "!",
			wantVal: "",
		},
		{
			name:    "regex match",
			input:   `way[highway=~/.*ary/] { color: purple; }`,
			wantKey: "highway",
			wantOp:  "=~",
			wantVal: "/.*ary/",
		},
		{
			name:    "less than",
			input:   "node[population<1000] { color: yellow; }",
			wantKey: "population",
			wantOp:  "<",
			wantVal: "1000",
		},
		{
			name:    "greater than or equal",
			input:   "node[population>=1000000] { color: orange; }",
			wantKey: "population",
			wantOp:  ">=",
			wantVal: "1000000",
		},
		{
			name:    "meta attribute",
			input:   "way[@id=171784106] { color: cyan; }",
			wantKey: "@id",
			wantOp:  "=",
			wantVal: "171784106",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss, err := ParseMapCSS(tt.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(ss.Rules) == 0 || len(ss.Rules[0].Selectors) == 0 {
				t.Fatal("expected at least one rule with one selector")
			}

			sel := ss.Rules[0].Selectors[0]
			if len(sel.Conditions) == 0 {
				t.Fatal("expected at least one condition")
			}

			cond := sel.Conditions[0]
			if cond.Key != tt.wantKey {
				t.Errorf("got key %q, want %q", cond.Key, tt.wantKey)
			}

			if cond.Operator != tt.wantOp {
				t.Errorf("got operator %q, want %q", cond.Operator, tt.wantOp)
			}

			if cond.Value != tt.wantVal {
				t.Errorf("got value %q, want %q", cond.Value, tt.wantVal)
			}
		})
	}
}

func TestParseMapCSSColors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantR     float64
		wantG     float64
		wantB     float64
		wantA     float64
		tolerance float64
	}{
		{
			name:      "hex color 6 digits",
			input:     "node { color: #ff0000; }",
			wantR:     1.0,
			wantG:     0.0,
			wantB:     0.0,
			wantA:     1.0,
			tolerance: 0.01,
		},
		{
			name:      "hex color 3 digits",
			input:     "node { color: #f00; }",
			wantR:     1.0,
			wantG:     0.0,
			wantB:     0.0,
			wantA:     1.0,
			tolerance: 0.01,
		},
		{
			name:      "hex color with alpha",
			input:     "node { color: #ff000080; }",
			wantR:     1.0,
			wantG:     0.0,
			wantB:     0.0,
			wantA:     0.5,
			tolerance: 0.02,
		},
		{
			name:      "named color red",
			input:     "node { color: red; }",
			wantR:     1.0,
			wantG:     0.0,
			wantB:     0.0,
			wantA:     1.0,
			tolerance: 0.01,
		},
		{
			name:      "named color blue",
			input:     "node { color: blue; }",
			wantR:     0.0,
			wantG:     0.0,
			wantB:     1.0,
			wantA:     1.0,
			tolerance: 0.01,
		},
		{
			name:      "rgb function",
			input:     "node { color: rgb(1.0, 0.5, 0.0); }",
			wantR:     1.0,
			wantG:     0.5,
			wantB:     0.0,
			wantA:     1.0,
			tolerance: 0.01,
		},
		{
			name:      "rgba function",
			input:     "node { color: rgba(1.0, 0.0, 0.0, 0.5); }",
			wantR:     1.0,
			wantG:     0.0,
			wantB:     0.0,
			wantA:     0.5,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss, err := ParseMapCSS(tt.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(ss.Rules) == 0 || len(ss.Rules[0].Declarations) == 0 {
				t.Fatal("expected at least one rule with one declaration")
			}

			decl := ss.Rules[0].Declarations[0]
			if decl.Value.Type != ValueTypeColor {
				t.Fatalf("expected color type, got %v", decl.Value.Type)
			}

			if decl.Value.Color == nil {
				t.Fatal("expected non-nil color")
			}

			c := decl.Value.Color
			if abs(c.R-tt.wantR) > tt.tolerance {
				t.Errorf("R: got %.3f, want %.3f", c.R, tt.wantR)
			}

			if abs(c.G-tt.wantG) > tt.tolerance {
				t.Errorf("G: got %.3f, want %.3f", c.G, tt.wantG)
			}

			if abs(c.B-tt.wantB) > tt.tolerance {
				t.Errorf("B: got %.3f, want %.3f", c.B, tt.wantB)
			}

			if abs(c.A-tt.wantA) > tt.tolerance {
				t.Errorf("A: got %.3f, want %.3f", c.A, tt.wantA)
			}
		})
	}
}

func TestParseMapCSSDeclarations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantProp string
		wantType ValueType
		wantRaw  string
	}{
		{
			name:     "number value",
			input:    "way { width: 5; }",
			wantProp: "width",
			wantType: ValueTypeNumber,
			wantRaw:  "5",
		},
		{
			name:     "url value",
			input:    "node { icon-image: url('icons/maki/cafe-18.png'); }",
			wantProp: "icon-image",
			wantType: ValueTypeURL,
			wantRaw:  "url(icons/maki/cafe-18.png)",
		},
		{
			name:     "dashes value",
			input:    "way { dashes: 5, 8; }",
			wantProp: "dashes",
			wantType: ValueTypeDashes,
			wantRaw:  "5, 8",
		},
		{
			name:     "keyword value",
			input:    "way { linecap: round; }",
			wantProp: "linecap",
			wantType: ValueTypeKeyword,
			wantRaw:  "round",
		},
		{
			name:     "eval value",
			input:    `node { opacity: eval("tag('population')/100000"); }`,
			wantProp: "opacity",
			wantType: ValueTypeEval,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss, err := ParseMapCSS(tt.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(ss.Rules) == 0 || len(ss.Rules[0].Declarations) == 0 {
				t.Fatal("expected at least one rule with one declaration")
			}

			decl := ss.Rules[0].Declarations[0]
			if decl.Property != tt.wantProp {
				t.Errorf("got property %q, want %q", decl.Property, tt.wantProp)
			}

			if decl.Value.Type != tt.wantType {
				t.Errorf("got type %v, want %v", decl.Value.Type, tt.wantType)
			}
		})
	}
}

func TestParseMapCSSMultipleSelectors(t *testing.T) {
	t.Parallel()

	input := "way[highway=primary], way[highway=secondary] { color: red; }"

	ss, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(ss.Rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(ss.Rules))
	}

	if len(ss.Rules[0].Selectors) != 2 {
		t.Fatalf("got %d selectors, want 2", len(ss.Rules[0].Selectors))
	}
}

func TestParseMapCSSMultipleRules(t *testing.T) {
	t.Parallel()

	input := `
		way[highway=primary] { color: red; }
		way[highway=secondary] { color: blue; }
		node[amenity=cafe] { icon-image: url('cafe.png'); }
	`

	styleSheet, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(styleSheet.Rules) != 3 {
		t.Fatalf("got %d rules, want 3", len(styleSheet.Rules))
	}
}

func TestParseMapCSSComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name: "C-style comment",
			input: `
				/* This is a comment */
				way { color: red; }
			`,
		},
		{
			name: "line comment",
			input: `
				// This is a comment
				way { color: red; }
			`,
		},
		{
			name: "comment in rule",
			input: `
				way {
					/* comment */ color: red;
				}
			`,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss, err := ParseMapCSS(tt.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(ss.Rules) != 1 {
				t.Fatalf("got %d rules, want 1", len(ss.Rules))
			}
		})
	}
}

func TestParseMapCSSSetDirective(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantProp string
		wantVal  string
	}{
		{
			name:     "set class",
			input:    "way[highway=footpath] { set .minor_road; }",
			wantProp: "set-class",
			wantVal:  "minor_road",
		},
		{
			name:     "set tag with value",
			input:    "way { set layer=5; }",
			wantProp: "set-tag:layer",
			wantVal:  "5",
		},
		{
			name:     "set tag without value",
			input:    "way { set bridge; }",
			wantProp: "set-tag:bridge",
			wantVal:  "yes",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss, err := ParseMapCSS(tt.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(ss.Rules) == 0 || len(ss.Rules[0].Declarations) == 0 {
				t.Fatal("expected at least one rule with one declaration")
			}

			decl := ss.Rules[0].Declarations[0]
			if decl.Property != tt.wantProp {
				t.Errorf("got property %q, want %q", decl.Property, tt.wantProp)
			}

			if decl.Value.Raw != tt.wantVal {
				t.Errorf("got value %q, want %q", decl.Value.Raw, tt.wantVal)
			}
		})
	}
}

func TestParseMapCSSPseudoClasses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		pseudo []string
	}{
		{
			name:   "hover",
			input:  "way:hover { color: red; }",
			pseudo: []string{"hover"},
		},
		{
			name:   "active",
			input:  "way:active { color: blue; }",
			pseudo: []string{"active"},
		},
		{
			name:   "tagged",
			input:  "node:tagged { color: green; }",
			pseudo: []string{"tagged"},
		},
		{
			name:   "multiple pseudo-classes",
			input:  "way:hover:active { color: purple; }",
			pseudo: []string{"hover", "active"},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss, err := ParseMapCSS(tt.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(ss.Rules) == 0 || len(ss.Rules[0].Selectors) == 0 {
				t.Fatal("expected at least one rule with one selector")
			}

			sel := ss.Rules[0].Selectors[0]
			if len(sel.PseudoClasses) != len(tt.pseudo) {
				t.Fatalf("got %d pseudo-classes, want %d", len(sel.PseudoClasses), len(tt.pseudo))
			}

			for i, p := range tt.pseudo {
				if sel.PseudoClasses[i] != p {
					t.Errorf("pseudo-class %d: got %q, want %q", i, sel.PseudoClasses[i], p)
				}
			}
		})
	}
}

func TestParseMapCSSClassSelectors(t *testing.T) {
	t.Parallel()

	input := "way.minor_road { color: gray; }"

	ss, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(ss.Rules) == 0 || len(ss.Rules[0].Selectors) == 0 {
		t.Fatal("expected at least one rule with one selector")
	}

	sel := ss.Rules[0].Selectors[0]
	if len(sel.Classes) != 1 || sel.Classes[0] != "minor_road" {
		t.Errorf("got classes %v, want [minor_road]", sel.Classes)
	}
}

func TestParseMapCSSLayer(t *testing.T) {
	t.Parallel()

	input := "way::casing { width: 10; }"

	ss, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(ss.Rules) == 0 || len(ss.Rules[0].Selectors) == 0 {
		t.Fatal("expected at least one rule with one selector")
	}

	sel := ss.Rules[0].Selectors[0]
	if sel.Layer != "casing" {
		t.Errorf("got layer %q, want %q", sel.Layer, "casing")
	}
}

func TestParseMapCSSZoomRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantMin int
		wantMax int
	}{
		{
			name:    "single zoom",
			input:   "way|z12 { color: red; }",
			wantMin: 12,
			wantMax: 12,
		},
		{
			name:    "zoom range",
			input:   "way|z1-11 { color: blue; }",
			wantMin: 1,
			wantMax: 11,
		},
		{
			name:    "zoom min only",
			input:   "way|z12- { color: green; }",
			wantMin: 12,
			wantMax: 0,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			styleSheet, err := ParseMapCSS(tt.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(styleSheet.Rules) == 0 || len(styleSheet.Rules[0].Selectors) == 0 {
				t.Fatal("expected at least one rule with one selector")
			}

			sel := styleSheet.Rules[0].Selectors[0]
			if sel.ZoomMin != tt.wantMin {
				t.Errorf("got ZoomMin %d, want %d", sel.ZoomMin, tt.wantMin)
			}

			if sel.ZoomMax != tt.wantMax {
				t.Errorf("got ZoomMax %d, want %d", sel.ZoomMax, tt.wantMax)
			}
		})
	}
}

func TestParseMapCSSDescendantSelector(t *testing.T) {
	t.Parallel()

	input := "relation[type=route] way[highway] { color: red; }"

	styleSheet, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(styleSheet.Rules) == 0 || len(styleSheet.Rules[0].Selectors) == 0 {
		t.Fatal("expected at least one rule with one selector")
	}

	sel := styleSheet.Rules[0].Selectors[0]
	if sel.Type != "way" {
		t.Errorf("got type %q, want %q", sel.Type, "way")
	}

	if sel.Parent == nil {
		t.Fatal("expected parent selector")
	}

	if sel.Parent.Type != "relation" {
		t.Errorf("got parent type %q, want %q", sel.Parent.Type, "relation")
	}
}

func TestParseMapCSSQuotedStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantKey string
		wantVal string
	}{
		{
			name:    "quoted key and value",
			input:   `way["highway"="primary"] { color: red; }`,
			wantKey: "highway",
			wantVal: "primary",
		},
		{
			name:    "single quoted value",
			input:   `way[name='Cafe'] { color: blue; }`,
			wantKey: "name",
			wantVal: "Cafe",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss, err := ParseMapCSS(tt.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(ss.Rules) == 0 || len(ss.Rules[0].Selectors) == 0 {
				t.Fatal("expected at least one rule with one selector")
			}

			sel := ss.Rules[0].Selectors[0]
			if len(sel.Conditions) == 0 {
				t.Fatal("expected at least one condition")
			}

			cond := sel.Conditions[0]
			if cond.Key != tt.wantKey {
				t.Errorf("got key %q, want %q", cond.Key, tt.wantKey)
			}

			if cond.Value != tt.wantVal {
				t.Errorf("got value %q, want %q", cond.Value, tt.wantVal)
			}
		})
	}
}

func TestParseMapCSSComplexStylesheet(t *testing.T) {
	t.Parallel()

	input := `
		/* Overpass Turbo MapCSS example */
		node[amenity=cafe] {
			icon-image: url('icons/maki/cafe-18.png');
			icon-width: 18;
			icon-height: 18;
		}

		way[highway=primary] {
			color: #ff0000;
			width: 8;
			casing-width: 2;
			casing-color: #660000;
		}

		way[highway=secondary], way[highway=tertiary] {
			color: #ffcc00;
			width: 6;
		}

		area[building] {
			fill-color: rgba(0.5, 0.5, 0.5, 0.3);
			color: #333333;
			width: 1;
		}

		way:hover {
			color: blue;
		}

		relation[type=route] way {
			color: purple;
			width: 3;
		}
	`

	ss, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(ss.Rules) != 6 {
		t.Errorf("got %d rules, want 6", len(ss.Rules))
	}
}

func TestExpandWithParsedStyle(t *testing.T) {
	t.Parallel()

	query := `[out:json];{{style: node { color: red; } }}node[amenity=cafe]({{bbox}});out;`
	opts := Options{
		BBox: &BBox{South: 48.0, West: 16.0, North: 49.0, East: 17.0},
	}

	res, err := Expand(query, opts)
	if err != nil {
		t.Fatalf("Expand() error = %v", err)
	}

	if len(res.Styles) != 1 {
		t.Fatalf("got %d styles, want 1", len(res.Styles))
	}

	if len(res.ParsedStyles) != 1 {
		t.Fatalf("got %d parsed styles, want 1", len(res.ParsedStyles))
	}

	if res.ParsedStyles[0] == nil {
		t.Fatal("expected non-nil parsed style")
	}

	if len(res.ParsedStyles[0].Rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(res.ParsedStyles[0].Rules))
	}
}

func TestColorHex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		color   Color
		wantHex string
	}{
		{Color{1, 0, 0, 1}, "#ff0000"},
		{Color{0, 1, 0, 1}, "#00ff00"},
		{Color{0, 0, 1, 1}, "#0000ff"},
		{Color{1, 1, 1, 1}, "#ffffff"},
		{Color{0, 0, 0, 1}, "#000000"},
		{Color{1, 0, 0, 0.5}, "#ff00007f"},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.wantHex, func(t *testing.T) {
			t.Parallel()

			got := tt.color.Hex()
			if got != tt.wantHex {
				t.Errorf("got %q, want %q", got, tt.wantHex)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unclosed brace",
			input: "way { color: red;",
		},
		{
			name:  "unclosed bracket",
			input: "way[highway { color: red; }",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseMapCSS(tt.input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}

	return x
}

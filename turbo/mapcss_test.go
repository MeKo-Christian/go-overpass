package turbo

import (
	"testing"
)

func testParseMapCSSBasicSelectorHelper(t *testing.T, input, wantType string) {
	t.Helper()

	stylesheet, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(stylesheet.Rules) != 1 {
		t.Errorf("got %d rules, want 1", len(stylesheet.Rules))
	}

	if len(stylesheet.Rules) > 0 && len(stylesheet.Rules[0].Selectors) > 0 {
		if stylesheet.Rules[0].Selectors[0].Type != wantType {
			t.Errorf("got type %q, want %q", stylesheet.Rules[0].Selectors[0].Type, wantType)
		}
	}
}

func TestParseMapCSSBasicSelector(t *testing.T) {
	t.Parallel()

	t.Run("node selector", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSBasicSelectorHelper(t, "node { color: red; }", "node")
	})

	t.Run("way selector", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSBasicSelectorHelper(t, "way { color: blue; }", "way")
	})

	t.Run("relation selector", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSBasicSelectorHelper(t, "relation { color: green; }", "relation")
	})

	t.Run("area selector", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSBasicSelectorHelper(t, "area { fill-color: yellow; }", "area")
	})

	t.Run("line selector", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSBasicSelectorHelper(t, "line { width: 2; }", "line")
	})

	t.Run("wildcard selector", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSBasicSelectorHelper(t, "* { opacity: 0.5; }", "*")
	})
}

func testParseMapCSSConditionHelper(t *testing.T, input, wantKey, wantOp, wantVal string) {
	t.Helper()

	stylesheet, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(stylesheet.Rules) == 0 || len(stylesheet.Rules[0].Selectors) == 0 {
		t.Fatal("expected at least one rule with one selector")
	}

	sel := stylesheet.Rules[0].Selectors[0]
	if len(sel.Conditions) == 0 {
		t.Fatal("expected at least one condition")
	}

	cond := sel.Conditions[0]
	if cond.Key != wantKey {
		t.Errorf("got key %q, want %q", cond.Key, wantKey)
	}

	if cond.Operator != wantOp {
		t.Errorf("got operator %q, want %q", cond.Operator, wantOp)
	}

	if cond.Value != wantVal {
		t.Errorf("got value %q, want %q", cond.Value, wantVal)
	}
}

func TestParseMapCSSConditions(t *testing.T) {
	t.Parallel()

	t.Run("equals", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSConditionHelper(t, "way[highway=primary] { color: red; }", "highway", "=", "primary")
	})

	t.Run("not equals", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSConditionHelper(t, "way[highway!=motorway] { color: blue; }", "highway", "!=", "motorway")
	})

	t.Run("exists", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSConditionHelper(t, "way[name] { color: green; }", "name", "", "")
	})

	t.Run("not exists", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSConditionHelper(t, "way[!name] { color: gray; }", "name", "!", "")
	})

	t.Run("regex match", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSConditionHelper(t, `way[highway=~/.*ary/] { color: purple; }`, "highway", "=~", "/.*ary/")
	})

	t.Run("less than", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSConditionHelper(t, "node[population<1000] { color: yellow; }", "population", "<", "1000")
	})

	t.Run("greater than or equal", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSConditionHelper(t, "node[population>=1000000] { color: orange; }", "population", ">=", "1000000")
	})

	t.Run("meta attribute", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSConditionHelper(t, "way[@id=171784106] { color: cyan; }", "@id", "=", "171784106")
	})
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

	for _, testCase := range tests {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			stylesheet, err := ParseMapCSS(testCase.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(stylesheet.Rules) == 0 || len(stylesheet.Rules[0].Declarations) == 0 {
				t.Fatal("expected at least one rule with one declaration")
			}

			decl := stylesheet.Rules[0].Declarations[0]
			if decl.Value.Type != ValueTypeColor {
				t.Fatalf("expected color type, got %v", decl.Value.Type)
			}

			if decl.Value.Color == nil {
				t.Fatal("expected non-nil color")
			}

			color := decl.Value.Color
			if abs(color.R-testCase.wantR) > testCase.tolerance {
				t.Errorf("R: got %.3f, want %.3f", color.R, testCase.wantR)
			}

			if abs(color.G-testCase.wantG) > testCase.tolerance {
				t.Errorf("G: got %.3f, want %.3f", color.G, testCase.wantG)
			}

			if abs(color.B-testCase.wantB) > testCase.tolerance {
				t.Errorf("B: got %.3f, want %.3f", color.B, testCase.wantB)
			}

			if abs(color.A-testCase.wantA) > testCase.tolerance {
				t.Errorf("A: got %.3f, want %.3f", color.A, testCase.wantA)
			}
		})
	}
}

func testParseMapCSSDeclarationHelper(t *testing.T, input, wantProp string, wantType ValueType) {
	t.Helper()

	stylesheet, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(stylesheet.Rules) == 0 || len(stylesheet.Rules[0].Declarations) == 0 {
		t.Fatal("expected at least one rule with one declaration")
	}

	decl := stylesheet.Rules[0].Declarations[0]
	if decl.Property != wantProp {
		t.Errorf("got property %q, want %q", decl.Property, wantProp)
	}

	if decl.Value.Type != wantType {
		t.Errorf("got type %v, want %v", decl.Value.Type, wantType)
	}
}

func TestParseMapCSSDeclarations(t *testing.T) {
	t.Parallel()

	t.Run("number value", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSDeclarationHelper(t, "way { width: 5; }", "width", ValueTypeNumber)
	})

	t.Run("url value", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSDeclarationHelper(t, "node { icon-image: url('icons/maki/cafe-18.png'); }", "icon-image", ValueTypeURL)
	})

	t.Run("dashes value", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSDeclarationHelper(t, "way { dashes: 5, 8; }", "dashes", ValueTypeDashes)
	})

	t.Run("keyword value", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSDeclarationHelper(t, "way { linecap: round; }", "linecap", ValueTypeKeyword)
	})

	t.Run("eval value", func(t *testing.T) {
		t.Parallel()
		testParseMapCSSDeclarationHelper(t, `node { opacity: eval("tag('population')/100000"); }`, "opacity", ValueTypeEval)
	})
}

func TestParseMapCSSMultipleSelectors(t *testing.T) {
	t.Parallel()

	input := "way[highway=primary], way[highway=secondary] { color: red; }"

	stylesheet, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(stylesheet.Rules))
	}

	if len(stylesheet.Rules[0].Selectors) != 2 {
		t.Fatalf("got %d selectors, want 2", len(stylesheet.Rules[0].Selectors))
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

	for _, testCase := range tests {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			stylesheet, err := ParseMapCSS(testCase.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(stylesheet.Rules) != 1 {
				t.Fatalf("got %d rules, want 1", len(stylesheet.Rules))
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

	for _, testCase := range tests {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			stylesheet, err := ParseMapCSS(testCase.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(stylesheet.Rules) == 0 || len(stylesheet.Rules[0].Declarations) == 0 {
				t.Fatal("expected at least one rule with one declaration")
			}

			decl := stylesheet.Rules[0].Declarations[0]
			if decl.Property != testCase.wantProp {
				t.Errorf("got property %q, want %q", decl.Property, testCase.wantProp)
			}

			if decl.Value.Raw != testCase.wantVal {
				t.Errorf("got value %q, want %q", decl.Value.Raw, testCase.wantVal)
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

	for _, testCase := range tests {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			stylesheet, err := ParseMapCSS(testCase.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(stylesheet.Rules) == 0 || len(stylesheet.Rules[0].Selectors) == 0 {
				t.Fatal("expected at least one rule with one selector")
			}

			sel := stylesheet.Rules[0].Selectors[0]
			if len(sel.PseudoClasses) != len(testCase.pseudo) {
				t.Fatalf("got %d pseudo-classes, want %d", len(sel.PseudoClasses), len(testCase.pseudo))
			}

			for i, p := range testCase.pseudo {
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

	stylesheet, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(stylesheet.Rules) == 0 || len(stylesheet.Rules[0].Selectors) == 0 {
		t.Fatal("expected at least one rule with one selector")
	}

	sel := stylesheet.Rules[0].Selectors[0]
	if len(sel.Classes) != 1 || sel.Classes[0] != "minor_road" {
		t.Errorf("got classes %v, want [minor_road]", sel.Classes)
	}
}

func TestParseMapCSSLayer(t *testing.T) {
	t.Parallel()

	input := "way::casing { width: 10; }"

	stylesheet, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(stylesheet.Rules) == 0 || len(stylesheet.Rules[0].Selectors) == 0 {
		t.Fatal("expected at least one rule with one selector")
	}

	sel := stylesheet.Rules[0].Selectors[0]
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

	for _, testCase := range tests {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			styleSheet, err := ParseMapCSS(testCase.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(styleSheet.Rules) == 0 || len(styleSheet.Rules[0].Selectors) == 0 {
				t.Fatal("expected at least one rule with one selector")
			}

			sel := styleSheet.Rules[0].Selectors[0]
			if sel.ZoomMin != testCase.wantMin {
				t.Errorf("got ZoomMin %d, want %d", sel.ZoomMin, testCase.wantMin)
			}

			if sel.ZoomMax != testCase.wantMax {
				t.Errorf("got ZoomMax %d, want %d", sel.ZoomMax, testCase.wantMax)
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

	for _, testCase := range tests {
		testCase := testCase // capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			stylesheet, err := ParseMapCSS(testCase.input)
			if err != nil {
				t.Fatalf("ParseMapCSS() error = %v", err)
			}

			if len(stylesheet.Rules) == 0 || len(stylesheet.Rules[0].Selectors) == 0 {
				t.Fatal("expected at least one rule with one selector")
			}

			sel := stylesheet.Rules[0].Selectors[0]
			if len(sel.Conditions) == 0 {
				t.Fatal("expected at least one condition")
			}

			cond := sel.Conditions[0]
			if cond.Key != testCase.wantKey {
				t.Errorf("got key %q, want %q", cond.Key, testCase.wantKey)
			}

			if cond.Value != testCase.wantVal {
				t.Errorf("got value %q, want %q", cond.Value, testCase.wantVal)
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

	stylesheet, err := ParseMapCSS(input)
	if err != nil {
		t.Fatalf("ParseMapCSS() error = %v", err)
	}

	if len(stylesheet.Rules) != 6 {
		t.Errorf("got %d rules, want 6", len(stylesheet.Rules))
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

	for _, testCase := range tests {
		testCase := testCase // capture range variable
		t.Run(testCase.wantHex, func(t *testing.T) {
			t.Parallel()

			got := testCase.color.Hex()
			if got != testCase.wantHex {
				t.Errorf("got %q, want %q", got, testCase.wantHex)
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

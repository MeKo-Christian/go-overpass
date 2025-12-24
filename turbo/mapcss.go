package turbo

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Stylesheet represents a parsed MapCSS stylesheet.
type Stylesheet struct {
	Rules []Rule
}

// Rule represents a single MapCSS rule with selectors and declarations.
type Rule struct {
	Selectors    []Selector
	Declarations []Declaration
}

// Selector represents a MapCSS selector.
type Selector struct {
	// Type is the object type: node, way, relation, area, line, canvas, meta, or * for any.
	Type string
	// Layer is the optional layer name (e.g., "::casing").
	Layer string
	// ZoomMin is the minimum zoom level (0 if unspecified).
	ZoomMin int
	// ZoomMax is the maximum zoom level (0 if unspecified, meaning unlimited).
	ZoomMax int
	// Conditions are tag filters like [highway=primary].
	Conditions []Condition
	// PseudoClasses are pseudo-class selectors like :hover, :active, :tagged.
	PseudoClasses []string
	// Classes are class selectors like .minor_road.
	Classes []string
	// Parent is an optional parent selector for descendant rules.
	Parent *Selector
}

// Condition represents a tag filter condition in a selector.
type Condition struct {
	Key      string
	Operator string // =, !=, =~, <, >, <=, >=, empty string for exists, ! for not exists
	Value    string
	Regex    *regexp.Regexp // compiled regex for =~ operator
}

// Declaration represents a property-value pair in a MapCSS rule.
type Declaration struct {
	Property string
	Value    Value
}

// Value represents a MapCSS property value.
type Value struct {
	// Raw is the original string value.
	Raw string
	// Type indicates the value type.
	Type ValueType
	// Parsed values for specific types.
	Color   *Color
	Number  float64
	URL     string
	Eval    string
	Dashes  []float64
	Strings []string
}

// ValueType indicates the type of a MapCSS value.
type ValueType int

const (
	ValueTypeString ValueType = iota
	ValueTypeNumber
	ValueTypeColor
	ValueTypeURL
	ValueTypeEval
	ValueTypeDashes
	ValueTypeKeyword
)

// Color represents an RGBA color.
type Color struct {
	R, G, B, A float64
}

// ParseError represents a MapCSS parsing error.
type ParseError struct {
	Line    int
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("mapcss: line %d, col %d: %s", e.Line, e.Column, e.Message)
	}

	return "mapcss: " + e.Message
}

// ErrInvalidHexColor is returned when an invalid hex color is encountered.
var (
	ErrInvalidHexColor      = errors.New("invalid hex color")
	ErrEndOfInput           = errors.New("end of input")
	ErrInvalidSelectorStart = errors.New("invalid selector start")
	ErrEmptySelector        = errors.New("empty selector")
	ErrEmptyDeclaration     = errors.New("empty declaration")
)

// ParseMapCSS parses a MapCSS stylesheet string into a Stylesheet structure.
func ParseMapCSS(input string) (*Stylesheet, error) {
	parser := &parser{
		input: input,
		pos:   0,
		line:  1,
		col:   1,
	}

	return parser.parse()
}

type parser struct {
	input string
	pos   int
	line  int
	col   int
}

func (p *parser) parse() (*Stylesheet, error) {
	var rules []Rule

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()

		if p.pos >= len(p.input) {
			break
		}

		// Skip @import statements (not fully supported)
		if p.peek() == '@' {
			p.skipAtRule()
			continue
		}

		rule, err := p.parseRule()
		if err != nil {
			return nil, err
		}

		if rule != nil {
			rules = append(rules, *rule)
		}
	}

	return &Stylesheet{Rules: rules}, nil
}

func (p *parser) parseRule() (*Rule, error) {
	selectors, err := p.parseSelectors()
	if err != nil {
		return nil, err
	}

	if len(selectors) == 0 {
		return nil, p.error("no selectors found")
	}

	p.skipWhitespaceAndComments()

	if p.pos >= len(p.input) || p.peek() != '{' {
		return nil, p.error("expected '{'")
	}

	p.advance()

	declarations, err := p.parseDeclarations()
	if err != nil {
		return nil, err
	}

	p.skipWhitespaceAndComments()

	if p.pos >= len(p.input) || p.peek() != '}' {
		return nil, p.error("expected '}'")
	}

	p.advance()

	return &Rule{
		Selectors:    selectors,
		Declarations: declarations,
	}, nil
}

func (p *parser) parseSelectors() ([]Selector, error) {
	var selectors []Selector

	for {
		p.skipWhitespaceAndComments()

		if p.pos >= len(p.input) {
			break
		}

		if p.peek() == '{' {
			break
		}

		sel, err := p.parseSelector()
		if err != nil {
			return nil, err
		}

		if sel != nil {
			selectors = append(selectors, *sel)
		}

		p.skipWhitespaceAndComments()

		if p.pos < len(p.input) && p.peek() == ',' {
			p.advance()
			continue
		}

		break
	}

	return selectors, nil
}

func (p *parser) shouldContinueParsing(ch byte) bool {
	return ch != '{' && ch != ',' && isTypeStart(ch)
}

func (p *parser) parseSelector() (*Selector, error) {
	p.skipWhitespaceAndComments()

	if p.pos >= len(p.input) {
		return nil, p.error("unexpected end of input")
	}

	selectors, err := p.parseSelectorChain()
	if err != nil {
		return nil, err
	}

	if len(selectors) == 0 {
		return nil, p.error("no selector parsed")
	}

	// Link selectors as parent chain (last is the main selector)
	result := selectors[len(selectors)-1]
	for i := len(selectors) - 2; i >= 0; i-- {
		selectors[i+1].Parent = selectors[i]
	}

	return result, nil
}

func (p *parser) parseSelectorChain() ([]*Selector, error) {
	var selectors []*Selector

	for {
		sel, err := p.parseSingleSelector()
		if err != nil {
			// End of input, invalid selector start, or empty selector are normal loop termination conditions
			if errors.Is(err, ErrEndOfInput) || errors.Is(err, ErrInvalidSelectorStart) || errors.Is(err, ErrEmptySelector) {
				break
			}

			return nil, err
		}

		selectors = append(selectors, sel)

		p.skipWhitespaceAndComments()
		// Check for descendant selector (space before next type)
		if p.pos < len(p.input) && p.shouldContinueParsing(p.peek()) {
			continue
		}

		break
	}

	return selectors, nil
}

func (p *parser) parseSingleSelector() (*Selector, error) {
	p.skipWhitespaceAndComments()

	if p.pos >= len(p.input) {
		return nil, ErrEndOfInput
	}

	char := p.peek()
	if char == '{' || char == ',' {
		return nil, ErrInvalidSelectorStart
	}

	sel := &Selector{}

	err := p.parseSelectorType(sel)
	if err != nil {
		return nil, err
	}

	p.parseSelectorLayer(sel)
	p.parseSelectorZoomRange(sel)

	err = p.parseSelectorModifiers(sel)
	if err != nil {
		return nil, err
	}

	if sel.Type == "" && len(sel.Conditions) == 0 && len(sel.PseudoClasses) == 0 && len(sel.Classes) == 0 {
		return nil, ErrEmptySelector
	}

	return sel, nil
}

func (p *parser) parseSelectorType(sel *Selector) error {
	char := p.peek()

	switch {
	case char == '*':
		sel.Type = "*"

		p.advance()
	case isLetter(char):
		sel.Type = p.parseIdent()
	case char != '[' && char != ':' && char != '.' && char != '|':
		return ErrInvalidSelectorStart
	}

	return nil
}

func (p *parser) parseSelectorLayer(sel *Selector) {
	if p.pos+1 < len(p.input) && p.input[p.pos:p.pos+2] == "::" {
		p.advance()
		p.advance()
		sel.Layer = p.parseIdent()
	}
}

func (p *parser) parseSelectorZoomRange(sel *Selector) {
	if p.pos >= len(p.input) || p.peek() != '|' {
		return
	}

	p.advance()

	if p.pos >= len(p.input) || p.peek() != 'z' {
		return
	}

	p.advance()
	p.parseZoomValues(sel)
}

func (p *parser) parseZoomValues(sel *Selector) {
	zoomStr := p.parseNumber()

	if strings.Contains(zoomStr, "-") {
		p.parseZoomRange(sel, zoomStr)
		return
	}

	zoom, _ := strconv.Atoi(zoomStr)
	sel.ZoomMin = zoom
	sel.ZoomMax = zoom

	// Check for range like |z12-
	if p.pos < len(p.input) && p.peek() == '-' {
		p.advance()

		maxStr := p.parseNumber()
		if maxStr != "" {
			sel.ZoomMax, _ = strconv.Atoi(maxStr)
		} else {
			sel.ZoomMax = 0 // unlimited
		}
	}
}

func (p *parser) parseZoomRange(sel *Selector, zoomStr string) {
	parts := strings.SplitN(zoomStr, "-", 2)
	if len(parts) == 2 {
		sel.ZoomMin, _ = strconv.Atoi(parts[0])
		if parts[1] != "" {
			sel.ZoomMax, _ = strconv.Atoi(parts[1])
		}
	}
}

func (p *parser) parseSelectorModifiers(sel *Selector) error {
	for p.pos < len(p.input) {
		char := p.peek()

		switch {
		case char == '[':
			err := p.parseSelectorCondition(sel)
			if err != nil {
				return err
			}
		case char == ':' && (p.pos+1 >= len(p.input) || p.input[p.pos+1] != ':'):
			p.parseSelectorPseudoClass(sel)
		case char == '.':
			p.parseSelectorClass(sel)
		default:
			return nil
		}
	}

	return nil
}

func (p *parser) parseSelectorCondition(sel *Selector) error {
	cond, err := p.parseCondition()
	if err != nil {
		return err
	}

	sel.Conditions = append(sel.Conditions, *cond)

	return nil
}

func (p *parser) parseSelectorPseudoClass(sel *Selector) {
	p.advance()

	pseudo := p.parseIdent()
	if pseudo != "" {
		sel.PseudoClasses = append(sel.PseudoClasses, pseudo)
	}
}

func (p *parser) parseSelectorClass(sel *Selector) {
	p.advance()

	class := p.parseIdent()
	if class != "" {
		sel.Classes = append(sel.Classes, class)
	}
}

func (p *parser) compileConditionRegex(operator, value string) (*regexp.Regexp, error) {
	if operator != "=~" && operator != "!~" {
		return nil, nil //nolint:nilnil // not an error condition, regex only used for =~ and !~
	}

	pattern := value
	// Remove surrounding slashes if present
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		pattern = pattern[1 : len(pattern)-1]
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, p.error(fmt.Sprintf("invalid regex: %s", err))
	}

	return re, nil
}

func (p *parser) parseCondition() (*Condition, error) {
	if p.peek() != '[' {
		return nil, p.error("expected '['")
	}

	p.advance()

	p.skipWhitespace()

	cond := &Condition{}

	// Check for negation
	if p.pos < len(p.input) && p.peek() == '!' {
		p.advance()

		cond.Operator = "!"
	}

	// Parse key (may be quoted or include @)
	key := p.parseKeyOrString()
	cond.Key = key

	p.skipWhitespace()

	// Parse operator and value
	if p.pos < len(p.input) && p.peek() != ']' {
		operator := p.parseOperator()
		if operator != "" {
			cond.Operator = operator

			p.skipWhitespace()
			cond.Value = p.parseValueString()

			// Compile regex for =~ operator
			re, err := p.compileConditionRegex(operator, cond.Value)
			if err != nil {
				return nil, err
			}

			cond.Regex = re
		}
	}

	p.skipWhitespace()

	if p.pos >= len(p.input) || p.peek() != ']' {
		return nil, p.error("expected ']'")
	}

	p.advance()

	return cond, nil
}

func (p *parser) parseDeclarations() ([]Declaration, error) {
	var decls []Declaration

	for {
		p.skipWhitespaceAndComments()

		if p.pos >= len(p.input) || p.peek() == '}' {
			break
		}

		decl, err := p.parseDeclaration()
		if err != nil {
			// Empty declaration is a normal skip condition
			if errors.Is(err, ErrEmptyDeclaration) {
				continue
			}

			return nil, err
		}

		decls = append(decls, *decl)
	}

	return decls, nil
}

func (p *parser) parseSetDeclaration() (*Declaration, error) {
	p.pos += 3 // skip "set"
	p.skipWhitespace()

	// set .className
	if p.pos < len(p.input) && p.peek() == '.' {
		p.advance()
		className := p.parseIdent()
		p.skipWhitespace()

		if p.pos < len(p.input) && p.peek() == ';' {
			p.advance()
		}

		return &Declaration{
			Property: "set-class",
			Value:    Value{Raw: className, Type: ValueTypeString},
		}, nil
	}

	// set tag=value or set tag
	tagName := p.parseIdent()
	p.skipWhitespace()

	value := "yes"

	if p.pos < len(p.input) && p.peek() == '=' {
		p.advance()
		p.skipWhitespace()
		value = p.parseValueString()
	}

	p.skipWhitespace()

	if p.pos < len(p.input) && p.peek() == ';' {
		p.advance()
	}

	return &Declaration{
		Property: "set-tag:" + tagName,
		Value:    Value{Raw: value, Type: ValueTypeString},
	}, nil
}

func (p *parser) parseDeclaration() (*Declaration, error) {
	p.skipWhitespaceAndComments()

	// Handle 'set' directive
	if p.pos+3 < len(p.input) && p.input[p.pos:p.pos+3] == "set" {
		return p.parseSetDeclaration()
	}

	// Parse property name
	prop := p.parseIdent()
	if prop == "" {
		return nil, ErrEmptyDeclaration
	}

	p.skipWhitespace()

	if p.pos >= len(p.input) || p.peek() != ':' {
		return nil, p.error("expected ':'")
	}

	p.advance()

	p.skipWhitespace()

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()

	if p.pos < len(p.input) && p.peek() == ';' {
		p.advance()
	}

	return &Declaration{
		Property: prop,
		Value:    *value,
	}, nil
}

func (p *parser) parseURLValue() (*Value, error) {
	p.pos += 4 // skip "url("
	content := p.parseUntilClosingParen()

	url := strings.Trim(content, `'"`)

	return &Value{
		Raw:  "url(" + content + ")",
		Type: ValueTypeURL,
		URL:  url,
	}, nil
}

func (p *parser) parseEvalValue() (*Value, error) {
	p.pos += 5 // skip "eval("
	content := p.parseUntilClosingParen()

	expr := strings.Trim(content, `'"`)

	return &Value{
		Raw:  "eval(" + content + ")",
		Type: ValueTypeEval,
		Eval: expr,
	}, nil
}

func (p *parser) determineValueType(rawStr string) *Value {
	val := &Value{Raw: rawStr}

	// Try to determine type
	if color := parseNamedColor(rawStr); color != nil {
		val.Type = ValueTypeColor
		val.Color = color

		return val
	}

	num, err := strconv.ParseFloat(rawStr, 64)
	if err == nil {
		val.Type = ValueTypeNumber
		val.Number = num

		return val
	}

	if strings.Contains(rawStr, ",") && !strings.ContainsAny(rawStr, "()") {
		// Might be dashes pattern
		parts := strings.Split(rawStr, ",")
		var dashes []float64
		allNumeric := true

		for _, part := range parts {
			num, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
			if err != nil {
				allNumeric = false
				break
			}

			dashes = append(dashes, num)
		}

		if allNumeric {
			val.Type = ValueTypeDashes
			val.Dashes = dashes

			return val
		}
	}

	val.Type = ValueTypeKeyword

	return val
}

func (p *parser) parseValue() (*Value, error) { //nolint:cyclop // multiple value type checks needed
	p.skipWhitespace()

	if p.pos >= len(p.input) {
		return &Value{}, nil
	}

	startPos := p.pos

	// Check for special functions
	if p.pos+4 < len(p.input) && p.input[p.pos:p.pos+4] == "url(" {
		return p.parseURLValue()
	}

	if p.pos+5 < len(p.input) && p.input[p.pos:p.pos+5] == "eval(" {
		return p.parseEvalValue()
	}

	if p.pos+4 < len(p.input) && p.input[p.pos:p.pos+4] == "rgb(" {
		return p.parseRGBColor()
	}

	if p.pos+5 < len(p.input) && p.input[p.pos:p.pos+5] == "rgba(" {
		return p.parseRGBAColor()
	}

	// Hex color
	if p.peek() == '#' {
		return p.parseHexColor()
	}

	// Collect until ; or }
	var raw strings.Builder

	for p.pos < len(p.input) {
		ch := p.peek()
		if ch == ';' || ch == '}' {
			break
		}

		raw.WriteByte(ch)
		p.advance()
	}

	rawStr := strings.TrimSpace(raw.String())
	val := p.determineValueType(rawStr)

	// Restore pos if we didn't consume anything
	if p.pos == startPos {
		return val, nil
	}

	return val, nil
}

func (p *parser) parseRGBColor() (*Value, error) {
	p.pos += 4 // skip "rgb("
	content := p.parseUntilClosingParen()

	parts := strings.Split(content, ",")
	if len(parts) != 3 {
		return nil, p.error("rgb() requires 3 values")
	}

	red, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	green, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	blue, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)

	return &Value{
		Raw:   fmt.Sprintf("rgb(%s)", content),
		Type:  ValueTypeColor,
		Color: &Color{R: red, G: green, B: blue, A: 1.0},
	}, nil
}

func (p *parser) parseRGBAColor() (*Value, error) {
	p.pos += 5 // skip "rgba("
	content := p.parseUntilClosingParen()

	parts := strings.Split(content, ",")
	if len(parts) != 4 {
		return nil, p.error("rgba() requires 4 values")
	}

	red, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	green, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	blue, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	alpha, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)

	return &Value{
		Raw:   fmt.Sprintf("rgba(%s)", content),
		Type:  ValueTypeColor,
		Color: &Color{R: red, G: green, B: blue, A: alpha},
	}, nil
}

func (p *parser) parseHexColor() (*Value, error) {
	if p.peek() != '#' {
		return nil, p.error("expected '#'")
	}

	p.advance()

	var hex strings.Builder
	hex.WriteByte('#')

	for p.pos < len(p.input) && isHexDigit(p.peek()) {
		hex.WriteByte(p.peek())
		p.advance()
	}

	hexStr := hex.String()

	color, err := parseHexColorValue(hexStr)
	if err != nil {
		return nil, p.error(err.Error())
	}

	return &Value{
		Raw:   hexStr,
		Type:  ValueTypeColor,
		Color: color,
	}, nil
}

func (p *parser) parseUntilClosingParen() string {
	var buf strings.Builder

	depth := 1
	for p.pos < len(p.input) && depth > 0 {
		char := p.peek()
		if char == '(' {
			depth++
		} else if char == ')' {
			depth--
			if depth == 0 {
				p.advance()
				break
			}
		}

		buf.WriteByte(char)
		p.advance()
	}

	return buf.String()
}

func (p *parser) parseIdent() string {
	var buf strings.Builder

	for p.pos < len(p.input) {
		ch := p.peek()
		if isIdent(ch) {
			buf.WriteByte(ch)
			p.advance()
		} else {
			break
		}
	}

	return buf.String()
}

func (p *parser) parseNumber() string {
	var buf strings.Builder

	for p.pos < len(p.input) {
		ch := p.peek()
		if isDigit(ch) || ch == '-' || ch == '.' {
			buf.WriteByte(ch)
			p.advance()
		} else {
			break
		}
	}

	return buf.String()
}

func (p *parser) parseKeyOrString() string {
	p.skipWhitespace()

	if p.pos >= len(p.input) {
		return ""
	}

	ch := p.peek()
	if ch == '"' || ch == '\'' {
		return p.parseQuotedString()
	}

	return p.parseIdent()
}

func (p *parser) parseQuotedString() string {
	if p.pos >= len(p.input) {
		return ""
	}

	quote := p.peek()
	if quote != '"' && quote != '\'' {
		return p.parseIdent()
	}

	p.advance()

	var buf strings.Builder

loop2:
	for p.pos < len(p.input) {
		char := p.peek()
		switch {
		case char == '\\' && p.pos+1 < len(p.input):
			p.advance()
			buf.WriteByte(p.peek())
			p.advance()
		case char == quote:
			p.advance()
			break loop2
		default:
			buf.WriteByte(char)
			p.advance()
		}
	}

	return buf.String()
}

func (p *parser) parseValueString() string {
	p.skipWhitespace()

	if p.pos >= len(p.input) {
		return ""
	}

	char := p.peek()
	if char == '"' || char == '\'' {
		return p.parseQuotedString()
	}

	// Regex pattern
	if char == '/' {
		return p.parseRegex()
	}

	var buf strings.Builder

	for p.pos < len(p.input) {
		ch := p.peek()
		if ch == ']' || ch == ';' || ch == '}' || isWhitespace(ch) {
			break
		}

		buf.WriteByte(ch)
		p.advance()
	}

	return buf.String()
}

func (p *parser) parseRegex() string {
	if p.peek() != '/' {
		return ""
	}

	p.advance()

	var buf strings.Builder
	buf.WriteByte('/')

loop3:
	for p.pos < len(p.input) {
		char := p.peek()
		switch {
		case char == '\\' && p.pos+1 < len(p.input):
			buf.WriteByte(char)
			p.advance()
			buf.WriteByte(p.peek())
			p.advance()
		case char == '/':
			buf.WriteByte(char)
			p.advance()

			break loop3
		default:
			buf.WriteByte(char)
			p.advance()
		}
	}

	return buf.String()
}

func (p *parser) parseOperator() string {
	if p.pos >= len(p.input) {
		return ""
	}

	// Two-char operators
	if p.pos+1 < len(p.input) {
		two := p.input[p.pos : p.pos+2]
		switch two {
		case "=~", "!~", "!=", "<=", ">=":
			p.pos += 2
			return two
		}
	}

	// Single-char operators
	ch := p.peek()
	switch ch {
	case '=', '<', '>':
		p.advance()
		return string(ch)
	}

	return ""
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && isWhitespace(p.peek()) {
		p.advance()
	}
}

func (p *parser) skipBlockComment() {
	// Skip /* ... */
	p.pos += 2 // skip /*
	for p.pos+1 < len(p.input) {
		if p.input[p.pos] == '*' && p.input[p.pos+1] == '/' {
			p.pos += 2
			break
		}

		p.advance()
	}
}

func (p *parser) skipLineComment() {
	// Skip // ... to end of line
	for p.pos < len(p.input) && p.peek() != '\n' {
		p.advance()
	}
}

func (p *parser) skipWhitespaceAndComments() {
	for p.pos < len(p.input) {
		char := p.peek()
		if isWhitespace(char) {
			p.advance()
			continue
		}

		// Check for comments
		if char == '/' && p.pos+1 < len(p.input) {
			nextChar := p.input[p.pos+1]
			if nextChar == '*' {
				p.skipBlockComment()
				continue
			}

			if nextChar == '/' {
				p.skipLineComment()
				continue
			}
		}

		break
	}
}

//nolint:nestif
func (p *parser) skipAtRule() {
	// Skip @import or other @ rules
	for p.pos < len(p.input) && p.peek() != ';' && p.peek() != '{' {
		p.advance()
	}

	if p.pos < len(p.input) {
		if p.peek() == '{' {
			// Skip block
			p.advance()

			depth := 1
			for p.pos < len(p.input) && depth > 0 {
				if p.peek() == '{' {
					depth++
				} else if p.peek() == '}' {
					depth--
				}

				p.advance()
			}
		} else {
			p.advance() // skip ;
		}
	}
}

func (p *parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}

	return p.input[p.pos]
}

func (p *parser) advance() {
	if p.pos < len(p.input) {
		if p.input[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}

		p.pos++
	}
}

func (p *parser) error(msg string) error {
	return &ParseError{
		Line:    p.line,
		Column:  p.col,
		Message: msg,
	}
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdent(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '-' || ch == '_' || ch == '@'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isTypeStart(ch byte) bool {
	return isLetter(ch) || ch == '*'
}

func parseHexColorValue(hex string) (*Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	var red, green, blue, alpha float64
	alpha = 1 // default alpha

	switch len(hex) {
	case 3: // #RGB
		red = float64(hexVal(hex[0])*17) / 255
		green = float64(hexVal(hex[1])*17) / 255
		blue = float64(hexVal(hex[2])*17) / 255
	case 4: // #RGBA
		red = float64(hexVal(hex[0])*17) / 255
		green = float64(hexVal(hex[1])*17) / 255
		blue = float64(hexVal(hex[2])*17) / 255
		alpha = float64(hexVal(hex[3])*17) / 255
	case 6: // #RRGGBB
		red = float64(hexVal(hex[0])*16+hexVal(hex[1])) / 255
		green = float64(hexVal(hex[2])*16+hexVal(hex[3])) / 255
		blue = float64(hexVal(hex[4])*16+hexVal(hex[5])) / 255
	case 8: // #RRGGBBAA
		red = float64(hexVal(hex[0])*16+hexVal(hex[1])) / 255
		green = float64(hexVal(hex[2])*16+hexVal(hex[3])) / 255
		blue = float64(hexVal(hex[4])*16+hexVal(hex[5])) / 255
		alpha = float64(hexVal(hex[6])*16+hexVal(hex[7])) / 255
	default:
		return nil, fmt.Errorf("%w: #%s", ErrInvalidHexColor, hex)
	}

	return &Color{R: red, G: green, B: blue, A: alpha}, nil
}

func hexVal(char byte) int {
	switch {
	case char >= '0' && char <= '9':
		return int(char - '0')
	case char >= 'a' && char <= 'f':
		return int(char-'a') + 10
	case char >= 'A' && char <= 'F':
		return int(char-'A') + 10
	}

	return 0
}

// Named CSS colors supported by MapCSS.
var namedColors = map[string]*Color{
	"black":   {0, 0, 0, 1},
	"white":   {1, 1, 1, 1},
	"red":     {1, 0, 0, 1},
	"green":   {0, 0.5, 0, 1},
	"blue":    {0, 0, 1, 1},
	"yellow":  {1, 1, 0, 1},
	"cyan":    {0, 1, 1, 1},
	"magenta": {1, 0, 1, 1},
	"gray":    {0.5, 0.5, 0.5, 1},
	"grey":    {0.5, 0.5, 0.5, 1},
	"orange":  {1, 0.647, 0, 1},
	"purple":  {0.5, 0, 0.5, 1},
	"brown":   {0.647, 0.165, 0.165, 1},
	"pink":    {1, 0.753, 0.796, 1},
	"lime":    {0, 1, 0, 1},
	"navy":    {0, 0, 0.5, 1},
	"teal":    {0, 0.5, 0.5, 1},
	"olive":   {0.5, 0.5, 0, 1},
	"maroon":  {0.5, 0, 0, 1},
	"aqua":    {0, 1, 1, 1},
	"silver":  {0.753, 0.753, 0.753, 1},
	"fuchsia": {1, 0, 1, 1},
}

func parseNamedColor(name string) *Color {
	if c, ok := namedColors[strings.ToLower(name)]; ok {
		return c
	}

	return nil
}

// String returns a CSS-like representation of the color.
func (c *Color) String() string {
	if c.A == 1.0 {
		return fmt.Sprintf("rgb(%.3f, %.3f, %.3f)", c.R, c.G, c.B)
	}

	return fmt.Sprintf("rgba(%.3f, %.3f, %.3f, %.3f)", c.R, c.G, c.B, c.A)
}

// Hex returns the color as a hex string.
func (c *Color) Hex() string {
	red := int(c.R * 255)
	green := int(c.G * 255)
	blue := int(c.B * 255)

	if c.A == 1.0 {
		return fmt.Sprintf("#%02x%02x%02x", red, green, blue)
	}

	alpha := int(c.A * 255)

	return fmt.Sprintf("#%02x%02x%02x%02x", red, green, blue, alpha)
}

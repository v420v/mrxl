package parser

import (
	"fmt"
	"strings"

	"github.com/v420v/mrxl/internal/ast"
)

type Parser struct {
	parsers map[string]diagramParseFunc
}

func NewParser() (*Parser, error) {
	p := &Parser{
		parsers: map[string]diagramParseFunc{
			strings.ToLower(strings.TrimSpace("sequenceDiagram")): parseSequenceDiagram,
			strings.ToLower(strings.TrimSpace("pie")):             parsePieChart,
			strings.ToLower(strings.TrimSpace("timeline")):        parseTimeline,
			strings.ToLower(strings.TrimSpace("quadrantChart")): parseQuadrantChart,
			strings.ToLower(strings.TrimSpace("journey")):        parseUserJourney,
			strings.ToLower(strings.TrimSpace("gantt")):          parseGantt,
		},
	}
	return p, nil
}

type diagramParseFunc func(lines []string) (ast.Diagram, error)

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}

func trimComment(line string) string {
	if i := strings.Index(line, "%%"); i >= 0 {
		line = line[:i]
	}
	return strings.TrimSpace(line)
}

func (p *Parser) Parse(input string) (ast.Diagram, error) {
	lines := normalizedLines(input)
	if len(lines) == 0 {
		return nil, fmt.Errorf("missing diagram header")
	}

	headerLine, body, ok := p.findFirstRegisteredHeader(lines)
	if !ok {
		return nil, fmt.Errorf("missing diagram header")
	}
	return p.parseByHeader(headerLine, body)
}

func normalizedLines(input string) []string {
	raw := splitLines(input)
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		clean := trimComment(strings.TrimSpace(line))
		if clean == "" {
			continue
		}
		lines = append(lines, clean)
	}
	return lines
}

func (p *Parser) parseByHeader(headerLine string, body []string) (ast.Diagram, error) {
	fields := strings.Fields(headerLine)
	header := strings.ToLower(fields[0])
	parseFn, ok := p.parsers[header]
	if !ok {
		return nil, fmt.Errorf("unsupported diagram header %q", headerLine)
	}
	// Inline options after the keyword (e.g. "pie showData title Foo") are
	// prepended to the body so diagram parsers can handle them uniformly.
	if len(fields) > 1 {
		rest := strings.TrimSpace(headerLine[len(fields[0]):])
		body = append([]string{rest}, body...)
	}
	return parseFn(body)
}

func (p *Parser) findFirstRegisteredHeader(lines []string) (headerLine string, body []string, ok bool) {
	for i, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		header := strings.ToLower(fields[0])
		if _, exists := p.parsers[header]; !exists {
			continue
		}
		return line, lines[i+1:], true
	}
	return "", nil, false
}

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
	header := strings.ToLower(strings.Fields(headerLine)[0])
	parseFn, ok := p.parsers[header]
	if !ok {
		return nil, fmt.Errorf("unsupported diagram header %q", headerLine)
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

package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/v420v/mrxl/internal/ast"
)

type quadrantParser struct {
	title     string
	xAxisLow  string
	xAxisHigh string
	yAxisLow  string
	yAxisHigh string
	quadrant1 string
	quadrant2 string
	quadrant3 string
	quadrant4 string
	points    []*ast.QuadrantPoint
}

func newQuadrantParser() *quadrantParser {
	return &quadrantParser{points: make([]*ast.QuadrantPoint, 0)}
}

func parseQuadrantChart(lines []string) (ast.Diagram, error) {
	p := newQuadrantParser()
	for _, line := range lines {
		if err := p.parseLine(line); err != nil {
			return nil, err
		}
	}
	return p.result()
}

func (p *quadrantParser) parseLine(line string) error {
	lower := strings.ToLower(line)

	if strings.HasPrefix(lower, "title ") {
		p.title = strings.TrimSpace(line[len("title "):])
		return nil
	}

	if strings.HasPrefix(lower, "x-axis ") {
		p.xAxisLow, p.xAxisHigh = parseAxisLabels(strings.TrimSpace(line[len("x-axis "):]))
		return nil
	}

	if strings.HasPrefix(lower, "y-axis ") {
		p.yAxisLow, p.yAxisHigh = parseAxisLabels(strings.TrimSpace(line[len("y-axis "):]))
		return nil
	}

	if strings.HasPrefix(lower, "quadrant-1 ") {
		p.quadrant1 = strings.TrimSpace(line[len("quadrant-1 "):])
		return nil
	}
	if strings.HasPrefix(lower, "quadrant-2 ") {
		p.quadrant2 = strings.TrimSpace(line[len("quadrant-2 "):])
		return nil
	}
	if strings.HasPrefix(lower, "quadrant-3 ") {
		p.quadrant3 = strings.TrimSpace(line[len("quadrant-3 "):])
		return nil
	}
	if strings.HasPrefix(lower, "quadrant-4 ") {
		p.quadrant4 = strings.TrimSpace(line[len("quadrant-4 "):])
		return nil
	}

	// data point: "Label: [x, y]"
	label, rest, ok := strings.Cut(line, ":")
	if !ok {
		return fmt.Errorf("unsupported quadrant chart line: %q", line)
	}
	label = strings.TrimSpace(label)
	rest = strings.TrimSpace(rest)

	if !strings.HasPrefix(rest, "[") || !strings.HasSuffix(rest, "]") {
		return fmt.Errorf("expected [x, y] in %q", line)
	}
	inner := rest[1 : len(rest)-1]
	parts := strings.SplitN(inner, ",", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected two values in %q", line)
	}
	x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return fmt.Errorf("invalid x in %q: %w", line, err)
	}
	y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return fmt.Errorf("invalid y in %q: %w", line, err)
	}

	p.points = append(p.points, ast.NewQuadrantPoint(label, x, y))
	return nil
}

// parseAxisLabels splits "Low Label --> High Label" into the two parts.
// If no "-->" is present, the whole string is used as both.
func parseAxisLabels(s string) (low, high string) {
	const sep = "-->"
	if idx := strings.Index(s, sep); idx >= 0 {
		return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+len(sep):])
	}
	return s, s
}

func (p *quadrantParser) result() (ast.Diagram, error) {
	if len(p.points) == 0 {
		return nil, fmt.Errorf("quadrant chart has no data points")
	}
	return ast.NewQuadrantChart(
		p.title,
		p.xAxisLow, p.xAxisHigh,
		p.yAxisLow, p.yAxisHigh,
		p.quadrant1, p.quadrant2, p.quadrant3, p.quadrant4,
		p.points,
	), nil
}

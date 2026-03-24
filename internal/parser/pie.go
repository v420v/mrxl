package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/v420v/mrxl/internal/ast"
)

type pieParser struct {
	title  string
	slices []*ast.PieSlice
}

func newPieParser() *pieParser {
	return &pieParser{
		slices: make([]*ast.PieSlice, 0),
	}
}

func parsePieChart(lines []string) (ast.Diagram, error) {
	p := newPieParser()
	for _, line := range lines {
		if err := p.parseLine(line); err != nil {
			return nil, err
		}
	}
	return p.result()
}

func (p *pieParser) parseLine(line string) error {
	lower := strings.ToLower(line)

	// title directive: "title My Title"
	if strings.HasPrefix(lower, "title ") {
		p.title = strings.TrimSpace(line[len("title "):])
		return nil
	}

	// skip "showData" keyword
	if lower == "showdata" {
		return nil
	}

	// slice entry: "Label" : value
	label, rest, ok := strings.Cut(line, ":")
	if !ok {
		return fmt.Errorf("unsupported or invalid line: %q", line)
	}
	label = strings.TrimSpace(label)
	label = strings.Trim(label, `"`)

	valStr := strings.TrimSpace(rest)
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return fmt.Errorf("invalid slice value in %q: %w", line, err)
	}

	p.slices = append(p.slices, ast.NewPieSlice(label, val))
	return nil
}

func (p *pieParser) result() (ast.Diagram, error) {
	if len(p.slices) == 0 {
		return nil, fmt.Errorf("pie chart has no slices")
	}
	return ast.NewPieChart(p.title, p.slices), nil
}

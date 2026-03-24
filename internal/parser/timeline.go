package parser

import (
	"fmt"
	"strings"

	"github.com/v420v/mrxl/internal/ast"
)

type timelineParser struct {
	title    string
	sections []*ast.TimeSection
	current  *ast.TimeSection
}

func newTimelineParser() *timelineParser {
	return &timelineParser{
		sections: make([]*ast.TimeSection, 0),
	}
}

func parseTimeline(lines []string) (ast.Diagram, error) {
	p := newTimelineParser()
	for _, line := range lines {
		if err := p.parseLine(line); err != nil {
			return nil, err
		}
	}
	return p.result()
}

func (p *timelineParser) parseLine(line string) error {
	lower := strings.ToLower(line)

	if strings.HasPrefix(lower, "title ") {
		p.title = strings.TrimSpace(line[len("title "):])
		return nil
	}

	if strings.HasPrefix(lower, "section ") {
		name := strings.TrimSpace(line[len("section "):])
		sec := ast.NewTimeSection(name)
		p.sections = append(p.sections, sec)
		p.current = sec
		return nil
	}

	// event line: "time : text" or ": text" (continuation at same time)
	timePart, textPart, ok := strings.Cut(line, ":")
	if !ok {
		return fmt.Errorf("unsupported timeline line: %q", line)
	}
	timePart = strings.TrimSpace(timePart)
	textPart = strings.TrimSpace(textPart)

	if p.current == nil {
		sec := ast.NewTimeSection("")
		p.sections = append(p.sections, sec)
		p.current = sec
	}

	if timePart == "" {
		// Continuation: add text to the last event
		if len(p.current.Events) == 0 {
			return fmt.Errorf("continuation line with no preceding event: %q", line)
		}
		last := p.current.Events[len(p.current.Events)-1]
		last.Texts = append(last.Texts, textPart)
	} else {
		p.current.Events = append(p.current.Events, ast.NewTimeEvent(timePart, textPart))
	}
	return nil
}

func (p *timelineParser) result() (ast.Diagram, error) {
	total := 0
	for _, sec := range p.sections {
		total += len(sec.Events)
	}
	if total == 0 {
		return nil, fmt.Errorf("timeline has no events")
	}
	return ast.NewTimeDiagram(p.title, p.sections), nil
}

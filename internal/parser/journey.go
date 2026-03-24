package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/v420v/mrxl/internal/ast"
)

type journeyParser struct {
	title    string
	sections []*ast.JourneySection
	current  *ast.JourneySection
}

func newJourneyParser() *journeyParser {
	return &journeyParser{sections: make([]*ast.JourneySection, 0)}
}

func parseUserJourney(lines []string) (ast.Diagram, error) {
	p := newJourneyParser()
	for _, line := range lines {
		if err := p.parseLine(line); err != nil {
			return nil, err
		}
	}
	return p.result()
}

func (p *journeyParser) parseLine(line string) error {
	lower := strings.ToLower(line)

	if strings.HasPrefix(lower, "title ") {
		p.title = strings.TrimSpace(line[len("title "):])
		return nil
	}

	if strings.HasPrefix(lower, "section ") {
		name := strings.TrimSpace(line[len("section "):])
		sec := ast.NewJourneySection(name)
		p.sections = append(p.sections, sec)
		p.current = sec
		return nil
	}

	// task line: "Task Title: score: Actor1, Actor2"
	parts := strings.SplitN(line, ":", 3)
	if len(parts) < 2 {
		return fmt.Errorf("unsupported journey line: %q", line)
	}

	title := strings.TrimSpace(parts[0])
	score, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return fmt.Errorf("invalid score in %q: %w", line, err)
	}

	var actors []string
	if len(parts) == 3 {
		for _, a := range strings.Split(parts[2], ",") {
			if a := strings.TrimSpace(a); a != "" {
				actors = append(actors, a)
			}
		}
	}

	if p.current == nil {
		sec := ast.NewJourneySection("")
		p.sections = append(p.sections, sec)
		p.current = sec
	}

	p.current.Tasks = append(p.current.Tasks, ast.NewJourneyTask(title, score, actors))
	return nil
}

func (p *journeyParser) result() (ast.Diagram, error) {
	total := 0
	for _, sec := range p.sections {
		total += len(sec.Tasks)
	}
	if total == 0 {
		return nil, fmt.Errorf("user journey has no tasks")
	}
	return ast.NewUserJourneyDiagram(p.title, p.sections), nil
}

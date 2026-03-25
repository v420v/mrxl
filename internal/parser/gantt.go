package parser

import (
	"fmt"
	"strings"

	"github.com/v420v/mrxl/internal/ast"
)

type ganttParser struct {
	title      string
	dateFormat string
	sections   []*ast.GanttSection
	current    *ast.GanttSection
}

func newGanttParser() *ganttParser {
	return &ganttParser{
		dateFormat: "YYYY-MM-DD",
		sections:   make([]*ast.GanttSection, 0),
	}
}

func parseGantt(lines []string) (ast.Diagram, error) {
	p := newGanttParser()
	for _, line := range lines {
		if err := p.parseLine(line); err != nil {
			return nil, err
		}
	}
	return p.result()
}

func (p *ganttParser) parseLine(line string) error {
	lower := strings.ToLower(line)

	if strings.HasPrefix(lower, "title ") {
		p.title = strings.TrimSpace(line[6:])
		return nil
	}
	if strings.HasPrefix(lower, "dateformat ") {
		p.dateFormat = strings.TrimSpace(line[11:])
		return nil
	}
	// Skip known directives we don't render
	for _, prefix := range []string{
		"axisformat ", "tickinterval ", "weekday ",
		"includes ", "excludes ", "todaymarker ",
	} {
		if strings.HasPrefix(lower, prefix) {
			return nil
		}
	}
	if strings.HasPrefix(lower, "section ") {
		name := strings.TrimSpace(line[8:])
		sec := ast.NewGanttSection(name)
		p.sections = append(p.sections, sec)
		p.current = sec
		return nil
	}

	// Task line: "name : ..."
	namePart, right, ok := strings.Cut(line, ":")
	if !ok {
		return fmt.Errorf("unsupported gantt line: %q", line)
	}

	name := strings.TrimSpace(namePart)
	right = strings.TrimSpace(right)
	task := parseGanttTaskRight(name, right)

	if p.current == nil {
		sec := ast.NewGanttSection("")
		p.sections = append(p.sections, sec)
		p.current = sec
	}
	p.current.Tasks = append(p.current.Tasks, task)
	return nil
}

// parseGanttTaskRight parses the right-hand side of a task definition.
//
// Mermaid Gantt task format (right of ':')):
//
//	[crit,] [done|active|milestone,] [id,] [start|after <id>,] end
func parseGanttTaskRight(name, right string) *ast.GanttTask {
	task := &ast.GanttTask{Name: name}

	parts := strings.Split(right, ",")
	var tokens []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		switch strings.ToLower(t) {
		case "crit":
			task.IsCrit = true
		case "done":
			task.IsDone = true
		case "active":
			task.IsActive = true
		case "milestone":
			task.IsMilestone = true
		default:
			if t != "" {
				tokens = append(tokens, t)
			}
		}
	}

	switch len(tokens) {
	case 0:
		// no date info
	case 1:
		task.EndRaw = tokens[0]
	case 2:
		lower0 := strings.ToLower(tokens[0])
		if strings.HasPrefix(lower0, "after ") {
			task.After = strings.TrimSpace(tokens[0][6:])
		} else if ganttLooksLikeDate(tokens[0]) {
			task.StartRaw = tokens[0]
		} else {
			// treat as ID
			task.ID = tokens[0]
		}
		task.EndRaw = tokens[1]
	case 3:
		task.ID = tokens[0]
		lower1 := strings.ToLower(tokens[1])
		if strings.HasPrefix(lower1, "after ") {
			task.After = strings.TrimSpace(tokens[1][6:])
		} else {
			task.StartRaw = tokens[1]
		}
		task.EndRaw = tokens[2]
	default:
		// best-effort: id, start, ..., end
		task.ID = tokens[0]
		task.StartRaw = tokens[1]
		task.EndRaw = tokens[len(tokens)-1]
	}

	// Auto-generate ID from name for "after" resolution if none given
	if task.ID == "" {
		task.ID = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	}

	return task
}

// ganttLooksLikeDate returns true if the token looks like a date rather than an ID.
func ganttLooksLikeDate(s string) bool {
	return strings.ContainsAny(s, "-/.")
}

func (p *ganttParser) result() (ast.Diagram, error) {
	total := 0
	for _, sec := range p.sections {
		total += len(sec.Tasks)
	}
	if total == 0 {
		return nil, fmt.Errorf("gantt has no tasks")
	}
	return ast.NewGanttDiagram(p.title, p.dateFormat, p.sections), nil
}

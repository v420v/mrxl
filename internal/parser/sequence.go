package parser

import (
	"fmt"
	"strings"

	"github.com/v420v/mrxl/internal/ast"
)

var sequenceArrowSpecs = []struct {
	token     string
	line      ast.LineStyle
	arrowHead ast.ArrowHead
}{
	{"-->>", ast.LineDashed, ast.ArrowFilled},
	{"-->", ast.LineDashed, ast.ArrowOpen},
	{"->>", ast.LineSolid, ast.ArrowFilled},
	{"->", ast.LineSolid, ast.ArrowOpen},
}

type sequenceParser struct {
	title        string
	autonumber   bool
	participants []*ast.Participant
	events       []ast.SequenceEvent
}

func newSequenceParser() *sequenceParser {
	return &sequenceParser{
		participants: make([]*ast.Participant, 0),
		events:       make([]ast.SequenceEvent, 0),
	}
}

func parseSequenceDiagram(lines []string) (ast.Diagram, error) {
	p := newSequenceParser()
	for _, line := range lines {
		if err := p.parseLine(line); err != nil {
			return nil, err
		}
	}
	return p.result()
}

func (p *sequenceParser) addParticipant(name string) *ast.Participant {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	for _, par := range p.participants {
		if par.Name == name {
			return par
		}
	}
	participant := ast.NewParticipant(name)
	p.participants = append(p.participants, participant)
	return participant
}

// parseParticipantLine supports:
//
//	participant Name
//	participant Name as Alias
//	actor Name
//	actor Name as Alias
func parseParticipantLine(line string) (id string, ok bool) {
	lower := strings.ToLower(line)
	var keyword string
	switch {
	case strings.HasPrefix(lower, "participant "):
		keyword = "participant "
	case strings.HasPrefix(lower, "actor "):
		keyword = "actor "
	default:
		return "", false
	}
	rest := strings.TrimSpace(line[len(keyword):])
	if rest == "" {
		return "", false
	}
	fields := strings.Fields(rest)
	if len(fields) >= 3 && strings.EqualFold(fields[1], "as") {
		return fields[0], true
	}
	return fields[0], true
}

func (p *sequenceParser) parseMessageLine(line string) (*ast.Message, *ast.Activation, error) {
	head, rest, ok := strings.Cut(line, ":")
	if !ok {
		return nil, nil, nil
	}
	head = strings.TrimSpace(head)
	text := strings.TrimSpace(rest)

	var (
		from, to string
		lineSt   ast.LineStyle
		arrowHd  ast.ArrowHead
		found    bool
	)
	for _, spec := range sequenceArrowSpecs {
		idx := strings.Index(head, spec.token)
		if idx <= 0 {
			continue
		}
		from = strings.TrimSpace(head[:idx])
		after := strings.TrimSpace(head[idx+len(spec.token):])
		if after == "" {
			return nil, nil, fmt.Errorf("missing target participant in %q", line)
		}
		to = after
		lineSt = spec.line
		arrowHd = spec.arrowHead
		found = true
		break
	}
	if !found {
		return nil, nil, nil
	}

	if from == "" || to == "" {
		return nil, nil, fmt.Errorf("invalid message %q", line)
	}

	// Handle +/- activation shorthand on the target (e.g. A->>+B or A-->>-B).
	var activation *ast.Activation
	if strings.HasPrefix(to, "+") {
		to = to[1:]
		activation = ast.NewActivation(p.addParticipant(to), true)
	} else if strings.HasPrefix(to, "-") {
		to = to[1:]
		activation = ast.NewActivation(p.addParticipant(to), false)
	}

	fromPar := p.addParticipant(from)
	toPar := p.addParticipant(to)

	kind := ast.KindCall
	if lineSt == ast.LineDashed {
		kind = ast.KindReturn
	}

	return ast.NewMessage(fromPar, toPar, lineSt, arrowHd, kind, text), activation, nil
}

// parseNoteLine supports:
//
//	note left of X: text
//	note right of X: text
//	note over X: text
//	note over X,Y: text
func (p *sequenceParser) parseNoteLine(line string) (*ast.Note, error) {
	lower := strings.ToLower(line)
	if !strings.HasPrefix(lower, "note ") {
		return nil, nil
	}
	posText, text, ok := strings.Cut(line, ":")
	if !ok {
		return nil, fmt.Errorf("note missing colon separator: %q", line)
	}
	text = strings.TrimSpace(text)
	posText = strings.TrimSpace(posText)
	posLower := strings.ToLower(posText)

	var pos ast.NotePosition
	var target string
	switch {
	case strings.HasPrefix(posLower, "note left of "):
		pos = ast.NoteLeft
		target = strings.TrimSpace(posText[len("note left of "):])
	case strings.HasPrefix(posLower, "note right of "):
		pos = ast.NoteRight
		target = strings.TrimSpace(posText[len("note right of "):])
	case strings.HasPrefix(posLower, "note over "):
		pos = ast.NoteOver
		target = strings.TrimSpace(posText[len("note over "):])
	default:
		return nil, fmt.Errorf("unsupported note syntax: %q", line)
	}

	if target == "" {
		return nil, fmt.Errorf("note missing participant: %q", line)
	}

	leftName, rightName, hasComma := strings.Cut(target, ",")
	left := p.addParticipant(strings.TrimSpace(leftName))
	right := left
	if hasComma {
		right = p.addParticipant(strings.TrimSpace(rightName))
	}
	return ast.NewNote(pos, left, right, text), nil
}

func (p *sequenceParser) parseLine(line string) error {
	if strings.HasPrefix(strings.ToLower(line), "title ") {
		p.title = strings.TrimSpace(line[len("title "):])
		return nil
	}

	if strings.EqualFold(line, "autonumber") {
		p.autonumber = true
		return nil
	}

	if participant, ok := parseParticipantLine(line); ok {
		p.addParticipant(participant)
		return nil
	}

	lower := strings.ToLower(line)
	if strings.HasPrefix(lower, "activate ") {
		name := strings.TrimSpace(line[len("activate "):])
		p.events = append(p.events, ast.NewActivation(p.addParticipant(name), true))
		return nil
	}
	if strings.HasPrefix(lower, "deactivate ") {
		name := strings.TrimSpace(line[len("deactivate "):])
		p.events = append(p.events, ast.NewActivation(p.addParticipant(name), false))
		return nil
	}

	note, err := p.parseNoteLine(line)
	if err != nil {
		return err
	}
	if note != nil {
		p.events = append(p.events, note)
		return nil
	}

	msg, activation, err := p.parseMessageLine(line)
	if err != nil {
		return err
	}
	if msg != nil {
		p.events = append(p.events, msg)
		if activation != nil {
			p.events = append(p.events, activation)
		}
		return nil
	}

	return fmt.Errorf("unsupported or invalid line: %q", line)
}

func (p *sequenceParser) result() (ast.Diagram, error) {
	return ast.NewSequenceDiagram(p.title, p.autonumber, p.participants, p.events), nil
}

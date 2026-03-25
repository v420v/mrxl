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
	participants []*ast.Participant
	messages     []*ast.Message
}

func newSequenceParser() *sequenceParser {
	return &sequenceParser{
		participants: make([]*ast.Participant, 0),
		messages:     make([]*ast.Message, 0),
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

func (p *sequenceParser) parseMessageLine(line string) (*ast.Message, error) {
	head, rest, ok := strings.Cut(line, ":")
	if !ok {
		return nil, nil
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
			return nil, fmt.Errorf("missing target participant in %q", line)
		}
		to = after
		lineSt = spec.line
		arrowHd = spec.arrowHead
		found = true
		break
	}
	if !found {
		return nil, nil
	}

	if from == "" || to == "" {
		return nil, fmt.Errorf("invalid message %q", line)
	}

	fromPar := p.addParticipant(from)
	toPar := p.addParticipant(to)

	kind := ast.KindCall
	if lineSt == ast.LineDashed {
		kind = ast.KindReturn
	}

	return ast.NewMessage(fromPar, toPar, lineSt, arrowHd, kind, text), nil
}

func (p *sequenceParser) parseLine(line string) error {
	if strings.HasPrefix(strings.ToLower(line), "title ") {
		p.title = strings.TrimSpace(line[len("title "):])
		return nil
	}

	if participant, ok := parseParticipantLine(line); ok {
		p.addParticipant(participant)
		return nil
	}

	msg, err := p.parseMessageLine(line)
	if err != nil {
		return err
	}
	if msg != nil {
		p.messages = append(p.messages, msg)
		return nil
	}

	return fmt.Errorf("unsupported or invalid line: %q", line)
}

func (p *sequenceParser) result() (ast.Diagram, error) {
	return ast.NewSequenceDiagram(p.title, p.participants, p.messages), nil
}

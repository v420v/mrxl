package parser

import (
	"testing"

	"github.com/v420v/mrxl/internal/ast"
)

func TestParserParse(t *testing.T) {
	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantTyp string // ast.Diagram.Type() value
	}{
		{
			name:    "sequenceDiagram",
			input:   "sequenceDiagram\nparticipant A\nA->A: hi",
			wantErr: false,
			wantTyp: "sequence",
		},
		{
			name:    "pie chart",
			input:   "pie\n\"A\": 10",
			wantErr: false,
			wantTyp: "pie",
		},
		{
			name:    "timeline",
			input:   "timeline\n2020 : event",
			wantErr: false,
			wantTyp: "timeline",
		},
		{
			name:    "quadrantChart",
			input:   "quadrantChart\nA: [0.5, 0.5]",
			wantErr: false,
			wantTyp: "quadrant",
		},
		{
			name:    "journey",
			input:   "journey\ntitle T\nsection S\nTask: 5",
			wantErr: false,
			wantTyp: "journey",
		},
		{
			name:    "gantt",
			input:   "gantt\nTask: 1d",
			wantErr: false,
			wantTyp: "gantt",
		},
		{
			name:    "unknown header",
			input:   "unknown\nfoo",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "comment before header",
			input:   "%% comment\nsequenceDiagram\nparticipant A\nA->A: hi",
			wantErr: false,
			wantTyp: "sequence",
		},
		{
			name:    "whitespace around header",
			input:   "  sequenceDiagram  \nparticipant A\nA->A: hi",
			wantErr: false,
			wantTyp: "sequence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got == nil {
				t.Fatal("Parse() returned nil diagram without error")
			}
			if got.Type() != tt.wantTyp {
				t.Errorf("Diagram.Type() = %q, want %q", got.Type(), tt.wantTyp)
			}
		})
	}
}

func TestParserParseConcreteTypes(t *testing.T) {
	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	t.Run("sequenceDiagram returns *ast.SequenceDiagram", func(t *testing.T) {
		got, err := p.Parse("sequenceDiagram\nparticipant A\nA->A: hi")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := got.(*ast.SequenceDiagram); !ok {
			t.Errorf("expected *ast.SequenceDiagram, got %T", got)
		}
	})

	t.Run("pie returns *ast.PieChart", func(t *testing.T) {
		got, err := p.Parse("pie\n\"A\": 10")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := got.(*ast.PieChart); !ok {
			t.Errorf("expected *ast.PieChart, got %T", got)
		}
	})

	t.Run("timeline returns *ast.TimeDiagram", func(t *testing.T) {
		got, err := p.Parse("timeline\n2020 : event")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := got.(*ast.TimeDiagram); !ok {
			t.Errorf("expected *ast.TimeDiagram, got %T", got)
		}
	})

	t.Run("quadrantChart returns *ast.QuadrantChart", func(t *testing.T) {
		got, err := p.Parse("quadrantChart\nA: [0.5, 0.5]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := got.(*ast.QuadrantChart); !ok {
			t.Errorf("expected *ast.QuadrantChart, got %T", got)
		}
	})

	t.Run("journey returns *ast.UserJourneyDiagram", func(t *testing.T) {
		got, err := p.Parse("journey\ntitle T\nsection S\nTask: 5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := got.(*ast.UserJourneyDiagram); !ok {
			t.Errorf("expected *ast.UserJourneyDiagram, got %T", got)
		}
	})

	t.Run("gantt returns *ast.GanttDiagram", func(t *testing.T) {
		got, err := p.Parse("gantt\nTask: 1d")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := got.(*ast.GanttDiagram); !ok {
			t.Errorf("expected *ast.GanttDiagram, got %T", got)
		}
	})
}

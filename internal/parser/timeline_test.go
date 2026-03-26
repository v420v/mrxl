package parser

import (
	"testing"

	"github.com/v420v/mrxl/internal/ast"
)

func TestParseTimeline(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, d *ast.TimeDiagram)
	}{
		{
			name:  "basic event",
			input: "2020 : thing happened",
			check: func(t *testing.T, d *ast.TimeDiagram) {
				t.Helper()
				if len(d.Sections) != 1 {
					t.Fatalf("len(Sections) = %d, want 1", len(d.Sections))
				}
				if len(d.Sections[0].Events) != 1 {
					t.Fatalf("len(Events) = %d, want 1", len(d.Sections[0].Events))
				}
				ev := d.Sections[0].Events[0]
				if ev.Time != "2020" {
					t.Errorf("Time = %q, want %q", ev.Time, "2020")
				}
				if len(ev.Texts) != 1 || ev.Texts[0] != "thing happened" {
					t.Errorf("Texts = %v, want [%q]", ev.Texts, "thing happened")
				}
			},
		},
		{
			name:  "section with events",
			input: "section Era\n2020 : event",
			check: func(t *testing.T, d *ast.TimeDiagram) {
				t.Helper()
				if len(d.Sections) != 1 {
					t.Fatalf("len(Sections) = %d, want 1", len(d.Sections))
				}
				if d.Sections[0].Name != "Era" {
					t.Errorf("Sections[0].Name = %q, want %q", d.Sections[0].Name, "Era")
				}
			},
		},
		{
			name:  "title",
			input: "title My Timeline\n2020 : event",
			check: func(t *testing.T, d *ast.TimeDiagram) {
				t.Helper()
				if d.Title != "My Timeline" {
					t.Errorf("Title = %q, want %q", d.Title, "My Timeline")
				}
			},
		},
		{
			name:  "continuation appends text",
			input: "2020 : first\n: second",
			check: func(t *testing.T, d *ast.TimeDiagram) {
				t.Helper()
				if len(d.Sections) != 1 {
					t.Fatalf("len(Sections) = %d, want 1", len(d.Sections))
				}
				if len(d.Sections[0].Events) != 1 {
					t.Fatalf("len(Events) = %d, want 1", len(d.Sections[0].Events))
				}
				ev := d.Sections[0].Events[0]
				if len(ev.Texts) != 2 {
					t.Fatalf("len(Texts) = %d, want 2", len(ev.Texts))
				}
				if ev.Texts[0] != "first" {
					t.Errorf("Texts[0] = %q, want %q", ev.Texts[0], "first")
				}
				if ev.Texts[1] != "second" {
					t.Errorf("Texts[1] = %q, want %q", ev.Texts[1], "second")
				}
			},
		},
		{
			name:    "continuation without preceding event",
			input:   ": orphan",
			wantErr: true,
		},
		{
			name:    "no events",
			input:   "title only",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeline(normalizedLines(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseTimeline() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			d, ok := got.(*ast.TimeDiagram)
			if !ok {
				t.Fatalf("expected *ast.TimeDiagram, got %T", got)
			}
			if tt.check != nil {
				tt.check(t, d)
			}
		})
	}
}

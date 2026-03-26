package parser

import (
	"testing"

	"github.com/v420v/mrxl/internal/ast"
)

func TestParseUserJourney(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, d *ast.UserJourneyDiagram)
	}{
		{
			name:  "basic task",
			input: "Task: 5",
			check: func(t *testing.T, d *ast.UserJourneyDiagram) {
				t.Helper()
				if len(d.Sections) != 1 {
					t.Fatalf("len(Sections) = %d, want 1", len(d.Sections))
				}
				if len(d.Sections[0].Tasks) != 1 {
					t.Fatalf("len(Tasks) = %d, want 1", len(d.Sections[0].Tasks))
				}
				task := d.Sections[0].Tasks[0]
				if task.Title != "Task" {
					t.Errorf("Title = %q, want %q", task.Title, "Task")
				}
				if task.Score != 5 {
					t.Errorf("Score = %v, want 5", task.Score)
				}
				if len(task.Actors) != 0 {
					t.Errorf("Actors = %v, want empty", task.Actors)
				}
			},
		},
		{
			name:  "task with actors",
			input: "Task: 3: Alice, Bob",
			check: func(t *testing.T, d *ast.UserJourneyDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if len(task.Actors) != 2 {
					t.Fatalf("len(Actors) = %d, want 2", len(task.Actors))
				}
				if task.Actors[0] != "Alice" {
					t.Errorf("Actors[0] = %q, want %q", task.Actors[0], "Alice")
				}
				if task.Actors[1] != "Bob" {
					t.Errorf("Actors[1] = %q, want %q", task.Actors[1], "Bob")
				}
			},
		},
		{
			name:  "section with tasks",
			input: "section Login\nOpen app: 5\nSign in: 3",
			check: func(t *testing.T, d *ast.UserJourneyDiagram) {
				t.Helper()
				if len(d.Sections) != 1 {
					t.Fatalf("len(Sections) = %d, want 1", len(d.Sections))
				}
				if d.Sections[0].Name != "Login" {
					t.Errorf("Sections[0].Name = %q, want %q", d.Sections[0].Name, "Login")
				}
				if len(d.Sections[0].Tasks) != 2 {
					t.Errorf("len(Tasks) = %d, want 2", len(d.Sections[0].Tasks))
				}
			},
		},
		{
			name:  "title",
			input: "title My Journey\nsection S\nTask: 5",
			check: func(t *testing.T, d *ast.UserJourneyDiagram) {
				t.Helper()
				if d.Title != "My Journey" {
					t.Errorf("Title = %q, want %q", d.Title, "My Journey")
				}
			},
		},
		{
			name:    "no tasks",
			input:   "title only",
			wantErr: true,
		},
		{
			name:    "invalid score",
			input:   "Task: notanumber",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUserJourney(normalizedLines(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseUserJourney() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			d, ok := got.(*ast.UserJourneyDiagram)
			if !ok {
				t.Fatalf("expected *ast.UserJourneyDiagram, got %T", got)
			}
			if tt.check != nil {
				tt.check(t, d)
			}
		})
	}
}

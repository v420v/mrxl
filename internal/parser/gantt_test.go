package parser

import (
	"testing"

	"github.com/v420v/mrxl/internal/ast"
)

func TestParseGantt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, d *ast.GanttDiagram)
	}{
		{
			name:  "task with end only",
			input: "Task A: 5d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				if len(d.Sections) == 0 || len(d.Sections[0].Tasks) == 0 {
					t.Fatal("expected 1 task")
				}
				task := d.Sections[0].Tasks[0]
				if task.Name != "Task A" {
					t.Errorf("Name = %q, want %q", task.Name, "Task A")
				}
				if task.EndRaw != "5d" {
					t.Errorf("EndRaw = %q, want %q", task.EndRaw, "5d")
				}
				if task.ID != "task_a" {
					t.Errorf("ID = %q, want %q", task.ID, "task_a")
				}
			},
		},
		{
			name:  "task with start and end",
			input: "Task: 2024-01-01, 30d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if task.StartRaw != "2024-01-01" {
					t.Errorf("StartRaw = %q, want %q", task.StartRaw, "2024-01-01")
				}
				if task.EndRaw != "30d" {
					t.Errorf("EndRaw = %q, want %q", task.EndRaw, "30d")
				}
			},
		},
		{
			name:  "task with after and end",
			input: "Task: after dep1, 5d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if task.After != "dep1" {
					t.Errorf("After = %q, want %q", task.After, "dep1")
				}
				if task.EndRaw != "5d" {
					t.Errorf("EndRaw = %q, want %q", task.EndRaw, "5d")
				}
			},
		},
		{
			name:  "task with id after and end",
			input: "Task: taskId, after dep1, 5d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if task.ID != "taskId" {
					t.Errorf("ID = %q, want %q", task.ID, "taskId")
				}
				if task.After != "dep1" {
					t.Errorf("After = %q, want %q", task.After, "dep1")
				}
				if task.EndRaw != "5d" {
					t.Errorf("EndRaw = %q, want %q", task.EndRaw, "5d")
				}
			},
		},
		{
			name:  "crit flag",
			input: "Task: crit, 5d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if !task.IsCrit {
					t.Error("IsCrit = false, want true")
				}
			},
		},
		{
			name:  "done flag",
			input: "Task: done, 5d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if !task.IsDone {
					t.Error("IsDone = false, want true")
				}
			},
		},
		{
			name:  "active flag",
			input: "Task: active, 5d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if !task.IsActive {
					t.Error("IsActive = false, want true")
				}
			},
		},
		{
			name:  "milestone flag",
			input: "Task: milestone, 5d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if !task.IsMilestone {
					t.Error("IsMilestone = false, want true")
				}
			},
		},
		{
			name:  "multiple flags",
			input: "Task: crit, done, 5d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if !task.IsCrit {
					t.Error("IsCrit = false, want true")
				}
				if !task.IsDone {
					t.Error("IsDone = false, want true")
				}
			},
		},
		{
			name:  "section groups tasks",
			input: "section Phase1\nTask A: 1d\nTask B: 2d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				if len(d.Sections) != 1 {
					t.Fatalf("len(Sections) = %d, want 1", len(d.Sections))
				}
				if d.Sections[0].Name != "Phase1" {
					t.Errorf("Sections[0].Name = %q, want %q", d.Sections[0].Name, "Phase1")
				}
				if len(d.Sections[0].Tasks) != 2 {
					t.Errorf("len(Sections[0].Tasks) = %d, want 2", len(d.Sections[0].Tasks))
				}
			},
		},
		{
			name:  "title and dateFormat",
			input: "title My Gantt\ndateFormat YYYY-MM-DD\nTask: 1d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				if d.Title != "My Gantt" {
					t.Errorf("Title = %q, want %q", d.Title, "My Gantt")
				}
				if d.DateFormat != "YYYY-MM-DD" {
					t.Errorf("DateFormat = %q, want %q", d.DateFormat, "YYYY-MM-DD")
				}
			},
		},
		{
			name:  "skipped directives",
			input: "axisFormat %m/%d\nTask: 1d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				totalTasks := 0
				for _, sec := range d.Sections {
					totalTasks += len(sec.Tasks)
				}
				if totalTasks != 1 {
					t.Errorf("total tasks = %d, want 1", totalTasks)
				}
			},
		},
		{
			name:  "auto-generated ID",
			input: "My Task Name: 5d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				task := d.Sections[0].Tasks[0]
				if task.ID != "my_task_name" {
					t.Errorf("ID = %q, want %q", task.ID, "my_task_name")
				}
			},
		},
		{
			name:  "default dateFormat when not specified",
			input: "Task: 1d",
			check: func(t *testing.T, d *ast.GanttDiagram) {
				t.Helper()
				if d.DateFormat != "YYYY-MM-DD" {
					t.Errorf("DateFormat = %q, want %q", d.DateFormat, "YYYY-MM-DD")
				}
			},
		},
		{
			name:    "no tasks",
			input:   "title only",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGantt(normalizedLines(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseGantt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			d, ok := got.(*ast.GanttDiagram)
			if !ok {
				t.Fatalf("expected *ast.GanttDiagram, got %T", got)
			}
			if tt.check != nil {
				tt.check(t, d)
			}
		})
	}
}

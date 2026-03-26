package parser

import (
	"testing"

	"github.com/v420v/mrxl/internal/ast"
)

func TestParseQuadrantChart(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, d *ast.QuadrantChart)
	}{
		{
			name:  "basic point",
			input: "A: [0.5, 0.7]",
			check: func(t *testing.T, d *ast.QuadrantChart) {
				t.Helper()
				if len(d.Points) != 1 {
					t.Fatalf("len(Points) = %d, want 1", len(d.Points))
				}
				pt := d.Points[0]
				if pt.Label != "A" {
					t.Errorf("Label = %q, want %q", pt.Label, "A")
				}
				if pt.X != 0.5 {
					t.Errorf("X = %v, want 0.5", pt.X)
				}
				if pt.Y != 0.7 {
					t.Errorf("Y = %v, want 0.7", pt.Y)
				}
			},
		},
		{
			name:  "title",
			input: "title Q Chart\nA: [0.1, 0.2]",
			check: func(t *testing.T, d *ast.QuadrantChart) {
				t.Helper()
				if d.Title != "Q Chart" {
					t.Errorf("Title = %q, want %q", d.Title, "Q Chart")
				}
			},
		},
		{
			name:  "x-axis with separator",
			input: "x-axis Low --> High\nA: [0.5, 0.5]",
			check: func(t *testing.T, d *ast.QuadrantChart) {
				t.Helper()
				if d.XAxisLow != "Low" {
					t.Errorf("XAxisLow = %q, want %q", d.XAxisLow, "Low")
				}
				if d.XAxisHigh != "High" {
					t.Errorf("XAxisHigh = %q, want %q", d.XAxisHigh, "High")
				}
			},
		},
		{
			name:  "y-axis with separator",
			input: "y-axis Bottom --> Top\nA: [0.5, 0.5]",
			check: func(t *testing.T, d *ast.QuadrantChart) {
				t.Helper()
				if d.YAxisLow != "Bottom" {
					t.Errorf("YAxisLow = %q, want %q", d.YAxisLow, "Bottom")
				}
				if d.YAxisHigh != "Top" {
					t.Errorf("YAxisHigh = %q, want %q", d.YAxisHigh, "Top")
				}
			},
		},
		{
			name:  "quadrant labels",
			input: "quadrant-1 Niche\nquadrant-2 Stars\nquadrant-3 Dogs\nquadrant-4 Cash\nA: [0.5, 0.5]",
			check: func(t *testing.T, d *ast.QuadrantChart) {
				t.Helper()
				if d.Quadrant1 != "Niche" {
					t.Errorf("Quadrant1 = %q, want %q", d.Quadrant1, "Niche")
				}
				if d.Quadrant2 != "Stars" {
					t.Errorf("Quadrant2 = %q, want %q", d.Quadrant2, "Stars")
				}
				if d.Quadrant3 != "Dogs" {
					t.Errorf("Quadrant3 = %q, want %q", d.Quadrant3, "Dogs")
				}
				if d.Quadrant4 != "Cash" {
					t.Errorf("Quadrant4 = %q, want %q", d.Quadrant4, "Cash")
				}
			},
		},
		{
			name:    "no points",
			input:   "title only",
			wantErr: true,
		},
		{
			name:    "invalid x",
			input:   "A: [bad, 0.5]",
			wantErr: true,
		},
		{
			name:    "invalid y",
			input:   "A: [0.5, bad]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseQuadrantChart(normalizedLines(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseQuadrantChart() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			d, ok := got.(*ast.QuadrantChart)
			if !ok {
				t.Fatalf("expected *ast.QuadrantChart, got %T", got)
			}
			if tt.check != nil {
				tt.check(t, d)
			}
		})
	}
}

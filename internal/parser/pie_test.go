package parser

import (
	"testing"

	"github.com/v420v/mrxl/internal/ast"
)

func TestParsePieChart(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantErr      bool
		wantTitle    string
		wantShowData bool
		wantSlices   int
		checkSlices  func(t *testing.T, slices []*ast.PieSlice)
	}{
		{
			name:       "two slices",
			input:      "\"A\": 30\n\"B\": 70",
			wantErr:    false,
			wantSlices: 2,
			checkSlices: func(t *testing.T, slices []*ast.PieSlice) {
				t.Helper()
				if slices[0].Label != "A" {
					t.Errorf("slices[0].Label = %q, want %q", slices[0].Label, "A")
				}
				if slices[0].Value != 30 {
					t.Errorf("slices[0].Value = %v, want 30", slices[0].Value)
				}
				if slices[1].Label != "B" {
					t.Errorf("slices[1].Label = %q, want %q", slices[1].Label, "B")
				}
				if slices[1].Value != 70 {
					t.Errorf("slices[1].Value = %v, want 70", slices[1].Value)
				}
			},
		},
		{
			name:       "title",
			input:      "title My Pie\n\"X\": 1",
			wantErr:    false,
			wantTitle:  "My Pie",
			wantSlices: 1,
		},
		{
			name:         "showData standalone",
			input:        "showData\n\"A\": 1",
			wantErr:      false,
			wantShowData: true,
			wantSlices:   1,
		},
		{
			name:         "showData inline with title",
			input:        "showData title Pets\n\"A\": 1",
			wantErr:      false,
			wantShowData: true,
			wantTitle:    "Pets",
			wantSlices:   1,
		},
		{
			name:       "unquoted label",
			input:      "A: 10",
			wantErr:    false,
			wantSlices: 1,
			checkSlices: func(t *testing.T, slices []*ast.PieSlice) {
				t.Helper()
				if slices[0].Label != "A" {
					t.Errorf("slices[0].Label = %q, want %q", slices[0].Label, "A")
				}
				if slices[0].Value != 10 {
					t.Errorf("slices[0].Value = %v, want 10", slices[0].Value)
				}
			},
		},
		{
			name:    "no slices title only",
			input:   "title T",
			wantErr: true,
		},
		{
			name:    "invalid float",
			input:   "\"A\": notanumber",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePieChart(normalizedLines(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePieChart() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			chart, ok := got.(*ast.PieChart)
			if !ok {
				t.Fatalf("expected *ast.PieChart, got %T", got)
			}

			if tt.wantTitle != "" && chart.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", chart.Title, tt.wantTitle)
			}

			if chart.ShowData != tt.wantShowData {
				t.Errorf("ShowData = %v, want %v", chart.ShowData, tt.wantShowData)
			}

			if len(chart.Slices) != tt.wantSlices {
				t.Errorf("len(Slices) = %d, want %d", len(chart.Slices), tt.wantSlices)
			}

			if tt.checkSlices != nil {
				tt.checkSlices(t, chart.Slices)
			}
		})
	}
}

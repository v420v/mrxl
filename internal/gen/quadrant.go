package gen

import (
	"fmt"
	"strings"

	"github.com/v420v/mrxl/internal/ast"
	"github.com/xuri/excelize/v2"
)

// QuadrantDrawing renders a QuadrantChart onto an Excel sheet.
type QuadrantDrawing struct {
	File    *excelize.File
	Sheet   string
	Diagram *ast.QuadrantChart
}

func (g *QuadrantDrawing) drawQuadrantChart() error {
	d := g.Diagram
	f := g.File

	// --- Data table header ---
	for col, hdr := range []string{"Label", "X", "Y"} {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return fmt.Errorf("header cell: %w", err)
		}
		if err := f.SetCellValue(g.Sheet, cell, hdr); err != nil {
			return fmt.Errorf("set header %q: %w", hdr, err)
		}
	}

	// --- Data point rows ---
	for i, pt := range d.Points {
		row := i + 2
		labelCell, _ := excelize.CoordinatesToCellName(1, row)
		xCell, _ := excelize.CoordinatesToCellName(2, row)
		yCell, _ := excelize.CoordinatesToCellName(3, row)
		if err := f.SetCellValue(g.Sheet, labelCell, pt.Label); err != nil {
			return fmt.Errorf("set label %q: %w", pt.Label, err)
		}
		if err := f.SetCellValue(g.Sheet, xCell, pt.X); err != nil {
			return fmt.Errorf("set x for %q: %w", pt.Label, err)
		}
		if err := f.SetCellValue(g.Sheet, yCell, pt.Y); err != nil {
			return fmt.Errorf("set y for %q: %w", pt.Label, err)
		}
	}

	// --- Quadrant divider data (rows after the data points) ---
	n := len(d.Points)
	vDiv1 := n + 3 // vertical divider point 1:   (0.5, 0.0)
	vDiv2 := n + 4 // vertical divider point 2:   (0.5, 1.0)
	hDiv1 := n + 5 // horizontal divider point 1: (0.0, 0.5)
	hDiv2 := n + 6 // horizontal divider point 2: (1.0, 0.5)

	type xy struct{ x, y float64 }
	divPts := []struct {
		row int
		xy  xy
	}{
		{vDiv1, xy{0.5, 0.0}},
		{vDiv2, xy{0.5, 1.0}},
		{hDiv1, xy{0.0, 0.5}},
		{hDiv2, xy{1.0, 0.5}},
	}
	for _, dp := range divPts {
		xCell, _ := excelize.CoordinatesToCellName(2, dp.row)
		yCell, _ := excelize.CoordinatesToCellName(3, dp.row)
		if err := f.SetCellValue(g.Sheet, xCell, dp.xy.x); err != nil {
			return fmt.Errorf("divider x row %d: %w", dp.row, err)
		}
		if err := f.SetCellValue(g.Sheet, yCell, dp.xy.y); err != nil {
			return fmt.Errorf("divider y row %d: %w", dp.row, err)
		}
	}

	// --- Build chart series ---
	// One series per data point so ShowSerName labels each point.
	series := make([]excelize.ChartSeries, 0, n+2)
	for i := range d.Points {
		row := i + 2
		series = append(series, excelize.ChartSeries{
			Name:       fmt.Sprintf("%s!$A$%d", g.Sheet, row),
			Categories: fmt.Sprintf("%s!$B$%d:$B$%d", g.Sheet, row, row),
			Values:     fmt.Sprintf("%s!$C$%d:$C$%d", g.Sheet, row, row),
			Line:       excelize.ChartLine{Type: excelize.ChartLineNone},
		})
	}

	// Quadrant divider lines: gray, no endpoint markers.
	divLine := excelize.ChartLine{
		Type:  excelize.ChartLineSolid,
		Width: 1.0,
		Fill:  excelize.Fill{Color: []string{"AAAAAA"}, Pattern: 1},
	}
	noMarker := excelize.ChartMarker{Symbol: "none"}

	series = append(series,
		excelize.ChartSeries{
			Name:       " ",
			Categories: fmt.Sprintf("%s!$B$%d:$B$%d", g.Sheet, vDiv1, vDiv2),
			Values:     fmt.Sprintf("%s!$C$%d:$C$%d", g.Sheet, vDiv1, vDiv2),
			Line:       divLine,
			Marker:     noMarker,
		},
		excelize.ChartSeries{
			Name:       " ",
			Categories: fmt.Sprintf("%s!$B$%d:$B$%d", g.Sheet, hDiv1, hDiv2),
			Values:     fmt.Sprintf("%s!$C$%d:$C$%d", g.Sheet, hDiv1, hDiv2),
			Line:       divLine,
			Marker:     noMarker,
		},
	)

	minVal := 0.0
	maxVal := 1.0

	title := d.Title
	if title == "" {
		title = "Quadrant Chart"
	}

	chart := &excelize.Chart{
		Type:   excelize.Scatter,
		Series: series,
		Title:  []excelize.RichTextRun{{Text: title}},
		PlotArea: excelize.ChartPlotArea{
			ShowSerName: true,
		},
		XAxis: excelize.ChartAxis{
			Minimum: &minVal,
			Maximum: &maxVal,
		},
		YAxis: excelize.ChartAxis{
			Minimum: &minVal,
			Maximum: &maxVal,
		},
	}

	xTitle := quadrantAxisTitle(d.XAxisLow, d.XAxisHigh)
	yTitle := quadrantAxisTitle(d.YAxisLow, d.YAxisHigh)
	if xTitle != "" {
		chart.XAxis.Title = []excelize.RichTextRun{{Text: xTitle}}
	}
	if yTitle != "" {
		chart.YAxis.Title = []excelize.RichTextRun{{Text: yTitle}}
	}

	if err := f.AddChart(g.Sheet, "E2", chart); err != nil {
		return fmt.Errorf("add quadrant chart: %w", err)
	}
	return nil
}

func quadrantAxisTitle(low, high string) string {
	low = strings.TrimSpace(low)
	high = strings.TrimSpace(high)
	if low == "" && high == "" {
		return ""
	}
	if low == high || high == "" {
		return low
	}
	if low == "" {
		return high
	}
	return low + " → " + high
}

package gen

import (
	"fmt"

	"github.com/v420v/mrxl/internal/ast"
	"github.com/xuri/excelize/v2"
)

type PieDrawing struct {
	File    *excelize.File
	Sheet   string
	Diagram *ast.PieChart
}

func (g *PieDrawing) drawPieChart() error {
	d := g.Diagram
	f := g.File

	// Write header row
	if err := f.SetCellValue(g.Sheet, "A1", "Label"); err != nil {
		return fmt.Errorf("set header label: %w", err)
	}
	if err := f.SetCellValue(g.Sheet, "B1", "Value"); err != nil {
		return fmt.Errorf("set header value: %w", err)
	}

	// Write slice data starting at row 2
	for i, slice := range d.Slices {
		row := i + 2
		labelCell, err := excelize.CoordinatesToCellName(1, row)
		if err != nil {
			return fmt.Errorf("label cell name: %w", err)
		}
		valueCell, err := excelize.CoordinatesToCellName(2, row)
		if err != nil {
			return fmt.Errorf("value cell name: %w", err)
		}
		if err := f.SetCellValue(g.Sheet, labelCell, slice.Label); err != nil {
			return fmt.Errorf("set label %q: %w", slice.Label, err)
		}
		if err := f.SetCellValue(g.Sheet, valueCell, slice.Value); err != nil {
			return fmt.Errorf("set value for %q: %w", slice.Label, err)
		}
	}

	lastRow := len(d.Slices) + 1
	categoriesRange := fmt.Sprintf("%s!$A$2:$A$%d", g.Sheet, lastRow)
	valuesRange := fmt.Sprintf("%s!$B$2:$B$%d", g.Sheet, lastRow)

	title := d.Title
	if title == "" {
		title = "Pie Chart"
	}

	chart := &excelize.Chart{
		Type: excelize.Pie,
		Series: []excelize.ChartSeries{
			{
				Name:       title,
				Categories: categoriesRange,
				Values:     valuesRange,
			},
		},
		Title: []excelize.RichTextRun{
			{Text: title},
		},
		Legend: excelize.ChartLegend{
			Position: "right",
		},
		PlotArea: excelize.ChartPlotArea{
			ShowPercent: true,
		},
	}

	if err := f.AddChart(g.Sheet, "D2", chart); err != nil {
		return fmt.Errorf("add pie chart: %w", err)
	}

	return nil
}

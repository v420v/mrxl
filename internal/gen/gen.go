package gen

import (
	"fmt"

	"github.com/v420v/mrxl/internal/ast"
	"github.com/xuri/excelize/v2"
)

func Generate(diagram ast.Diagram, outFile string) error {
	f := excelize.NewFile()

	switch d := diagram.(type) {
	case *ast.SequenceDiagram:
		const sheet = "SequenceDiagram"
		if err := f.SetSheetName("Sheet1", sheet); err != nil {
			return fmt.Errorf("set sheet name: %v", err)
		}
		g := &SequenceDrawing{File: f, Sheet: sheet, Diagram: d}
		if err := g.drawSequenceDiagram(); err != nil {
			return fmt.Errorf("draw sequence diagram: %w", err)
		}
	case *ast.PieChart:
		const sheet = "PieChart"
		if err := f.SetSheetName("Sheet1", sheet); err != nil {
			return fmt.Errorf("set sheet name: %v", err)
		}
		g := &PieDrawing{File: f, Sheet: sheet, Diagram: d}
		if err := g.drawPieChart(); err != nil {
			return fmt.Errorf("draw pie chart: %w", err)
		}
	case *ast.UserJourneyDiagram:
		const sheet = "UserJourney"
		if err := f.SetSheetName("Sheet1", sheet); err != nil {
			return fmt.Errorf("set sheet name: %v", err)
		}
		g := &JourneyDrawing{File: f, Sheet: sheet, Diagram: d}
		if err := g.drawUserJourney(); err != nil {
			return fmt.Errorf("draw user journey: %w", err)
		}
	case *ast.QuadrantChart:
		const sheet = "QuadrantChart"
		if err := f.SetSheetName("Sheet1", sheet); err != nil {
			return fmt.Errorf("set sheet name: %v", err)
		}
		g := &QuadrantDrawing{File: f, Sheet: sheet, Diagram: d}
		if err := g.drawQuadrantChart(); err != nil {
			return fmt.Errorf("draw quadrant chart: %w", err)
		}
	case *ast.TimeDiagram:
		const sheet = "Timeline"
		if err := f.SetSheetName("Sheet1", sheet); err != nil {
			return fmt.Errorf("set sheet name: %v", err)
		}
		g := &TimelineDrawing{File: f, Sheet: sheet, Diagram: d}
		if err := g.drawTimeline(); err != nil {
			return fmt.Errorf("draw timeline: %w", err)
		}
	case *ast.GanttDiagram:
		const sheet = "Gantt"
		if err := f.SetSheetName("Sheet1", sheet); err != nil {
			return fmt.Errorf("set sheet name: %v", err)
		}
		g := &GanttDrawing{File: f, Sheet: sheet, Diagram: d}
		if err := g.drawGantt(); err != nil {
			return fmt.Errorf("draw gantt: %w", err)
		}
	default:
		return fmt.Errorf("unsupported diagram type: %T", d)
	}

	return f.SaveAs(outFile)
}

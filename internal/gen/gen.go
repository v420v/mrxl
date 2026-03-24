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
	default:
		return fmt.Errorf("unsupported diagram type: %T", d)
	}

	return f.SaveAs(outFile)
}

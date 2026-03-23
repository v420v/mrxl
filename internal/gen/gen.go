package gen

import (
	"fmt"

	"github.com/v420v/mrxl/internal/ast"
	"github.com/xuri/excelize/v2"
)

func Generate(diagram ast.Diagram, outFile string) error {
	const sheet = "SequenceDiagram"
	f := excelize.NewFile()
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return fmt.Errorf("set sheet name: %v", err)
	}

	switch d := diagram.(type) {
	case *ast.SequenceDiagram:
		g := &SequenceDrawing{File: f, Sheet: sheet, Diagram: d}
		if err := g.drawSequenceDiagram(); err != nil {
			return fmt.Errorf("draw sequence diagram: %w", err)
		}
	default:
		return fmt.Errorf("unsupported diagram type: %T", d)
	}

	return f.SaveAs(outFile)
}

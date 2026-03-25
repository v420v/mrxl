package gen

import (
	"fmt"
	"math"

	"github.com/v420v/mrxl/internal/ast"
	"github.com/xuri/excelize/v2"
)

const (
	defaultColWidthExcel = 9.140625
	defaultRowHeightVal  = 15.0
	// Convert the message arrow height (in pt) to the AddShape height (in pixels) using the formula: cy = px × 9525 EMU, where 1 pt = 12,700 EMU.
	sequenceArrowHpx = (28*12700 + 9525/2) / 9525
	participantLine  = 1.5
	lifelineLine     = 1.0
	// Arrow outlines: Since Excelize does not provide a "no line" option, use the minimum allowable width of 0.25pt and set the color to match the fill to make the outline invisible.
	arrowLinePt = 0.25
)

const (
	firstMessageRow = 5
	messageRowStep  = 2
	// Left to right: column A = narrow margin; each participant uses one wide column, then sequenceGapColumns narrow spacer columns.
	sequenceFirstLaneCol = 2 // First participant column (1-based). Column 1 is narrow margin.
	// Narrow columns between adjacent participant lanes (was 1; now 2).
	sequenceGapColumns = 2
	// Offset from one lane column to the next: one wide lane + sequenceGapColumns narrow gaps.
	sequenceLaneStride = 1 + sequenceGapColumns
	sequenceWideColW   = 11.5
	sequenceNarrowColW = 2.8
	// Practical cap aligned with excelize column limits.
	sequenceMaxParticipants = 8000
)

// SequenceDrawing lays out sequence diagram shapes with excelize.AddShape.
type SequenceDrawing struct {
	File    *excelize.File
	Sheet   string
	Diagram *ast.SequenceDiagram
}

// SequenceSheetColumnCount returns how many columns to size for n participants (widths set for 1 through this index).
// Layout: column 1 = narrow margin; lanes at 2, 2+stride, 2+2*stride, … (wide); columns between lanes are narrow gaps.
// Last used column index = sequenceFirstLaneCol + (n-1)*sequenceLaneStride.
func sequenceSheetColumnCount(participantCount int) int {
	if participantCount < 1 {
		return 0
	}
	return sequenceFirstLaneCol + (participantCount-1)*sequenceLaneStride
}

// SequenceLaneColumns returns 1-based lane column indices for n participants, packed from the left (not evenly spread).
func sequenceLaneColumns(n int) ([]int, error) {
	if n < 1 {
		return nil, fmt.Errorf("at least one participant is required")
	}
	if n > sequenceMaxParticipants {
		return nil, fmt.Errorf("at most %d participants are supported", sequenceMaxParticipants)
	}
	cols := make([]int, n)
	for i := range n {
		cols[i] = sequenceFirstLaneCol + i*sequenceLaneStride
	}
	return cols, nil
}

// applySequenceColumnWidths sets widths only for columns 1..SequenceSheetColumnCount (call before drawing).
func applySequenceColumnWidths(f *excelize.File, sheet string, laneCols []int) error {
	if len(laneCols) == 0 {
		return fmt.Errorf("laneCols is empty")
	}
	n := len(laneCols)
	totalCols := sequenceSheetColumnCount(n)
	wide := make(map[int]bool, n)
	for _, c := range laneCols {
		wide[c] = true
	}
	for c := 1; c <= totalCols; c++ {
		colName, err := excelize.ColumnNumberToName(c)
		if err != nil {
			return err
		}
		w := sequenceNarrowColW
		if wide[c] {
			w = sequenceWideColW
		}
		if err := f.SetColWidth(sheet, colName, colName, w); err != nil {
			return err
		}
	}
	return nil
}

// requiredRowCount returns the row count to use for row heights given the diagram.
func requiredRowCount(d *ast.SequenceDiagram) int {
	if d == nil || len(d.Events) == 0 {
		return 15
	}
	lastEvent := firstMessageRow + (len(d.Events)-1)*messageRowStep
	end := max(lastEvent+3, 10)
	return end + 2
}

func (g *SequenceDrawing) sumRowHeightsPx(fromRow, toRow int) int {
	s := 0
	for r := fromRow; r <= toRow; r++ {
		s += g.rowHeightPx(r)
	}
	return s
}

func messageStyle(m *ast.Message) (fill, fg string) {
	if m.LineStyle == ast.LineDashed || m.MessageKind == ast.KindReturn {
		return "E8E8E8", "404040"
	}
	return "D6EAF9", "1F4E79"
}

func (g *SequenceDrawing) colWidthPx(col1Based int) int {
	name, err := excelize.ColumnNumberToName(col1Based)
	if err != nil {
		return 64
	}
	w, err := g.File.GetColWidth(g.Sheet, name)
	if err != nil || w == 0 {
		w = defaultColWidthExcel
	}
	return int(w*8 + 0.5)
}

func (g *SequenceDrawing) rowHeightPx(row1Based int) int {
	h, err := g.File.GetRowHeight(g.Sheet, row1Based)
	if err != nil || h == 0 {
		h = defaultRowHeightVal
	}
	return int(math.Ceil(4.0 / 3.4 * h))
}

func (g *SequenceDrawing) laneAnchorOffX(col1Based int) int {
	cw := g.colWidthPx(col1Based)
	return max(0, cw/2-3)
}

// Returns the width in pixels from the from-column lifeline (fromOff from left) to the to-column lifeline (toOff from left).
func (g *SequenceDrawing) arrowWidthBetweenLifelines(fromCol int, fromOff int, toCol int, toOff int) int {
	w := 0
	if fromCol < toCol {
		w = g.colWidthPx(fromCol) - fromOff
		for c := fromCol + 1; c < toCol; c++ {
			w += g.colWidthPx(c)
		}
		w += toOff
	} else if fromCol > toCol {
		w = fromOff
		for c := toCol + 1; c < fromCol; c++ {
			w += g.colWidthPx(c)
		}
		w += toOff
	} else {
		w = max(fromOff, toOff) - min(fromOff, toOff)
	}
	// set the minimum width
	if w < 28 {
		w = 28
	}
	return w
}

func (g *SequenceDrawing) arrowOffsetY(row1Based int, shapeHPx int) int {
	rh := g.rowHeightPx(row1Based)
	if rh > shapeHPx {
		return (rh - shapeHPx) / 2
	}
	return 0
}

// Aligns the shape left edge to the left lifeline and right edge to the right lifeline.
// fromCol/toCol are sender/receiver; when from is right of to, uses leftArrow so the head points at to on the left.
func (g *SequenceDrawing) addDirectionalArrow(fromCol int, toCol int, row int, shapeName string, text string, lwA *float64, fill string, fg string) error {
	leftCol := min(fromCol, toCol)
	rightCol := max(fromCol, toCol)
	offL := g.laneAnchorOffX(leftCol)
	offR := g.laneAnchorOffX(rightCol)
	wPx := g.arrowWidthBetweenLifelines(leftCol, offL, rightCol, offR)

	colName, err := excelize.ColumnNumberToName(leftCol)
	if err != nil {
		return err
	}
	cell := fmt.Sprintf("%s%d", colName, row)
	_, r, err := excelize.CellNameToCoordinates(cell)
	if err != nil {
		return err
	}
	offY := g.arrowOffsetY(r, sequenceArrowHpx)

	prst := "rightArrow"
	if fromCol > toCol {
		prst = "leftArrow"
	}

	return g.File.AddShape(g.Sheet, &excelize.Shape{
		Cell:   cell,
		Type:   prst,
		Width:  uint(wPx),
		Height: uint(sequenceArrowHpx),
		Line:   excelize.ShapeLine{Color: fill, Width: lwA},
		Fill:   excelize.Fill{Color: []string{fill}, Pattern: 1},
		Paragraph: []excelize.RichTextRun{{
			Text: text,
			Font: &excelize.Font{Bold: true, Size: 8, Color: fg, Family: "Calibri"},
		}},
		Format: excelize.GraphicOptions{
			Name:    shapeName,
			OffsetX: offL,
			OffsetY: offY,
		},
	})
}

func (g *SequenceDrawing) drawSequenceDiagram() error {
	laneCols, err := sequenceLaneColumns(len(g.Diagram.Participants))
	if err != nil {
		return err
	}

	if err := applySequenceColumnWidths(g.File, g.Sheet, laneCols); err != nil {
		return err
	}

	nRows := requiredRowCount(g.Diagram)
	for row := 1; row <= nRows; row++ {
		h := 22.0
		if row == 2 {
			h = 32
		}
		if err := g.File.SetRowHeight(g.Sheet, row, h); err != nil {
			return err
		}
	}

	if len(g.Diagram.Participants) == 0 {
		return fmt.Errorf("sequence diagram: no participants")
	}

	if g.Diagram.Title != "" {
		lastCol := sequenceSheetColumnCount(len(g.Diagram.Participants))
		lastColName, err := excelize.ColumnNumberToName(lastCol)
		if err != nil {
			return err
		}
		if err := g.File.MergeCell(g.Sheet, "A1", lastColName+"1"); err != nil {
			return err
		}
		titleStyle, err := g.File.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 14, Color: "1F2A44", Family: "Calibri"},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		})
		if err != nil {
			return err
		}
		if err := g.File.SetCellValue(g.Sheet, "A1", g.Diagram.Title); err != nil {
			return err
		}
		if err := g.File.SetCellStyle(g.Sheet, "A1", lastColName+"1", titleStyle); err != nil {
			return err
		}
	}

	nameToCol := make(map[string]int)
	for i, p := range g.Diagram.Participants {
		if p.Name == "" {
			return fmt.Errorf("participant %d: empty name", i)
		}
		if _, dup := nameToCol[p.Name]; dup {
			return fmt.Errorf("duplicate participant name: %q", p.Name)
		}
		nameToCol[p.Name] = laneCols[i]
	}

	lwP := participantLine
	lwL := lifelineLine
	lwA := arrowLinePt

	lastMsgRow := firstMessageRow
	if len(g.Diagram.Events) > 0 {
		lastMsgRow = firstMessageRow + (len(g.Diagram.Events)-1)*messageRowStep
	}
	lifelineEndRow := lastMsgRow + 3
	if lifelineEndRow < 10 {
		lifelineEndRow = 10
	}
	lifelineH := g.sumRowHeightsPx(3, lifelineEndRow)

	for i, p := range g.Diagram.Participants {
		colNum := laneCols[i]
		colName, err := excelize.ColumnNumberToName(colNum)
		if err != nil {
			return err
		}
		if err := g.File.AddShape(g.Sheet, &excelize.Shape{
			Cell:   colName + "2",
			Type:   "rect",
			Width:  88,
			Height: 30,
			Line:   excelize.ShapeLine{Color: "2F5597", Width: &lwP},
			Fill:   excelize.Fill{Color: []string{"B4C6E7"}, Pattern: 1},
			Paragraph: []excelize.RichTextRun{{
				Text: p.Name,
				Font: &excelize.Font{Bold: true, Size: 11, Color: "1F2A44", Family: "Calibri"},
			}},
			Format: excelize.GraphicOptions{Name: "P_" + p.Name},
		}); err != nil {
			return fmt.Errorf("participant %q: %w", p.Name, err)
		}

		lifelineWPx := 5
		offLX := max(0, (g.colWidthPx(colNum)-lifelineWPx)/2)
		if err := g.File.AddShape(g.Sheet, &excelize.Shape{
			Cell:   colName + "3",
			Type:   "rect",
			Width:  uint(lifelineWPx),
			Height: uint(lifelineH),
			Line:   excelize.ShapeLine{Color: "8A8A8A", Width: &lwL},
			Fill:   excelize.Fill{Color: []string{"FFFFFF"}, Pattern: 1},
			Paragraph: []excelize.RichTextRun{{
				Text: " ",
				Font: &excelize.Font{Size: 6, Color: "000000", Family: "Calibri"},
			}},
			Format: excelize.GraphicOptions{Name: "LL_" + p.Name, OffsetX: offLX},
		}); err != nil {
			return fmt.Errorf("lifeline %q: %w", p.Name, err)
		}
	}

	msgCounter := 0
	for i, event := range g.Diagram.Events {
		row := firstMessageRow + i*messageRowStep
		switch ev := event.(type) {
		case *ast.Message:
			fromCol, ok := nameToCol[ev.From.Name]
			if !ok {
				return fmt.Errorf("message %d: unknown sender %q", i, ev.From.Name)
			}
			toCol, ok := nameToCol[ev.To.Name]
			if !ok {
				return fmt.Errorf("message %d: unknown receiver %q", i, ev.To.Name)
			}
			msgCounter++
			fill, fg := messageStyle(ev)
			label := ev.Text
			if g.Diagram.Autonumber {
				label = fmt.Sprintf("%d. %s", msgCounter, label)
			}
			if label == "" {
				label = " "
			}
			if err := g.addDirectionalArrow(fromCol, toCol, row, fmt.Sprintf("msg_%d", i), label, &lwA, fill, fg); err != nil {
				return fmt.Errorf("message %d: %w", i, err)
			}
		case *ast.Note:
			leftCol, ok := nameToCol[ev.Left.Name]
			if !ok {
				return fmt.Errorf("note %d: unknown participant %q", i, ev.Left.Name)
			}
			rightCol := leftCol
			if ev.Right != ev.Left {
				rightCol, ok = nameToCol[ev.Right.Name]
				if !ok {
					return fmt.Errorf("note %d: unknown participant %q", i, ev.Right.Name)
				}
			}
			if err := g.addNote(ev, leftCol, rightCol, row, fmt.Sprintf("note_%d", i)); err != nil {
				return fmt.Errorf("note %d: %w", i, err)
			}
		}
	}

	return nil
}

// addNote renders a note box at the given row.
// Layout rules (lane cols: 2, 5, 8… stride=3, gap=2):
//   - NoteLeft:  box sits in the 2 gap columns immediately left of the participant (col-2..col-1), anchored at col-2
//   - NoteRight: box sits in the 2 gap columns immediately right of the participant, anchored at col+1
//   - NoteOver:  box sits on the participant column(s), anchored at leftCol
func (g *SequenceDrawing) addNote(n *ast.Note, leftCol, rightCol, row int, name string) error {
	const noteH = 30
	noteLW := arrowLinePt

	var anchorCol, offsetX, width int
	switch n.Position {
	case ast.NoteLeft:
		if leftCol > sequenceFirstLaneCol {
			anchorCol = leftCol - sequenceGapColumns
		} else {
			anchorCol = 1
		}
		offsetX = 0
		width = 88
	case ast.NoteRight:
		anchorCol = leftCol
		offsetX = g.laneAnchorOffX(leftCol) + 4
		width = 88
	default: // NoteOver
		anchorCol = leftCol
		offsetX = 0
		if leftCol == rightCol {
			width = 88
		} else {
			w := 0
			for c := leftCol; c <= rightCol; c++ {
				w += g.colWidthPx(c)
			}
			width = w
		}
	}

	colName, err := excelize.ColumnNumberToName(anchorCol)
	if err != nil {
		return err
	}
	cell := fmt.Sprintf("%s%d", colName, row)
	_, r, err := excelize.CellNameToCoordinates(cell)
	if err != nil {
		return err
	}
	offY := g.arrowOffsetY(r, noteH)

	return g.File.AddShape(g.Sheet, &excelize.Shape{
		Cell:   cell,
		Type:   "rect",
		Width:  uint(width),
		Height: noteH,
		Line:   excelize.ShapeLine{Color: "D6B656", Width: &noteLW},
		Fill:   excelize.Fill{Color: []string{"FFF2CC"}, Pattern: 1},
		Paragraph: []excelize.RichTextRun{{
			Text: n.Text,
			Font: &excelize.Font{Size: 9, Color: "1F2A44", Family: "Calibri"},
		}},
		Format: excelize.GraphicOptions{Name: name, OffsetX: offsetX, OffsetY: offY},
	})
}

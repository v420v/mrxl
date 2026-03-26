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

type elseInfo struct {
	label    string
	startRow int
}

type blockRenderInfo struct {
	kind     string
	label    string
	startRow int
	afterRow int // exclusive upper bound (first row after the block)
	depth    int // nesting depth (0 = outermost)
	elses    []elseInfo
}

func collectLeafRows(events []ast.SequenceEvent, nextRow *int) (map[ast.SequenceEvent]int, []blockRenderInfo) {
	rowMap := make(map[ast.SequenceEvent]int)
	var blocks []blockRenderInfo
	collectLeafRowsInto(events, nextRow, rowMap, &blocks, 0)
	return rowMap, blocks
}

func collectLeafRowsInto(events []ast.SequenceEvent, nextRow *int, rowMap map[ast.SequenceEvent]int, blocks *[]blockRenderInfo, depth int) {
	for _, ev := range events {
		blk, ok := ev.(*ast.InteractionBlock)
		if !ok {
			rowMap[ev] = *nextRow
			*nextRow += messageRowStep
			continue
		}
		startRow := *nextRow
		blockIdx := len(*blocks)
		*blocks = append(*blocks, blockRenderInfo{kind: blk.Kind, label: blk.Branches[0].Label, startRow: startRow, depth: depth})
		var elses []elseInfo
		for bi, branch := range blk.Branches {
			if bi > 0 {
				elses = append(elses, elseInfo{branch.Label, *nextRow})
			}
			collectLeafRowsInto(branch.Events, nextRow, rowMap, blocks, depth+1)
		}
		if *nextRow == startRow {
			*nextRow += messageRowStep // ensure non-zero height for empty block
		}
		(*blocks)[blockIdx].afterRow = *nextRow
		(*blocks)[blockIdx].elses = elses
	}
}

var blockStyleMap = map[string][2]string{
	"loop":  {"4472C4", "EBF0F8"},
	"alt":   {"70AD47", "EBF5E4"},
	"opt":   {"ED7D31", "FEF2EA"},
	"break": {"C00000", "FCE4D6"},
}

func blockStyle(kind string) (border, fill string) {
	if s, ok := blockStyleMap[kind]; ok {
		return s[0], s[1]
	}
	return "595959", "F2F2F2"
}

func (g *SequenceDrawing) totalWidthPx(totalCols int) int {
	w := 0
	for c := 1; c <= totalCols; c++ {
		w += g.colWidthPx(c)
	}
	return w
}

func (g *SequenceDrawing) drawBlocks(blocks []blockRenderInfo, totalCols int) error {
	bw := g.totalWidthPx(totalCols)
	borderLW := 1.0
	labelLW := 0.5
	colName, _ := excelize.ColumnNumberToName(1)

	// Phase 1: all background rectangles (outer first → inner on top in z-order).
	for i, b := range blocks {
		border, fill := blockStyle(b.kind)
		cell := fmt.Sprintf("%s%d", colName, b.startRow)
		h := g.sumRowHeightsPx(b.startRow, b.afterRow-1)
		if h < 8 {
			h = 8
		}
		if err := g.File.AddShape(g.Sheet, &excelize.Shape{
			Cell:   cell,
			Type:   "rect",
			Width:  uint(bw),
			Height: uint(h),
			Line:   excelize.ShapeLine{Color: border, Width: &borderLW},
			Fill:   excelize.Fill{Color: []string{fill}, Pattern: 1},
			Paragraph: []excelize.RichTextRun{{Text: " ", Font: &excelize.Font{Size: 1}}},
			Format: excelize.GraphicOptions{Name: fmt.Sprintf("blk_bg_%d", i)},
		}); err != nil {
			return fmt.Errorf("block %q background: %w", b.kind, err)
		}
	}

	// Phase 2: all labels and separators on top of every background.
	// Labels are offset vertically by depth so nested labels don't overlap.
	for i, b := range blocks {
		border, fill := blockStyle(b.kind)
		cell := fmt.Sprintf("%s%d", colName, b.startRow)
		labelOffY := 2 + b.depth*20

		kindLabel := fmt.Sprintf("[%s]", b.kind)
		if b.label != "" {
			kindLabel = fmt.Sprintf("[%s] %s", b.kind, b.label)
		}
		labelW := min(bw-4, 160)
		if err := g.File.AddShape(g.Sheet, &excelize.Shape{
			Cell:   cell,
			Type:   "rect",
			Width:  uint(labelW),
			Height: 18,
			Line:   excelize.ShapeLine{Color: border, Width: &labelLW},
			Fill:   excelize.Fill{Color: []string{border}, Pattern: 1},
			Paragraph: []excelize.RichTextRun{{
				Text: kindLabel,
				Font: &excelize.Font{Bold: true, Size: 8, Color: "FFFFFF", Family: "Calibri"},
			}},
			Format: excelize.GraphicOptions{Name: fmt.Sprintf("blk_lbl_%d", i), OffsetX: 2, OffsetY: labelOffY},
		}); err != nil {
			return fmt.Errorf("block %q label: %w", b.kind, err)
		}

		for j, el := range b.elses {
			elseCell := fmt.Sprintf("%s%d", colName, el.startRow)
			sepLW := 0.75
			if err := g.File.AddShape(g.Sheet, &excelize.Shape{
				Cell:   elseCell,
				Type:   "rect",
				Width:  uint(bw),
				Height: 2,
				Line:   excelize.ShapeLine{Color: border, Width: &sepLW},
				Fill:   excelize.Fill{Color: []string{border}, Pattern: 1},
				Paragraph: []excelize.RichTextRun{{Text: " ", Font: &excelize.Font{Size: 1}}},
				Format: excelize.GraphicOptions{Name: fmt.Sprintf("blk_sep_%d_%d", i, j)},
			}); err != nil {
				return fmt.Errorf("block %q else separator: %w", b.kind, err)
			}
			elseLabel := "[else]"
			if el.label != "" {
				elseLabel = fmt.Sprintf("[else] %s", el.label)
			}
			if err := g.File.AddShape(g.Sheet, &excelize.Shape{
				Cell:   elseCell,
				Type:   "rect",
				Width:  uint(min(bw-4, 140)),
				Height: 16,
				Line:   excelize.ShapeLine{Color: border, Width: &labelLW},
				Fill:   excelize.Fill{Color: []string{fill}, Pattern: 1},
				Paragraph: []excelize.RichTextRun{{
					Text: elseLabel,
					Font: &excelize.Font{Italic: true, Size: 8, Color: border, Family: "Calibri"},
				}},
				Format: excelize.GraphicOptions{Name: fmt.Sprintf("blk_else_%d_%d", i, j), OffsetX: 4, OffsetY: 3},
			}); err != nil {
				return fmt.Errorf("block %q else label: %w", b.kind, err)
			}
		}
	}
	return nil
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

	// Assign rows to all leaf events and collect block render info.
	nextRow := firstMessageRow
	rowMap, blockInfos := collectLeafRows(g.Diagram.Events, &nextRow)

	nRows := max(nextRow+5, firstMessageRow+10)
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

	lastEventRow := nextRow - messageRowStep
	if lastEventRow < firstMessageRow {
		lastEventRow = firstMessageRow
	}
	lifelineEndRow := max(lastEventRow+3, 10)
	lifelineH := g.sumRowHeightsPx(3, lifelineEndRow)

	// Draw block backgrounds first so all other shapes render on top.
	totalCols := sequenceSheetColumnCount(len(g.Diagram.Participants))
	if err := g.drawBlocks(blockInfos, totalCols); err != nil {
		return err
	}

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

	if err := g.drawActivations(nameToCol, rowMap, lifelineEndRow); err != nil {
		return err
	}

	shapeIdx := 0
	var drawEvents func([]ast.SequenceEvent) error
	drawEvents = func(events []ast.SequenceEvent) error {
		for _, event := range events {
			switch ev := event.(type) {
			case *ast.InteractionBlock:
				for _, branch := range ev.Branches {
					if err := drawEvents(branch.Events); err != nil {
						return err
					}
				}
			case *ast.Activation:
				// handled by drawActivations
			case *ast.Message:
				row := rowMap[ev]
				fromCol, ok := nameToCol[ev.From.Name]
				if !ok {
					return fmt.Errorf("message: unknown sender %q", ev.From.Name)
				}
				toCol, ok := nameToCol[ev.To.Name]
				if !ok {
					return fmt.Errorf("message: unknown receiver %q", ev.To.Name)
				}
				fill, fg := messageStyle(ev)
				label := ev.Text
				if g.Diagram.Autonumber {
					shapeIdx++
					label = fmt.Sprintf("%d. %s", shapeIdx, label)
				}
				if label == "" {
					label = " "
				}
				if err := g.addDirectionalArrow(fromCol, toCol, row, fmt.Sprintf("msg_%d", shapeIdx), label, &lwA, fill, fg); err != nil {
					return fmt.Errorf("message: %w", err)
				}
				shapeIdx++
			case *ast.Note:
				row := rowMap[ev]
				leftCol, ok := nameToCol[ev.Left.Name]
				if !ok {
					return fmt.Errorf("note: unknown participant %q", ev.Left.Name)
				}
				rightCol := leftCol
				if ev.Right != ev.Left {
					rightCol, ok = nameToCol[ev.Right.Name]
					if !ok {
						return fmt.Errorf("note: unknown participant %q", ev.Right.Name)
					}
				}
				if err := g.addNote(ev, leftCol, rightCol, row, fmt.Sprintf("note_%d", shapeIdx)); err != nil {
					return fmt.Errorf("note: %w", err)
				}
				shapeIdx++
			}
		}
		return nil
	}
	return drawEvents(g.Diagram.Events)
}

type activationSpan struct {
	col      int
	startRow int
	endRow   int
	depth    int // nesting depth (0 = outermost)
}

// computeActivationSpans walks events and returns a span for each matched activate/deactivate pair.
// Unclosed activations are closed at lifelineEndRow.
func computeActivationSpans(events []ast.SequenceEvent, nameToCol map[string]int, rowMap map[ast.SequenceEvent]int, lifelineEndRow int) []activationSpan {
	type stackEntry struct{ startRow, depth int }
	stacks := make(map[int][]stackEntry) // col -> stack
	var spans []activationSpan

	var walk func([]ast.SequenceEvent)
	walk = func(evs []ast.SequenceEvent) {
		for _, ev := range evs {
			switch e := ev.(type) {
			case *ast.InteractionBlock:
				for _, branch := range e.Branches {
					walk(branch.Events)
				}
			case *ast.Activation:
				row, ok := rowMap[e]
				if !ok {
					continue
				}
				col, exists := nameToCol[e.Participant.Name]
				if !exists {
					continue
				}
				if e.Active {
					depth := len(stacks[col])
					stacks[col] = append(stacks[col], stackEntry{row, depth})
				} else {
					stk := stacks[col]
					if len(stk) > 0 {
						top := stk[len(stk)-1]
						stacks[col] = stk[:len(stk)-1]
						spans = append(spans, activationSpan{col, top.startRow, row, top.depth})
					}
				}
			}
		}
	}
	walk(events)

	// Close any unclosed activations.
	for col, stk := range stacks {
		for len(stk) > 0 {
			top := stk[len(stk)-1]
			stk = stk[:len(stk)-1]
			spans = append(spans, activationSpan{col, top.startRow, lifelineEndRow, top.depth})
		}
	}
	return spans
}

func (g *SequenceDrawing) drawActivations(nameToCol map[string]int, rowMap map[ast.SequenceEvent]int, lifelineEndRow int) error {
	spans := computeActivationSpans(g.Diagram.Events, nameToCol, rowMap, lifelineEndRow)
	const (
		activationW = 10
		lifelineW   = 5
		depthOffset = 4
	)
	activLW := 0.5
	for i, span := range spans {
		colName, err := excelize.ColumnNumberToName(span.col)
		if err != nil {
			return err
		}
		cell := fmt.Sprintf("%s%d", colName, span.startRow)
		h := g.sumRowHeightsPx(span.startRow, span.endRow)
		if h < 4 {
			h = 4
		}
		centerX := g.laneAnchorOffX(span.col) + lifelineW/2
		offX := centerX - activationW/2 + span.depth*depthOffset
		if offX < 0 {
			offX = 0
		}
		if err := g.File.AddShape(g.Sheet, &excelize.Shape{
			Cell:   cell,
			Type:   "rect",
			Width:  activationW,
			Height: uint(h),
			Line:   excelize.ShapeLine{Color: "2F5597", Width: &activLW},
			Fill:   excelize.Fill{Color: []string{"BDD7EE"}, Pattern: 1},
			Paragraph: []excelize.RichTextRun{{
				Text: " ",
				Font: &excelize.Font{Size: 1},
			}},
			Format: excelize.GraphicOptions{
				Name:    fmt.Sprintf("act_%d", i),
				OffsetX: offX,
			},
		}); err != nil {
			return fmt.Errorf("activation span %d: %w", i, err)
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

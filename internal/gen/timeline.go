package gen

import (
	"fmt"
	"math"
	"strings"

	"github.com/v420v/mrxl/internal/ast"
	"github.com/xuri/excelize/v2"
)

const (
	timelineFirstEventColIdx = 2   // column index (1-based) of the first event lane
	timelineEventStride      = 2   // one wide event col + one narrow gap col per slot
	timelineEventColWidth    = 14.0
	timelineGapColWidth      = 2.0
	timelineMarginColWidth   = 2.0

	// row indices (1-based)
	timelineTitleRow   = 1
	timelineSectionRow = 2
	timelineSpacerRow  = 3
	timelineTimeRow    = 4
	timelineArrowRow   = 5
	timelineEventRow   = 6
	timelineBottomRow  = 7

	// row heights in Excel pt units
	timelineTitlePt   = 28.0
	timelineSectionPt = 40.0
	timelineSpacerPt  = 12.0
	timelineTimePt    = 40.0
	timelineArrowPt   = 24.0
	timelineEventPt   = 40.0
	timelineBottomPt  = 24.0

	// shape heights in pixels
	timelineSectionBoxH = 36
	timelineSlotBoxH    = 36
	timelineArrowShapeH = 12
	timelineConnW       = 2
)

type tlColor struct{ fill, text, border string }

var timelinePalette = []tlColor{
	{"7B7FC4", "FFFFFF", "5A5FA0"},
	{"F5F566", "444444", "C5C530"},
	{"7BC47B", "FFFFFF", "5AA058"},
	{"F5A566", "444444", "C87830"},
}

// TimelineDrawing renders a TimeDiagram onto an Excel sheet.
type TimelineDrawing struct {
	File    *excelize.File
	Sheet   string
	Diagram *ast.TimeDiagram
}

func tlColor_(sectionIdx int) tlColor {
	return timelinePalette[sectionIdx%len(timelinePalette)]
}

func tlRowPx(ptHeight float64) int {
	return int(math.Ceil(4.0 / 3.4 * ptHeight))
}

func tlColPx(w float64) int {
	return int(w*8 + 0.5)
}

func tlEventColNum(globalIdx int) int {
	return timelineFirstEventColIdx + globalIdx*timelineEventStride
}

// tlSpanPx returns the pixel width that spans n event slots (no trailing gap).
func tlSpanPx(numEvents int) int {
	if numEvents <= 0 {
		return 0
	}
	ew := tlColPx(timelineEventColWidth)
	gw := tlColPx(timelineGapColWidth)
	return numEvents*ew + (numEvents-1)*gw
}

func (g *TimelineDrawing) drawTimeline() error {
	d := g.Diagram
	f := g.File

	totalEvents := 0
	for _, sec := range d.Sections {
		totalEvents += len(sec.Events)
	}
	if totalEvents == 0 {
		return fmt.Errorf("timeline: no events")
	}

	// --- Row heights ---
	for _, rh := range []struct {
		row int
		pt  float64
	}{
		{timelineTitleRow, timelineTitlePt},
		{timelineSectionRow, timelineSectionPt},
		{timelineSpacerRow, timelineSpacerPt},
		{timelineTimeRow, timelineTimePt},
		{timelineArrowRow, timelineArrowPt},
		{timelineEventRow, timelineEventPt},
		{timelineBottomRow, timelineBottomPt},
	} {
		if err := f.SetRowHeight(g.Sheet, rh.row, rh.pt); err != nil {
			return fmt.Errorf("set row %d height: %w", rh.row, err)
		}
	}

	// --- Column widths ---
	marginName, _ := excelize.ColumnNumberToName(1)
	if err := f.SetColWidth(g.Sheet, marginName, marginName, timelineMarginColWidth); err != nil {
		return fmt.Errorf("set margin col width: %w", err)
	}
	for i := 0; i < totalEvents; i++ {
		evtNum := tlEventColNum(i)
		gapNum := evtNum + 1
		evtName, _ := excelize.ColumnNumberToName(evtNum)
		gapName, _ := excelize.ColumnNumberToName(gapNum)
		if err := f.SetColWidth(g.Sheet, evtName, evtName, timelineEventColWidth); err != nil {
			return fmt.Errorf("set event col %d width: %w", i, err)
		}
		if err := f.SetColWidth(g.Sheet, gapName, gapName, timelineGapColWidth); err != nil {
			return fmt.Errorf("set gap col after %d width: %w", i, err)
		}
	}

	// --- Title ---
	if d.Title != "" {
		firstCol, _ := excelize.ColumnNumberToName(timelineFirstEventColIdx)
		lastCol, _ := excelize.ColumnNumberToName(tlEventColNum(totalEvents - 1))
		titleCell := fmt.Sprintf("%s%d", firstCol, timelineTitleRow)
		lastCell := fmt.Sprintf("%s%d", lastCol, timelineTitleRow)
		if err := f.MergeCell(g.Sheet, titleCell, lastCell); err != nil {
			return fmt.Errorf("merge title cells: %w", err)
		}
		if err := f.SetCellValue(g.Sheet, titleCell, d.Title); err != nil {
			return fmt.Errorf("set title value: %w", err)
		}
		style, err := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 14, Color: "333333", Family: "Calibri"},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		})
		if err != nil {
			return fmt.Errorf("create title style: %w", err)
		}
		if err := f.SetCellStyle(g.Sheet, titleCell, lastCell, style); err != nil {
			return fmt.Errorf("set title style: %w", err)
		}
	}

	// Precompute pixel metrics used across all shapes.
	sectionRowPx := tlRowPx(timelineSectionPt) // 48
	timeRowPx := tlRowPx(timelineTimePt)        // 48
	arrowRowPx := tlRowPx(timelineArrowPt)      // 29
	eventRowPx := tlRowPx(timelineEventPt)      // 48
	evtColPx := tlColPx(timelineEventColWidth)  // 112

	sectionBoxOffY := (sectionRowPx - timelineSectionBoxH) / 2
	slotBoxOffY := (timeRowPx - timelineSlotBoxH) / 2
	arrowOffY := (arrowRowPx - timelineArrowShapeH) / 2

	// Connector: spans from center of time row down through arrow row to center of event row.
	connOffX := evtColPx/2 - timelineConnW/2
	connOffY := timeRowPx / 2
	connH := timeRowPx/2 + arrowRowPx + eventRowPx/2

	lwBox := 1.5
	lwConn := 0.25

	// --- Phase 1: section header boxes (row 2) ---
	globalIdx := 0
	for sectionIdx, sec := range d.Sections {
		color := tlColor_(sectionIdx)
		numEvts := len(sec.Events)
		firstColNum := tlEventColNum(globalIdx)
		firstColName, _ := excelize.ColumnNumberToName(firstColNum)

		if sec.Name != "" {
			secCell := fmt.Sprintf("%s%d", firstColName, timelineSectionRow)
			if err := f.AddShape(g.Sheet, &excelize.Shape{
				Cell:   secCell,
				Type:   "rect",
				Width:  uint(tlSpanPx(numEvts)),
				Height: uint(timelineSectionBoxH),
				Line:   excelize.ShapeLine{Color: color.border, Width: &lwBox},
				Fill:   excelize.Fill{Color: []string{color.fill}, Pattern: 1},
				Paragraph: []excelize.RichTextRun{{
					Text: sec.Name,
					Font: &excelize.Font{Bold: true, Size: 12, Color: color.text, Family: "Calibri"},
				}},
				Format: excelize.GraphicOptions{
					Name:    fmt.Sprintf("sec_%d", sectionIdx),
					OffsetY: sectionBoxOffY,
				},
			}); err != nil {
				return fmt.Errorf("section %q box: %w", sec.Name, err)
			}
		}
		globalIdx += numEvts
	}

	// --- Phase 2: vertical connectors (behind time/event boxes and arrow) ---
	globalIdx = 0
	for sectionIdx, sec := range d.Sections {
		for evtIdx, _ := range sec.Events {
			evtColName, _ := excelize.ColumnNumberToName(tlEventColNum(globalIdx))
			connCell := fmt.Sprintf("%s%d", evtColName, timelineTimeRow)
			if err := f.AddShape(g.Sheet, &excelize.Shape{
				Cell:   connCell,
				Type:   "rect",
				Width:  uint(timelineConnW),
				Height: uint(connH),
				Line:   excelize.ShapeLine{Color: "AAAAAA", Width: &lwConn},
				Fill:   excelize.Fill{Color: []string{"AAAAAA"}, Pattern: 1},
				Paragraph: []excelize.RichTextRun{{
					Text: " ",
					Font: &excelize.Font{Size: 1},
				}},
				Format: excelize.GraphicOptions{
					Name:    fmt.Sprintf("conn_%d_%d", sectionIdx, evtIdx),
					OffsetX: connOffX,
					OffsetY: connOffY,
				},
			}); err != nil {
				return fmt.Errorf("connector %d_%d: %w", sectionIdx, evtIdx, err)
			}
			globalIdx++
		}
	}

	// --- Phase 3: horizontal timeline arrow (row 5) ---
	firstEvtColName, _ := excelize.ColumnNumberToName(timelineFirstEventColIdx)
	arrowCell := fmt.Sprintf("%s%d", firstEvtColName, timelineArrowRow)
	lwArrow := 0.25
	if err := f.AddShape(g.Sheet, &excelize.Shape{
		Cell:   arrowCell,
		Type:   "rightArrow",
		Width:  uint(tlSpanPx(totalEvents)),
		Height: uint(timelineArrowShapeH),
		Line:   excelize.ShapeLine{Color: "333333", Width: &lwArrow},
		Fill:   excelize.Fill{Color: []string{"333333"}, Pattern: 1},
		Paragraph: []excelize.RichTextRun{{
			Text: " ",
			Font: &excelize.Font{Size: 1},
		}},
		Format: excelize.GraphicOptions{
			Name:    "timeline_arrow",
			OffsetY: arrowOffY,
		},
	}); err != nil {
		return fmt.Errorf("timeline arrow: %w", err)
	}

	// --- Phase 4: time label and event description boxes ---
	globalIdx = 0
	for sectionIdx, sec := range d.Sections {
		color := tlColor_(sectionIdx)
		for evtIdx, evt := range sec.Events {
			evtColName, _ := excelize.ColumnNumberToName(tlEventColNum(globalIdx))
			boxW := uint(evtColPx - 4) // 2px margin each side

			timeCell := fmt.Sprintf("%s%d", evtColName, timelineTimeRow)
			timeText := evt.Time
			if timeText == "" {
				timeText = " "
			}
			if err := f.AddShape(g.Sheet, &excelize.Shape{
				Cell:   timeCell,
				Type:   "rect",
				Width:  boxW,
				Height: uint(timelineSlotBoxH),
				Line:   excelize.ShapeLine{Color: color.border, Width: &lwBox},
				Fill:   excelize.Fill{Color: []string{color.fill}, Pattern: 1},
				Paragraph: []excelize.RichTextRun{{
					Text: timeText,
					Font: &excelize.Font{Bold: true, Size: 11, Color: color.text, Family: "Calibri"},
				}},
				Format: excelize.GraphicOptions{
					Name:    fmt.Sprintf("time_%d_%d", sectionIdx, evtIdx),
					OffsetX: 2,
					OffsetY: slotBoxOffY,
				},
			}); err != nil {
				return fmt.Errorf("time label %q: %w", evt.Time, err)
			}

			evtCell := fmt.Sprintf("%s%d", evtColName, timelineEventRow)
			evtText := strings.Join(evt.Texts, "\n")
			if evtText == "" {
				evtText = " "
			}
			if err := f.AddShape(g.Sheet, &excelize.Shape{
				Cell:   evtCell,
				Type:   "rect",
				Width:  boxW,
				Height: uint(timelineSlotBoxH),
				Line:   excelize.ShapeLine{Color: color.border, Width: &lwBox},
				Fill:   excelize.Fill{Color: []string{color.fill}, Pattern: 1},
				Paragraph: []excelize.RichTextRun{{
					Text: evtText,
					Font: &excelize.Font{Bold: false, Size: 10, Color: color.text, Family: "Calibri"},
				}},
				Format: excelize.GraphicOptions{
					Name:    fmt.Sprintf("evt_%d_%d", sectionIdx, evtIdx),
					OffsetX: 2,
					OffsetY: slotBoxOffY,
				},
			}); err != nil {
				return fmt.Errorf("event description %q: %w", evtText, err)
			}

			globalIdx++
		}
	}

	return nil
}

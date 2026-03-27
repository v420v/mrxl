package gen

import (
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/v420v/mrxl/internal/ast"
	"github.com/xuri/excelize/v2"
)

const (
	journeyFirstTaskColIdx    = 2    // column index (1-based) of first task lane
	journeyTaskStride         = 2    // one wide task col + one narrow gap col per slot
	journeyTaskColWidth       = 14.0
	journeyGapColWidth_       = 2.0
	journeyMarginColWidth_    = 2.0
	journeyActorLabelColWidth = 12.0 // column 1 width when actors are present

	// row indices (1-based)
	journeyTitleRow   = 1
	journeySectionRow = 2
	journeySpacerRow  = 3
	journeyTaskRow    = 4
	journeyArrowRow   = 5
	journeyScoreRow   = 6
	// actor rows start at journeyScoreRow+1 and are computed dynamically

	// row heights in Excel pt units
	journeyTitlePt   = 28.0
	journeySectionPt = 40.0
	journeySpacerPt  = 12.0
	journeyTaskPt    = 40.0
	journeyArrowPt   = 24.0
	journeyScorePt   = 40.0
	journeyActorPt   = 30.0
	journeyBottomPt  = 24.0

	// shape heights in pixels
	journeySectionBoxH = 36
	journeySlotBoxH    = 36
	journeyActorBoxH   = 22
	journeyArrowShapeH = 12
	journeyConnW_      = 2
)

type journeySecColor struct{ sectionFill, taskFill, text, border string }

// Light pastel palette matching Mermaid's user journey section colors.
var journeyPalette = []journeySecColor{
	{"C8C8F0", "DDDDF8", "333333", "8080C0"}, // blue-purple
	{"F0F0C0", "F8F8E0", "333333", "B0B010"}, // yellow
	{"F5E6F5", "FAF0FA", "333333", "C080C0"}, // lavender/pink
	{"C8F0C8", "E0F8E0", "333333", "40A040"}, // green
}

// JourneyDrawing renders a UserJourneyDiagram as a timeline-style layout.
type JourneyDrawing struct {
	File    *excelize.File
	Sheet   string
	Diagram *ast.UserJourneyDiagram
}

func journeySecColor_(idx int) journeySecColor {
	return journeyPalette[idx%len(journeyPalette)]
}

func journeyRowPx(pt float64) int {
	return int(math.Ceil(4.0 / 3.4 * pt))
}

func journeyColPx_(w float64) int {
	return int(w*8 + 0.5)
}

func journeyTaskColNum(globalIdx int) int {
	return journeyFirstTaskColIdx + globalIdx*journeyTaskStride
}

func journeySpanPx(n int) int {
	if n <= 0 {
		return 0
	}
	ew := journeyColPx_(journeyTaskColWidth)
	gw := journeyColPx_(journeyGapColWidth_)
	return n*ew + (n-1)*gw
}

// journeyScoreStyle returns fill/text colors based on score (1–5 scale).
// green (happy) ≥ 4, yellow (neutral) = 3, red (sad) ≤ 2
func journeyScoreStyle(score float64) (fill, text string) {
	switch {
	case score >= 4:
		return "6AAF4A", "FFFFFF"
	case score >= 3:
		return "F0D040", "333333"
	default:
		return "D45555", "FFFFFF"
	}
}

func (g *JourneyDrawing) drawUserJourney() error {
	d := g.Diagram
	f := g.File

	totalTasks := 0
	for _, sec := range d.Sections {
		totalTasks += len(sec.Tasks)
	}
	if totalTasks == 0 {
		return fmt.Errorf("user journey: no tasks")
	}

	// Collect unique actors in order of first appearance.
	var actorOrder []string
	actorSeen := map[string]bool{}
	for _, sec := range d.Sections {
		for _, task := range sec.Tasks {
			for _, a := range task.Actors {
				if !actorSeen[a] {
					actorSeen[a] = true
					actorOrder = append(actorOrder, a)
				}
			}
		}
	}
	numActors := len(actorOrder)
	firstActorRow := journeyScoreRow + 1
	bottomRow := firstActorRow + numActors

	// --- Row heights ---
	for _, rh := range []struct {
		row int
		pt  float64
	}{
		{journeyTitleRow, journeyTitlePt},
		{journeySectionRow, journeySectionPt},
		{journeySpacerRow, journeySpacerPt},
		{journeyTaskRow, journeyTaskPt},
		{journeyArrowRow, journeyArrowPt},
		{journeyScoreRow, journeyScorePt},
	} {
		if err := f.SetRowHeight(g.Sheet, rh.row, rh.pt); err != nil {
			return fmt.Errorf("set row %d height: %w", rh.row, err)
		}
	}
	for i := range numActors {
		if err := f.SetRowHeight(g.Sheet, firstActorRow+i, journeyActorPt); err != nil {
			return fmt.Errorf("set actor row %d height: %w", i, err)
		}
	}
	if err := f.SetRowHeight(g.Sheet, bottomRow, journeyBottomPt); err != nil {
		return fmt.Errorf("set bottom row height: %w", err)
	}

	// --- Column widths ---
	marginName, _ := excelize.ColumnNumberToName(1)
	labelColWidth := journeyMarginColWidth_
	if numActors > 0 {
		labelColWidth = journeyActorLabelColWidth
	}
	if err := f.SetColWidth(g.Sheet, marginName, marginName, labelColWidth); err != nil {
		return fmt.Errorf("set margin col width: %w", err)
	}
	for i := 0; i < totalTasks; i++ {
		evtNum := journeyTaskColNum(i)
		gapNum := evtNum + 1
		evtName, _ := excelize.ColumnNumberToName(evtNum)
		gapName, _ := excelize.ColumnNumberToName(gapNum)
		if err := f.SetColWidth(g.Sheet, evtName, evtName, journeyTaskColWidth); err != nil {
			return fmt.Errorf("set task col %d width: %w", i, err)
		}
		if err := f.SetColWidth(g.Sheet, gapName, gapName, journeyGapColWidth_); err != nil {
			return fmt.Errorf("set gap col %d width: %w", i, err)
		}
	}

	// --- Title ---
	if d.Title != "" {
		firstCol, _ := excelize.ColumnNumberToName(journeyFirstTaskColIdx)
		lastCol, _ := excelize.ColumnNumberToName(journeyTaskColNum(totalTasks - 1))
		titleCell := fmt.Sprintf("%s%d", firstCol, journeyTitleRow)
		lastCell := fmt.Sprintf("%s%d", lastCol, journeyTitleRow)
		if err := f.MergeCell(g.Sheet, titleCell, lastCell); err != nil {
			return fmt.Errorf("merge title cells: %w", err)
		}
		if err := f.SetCellValue(g.Sheet, titleCell, d.Title); err != nil {
			return fmt.Errorf("set title: %w", err)
		}
		style, err := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 16, Color: "222222", Family: "Calibri"},
			Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		})
		if err != nil {
			return fmt.Errorf("create title style: %w", err)
		}
		if err := f.SetCellStyle(g.Sheet, titleCell, lastCell, style); err != nil {
			return fmt.Errorf("set title style: %w", err)
		}
	}

	// Precompute pixel metrics.
	sectionRowPx := journeyRowPx(journeySectionPt) // 48
	taskRowPx := journeyRowPx(journeyTaskPt)        // 48
	arrowRowPx := journeyRowPx(journeyArrowPt)      // 29
	scoreRowPx := journeyRowPx(journeyScorePt)      // 48
	evtColPx := journeyColPx_(journeyTaskColWidth)  // 112

	sectionBoxOffY := (sectionRowPx - journeySectionBoxH) / 2
	slotBoxOffY := (taskRowPx - journeySlotBoxH) / 2
	arrowOffY := (arrowRowPx - journeyArrowShapeH) / 2

	connOffX := evtColPx/2 - journeyConnW_/2
	connOffY := taskRowPx / 2
	connH := taskRowPx/2 + arrowRowPx + scoreRowPx/2

	lw := 1.5
	lwConn := 0.25

	// --- Phase 1: Section header boxes (row 2) ---
	globalIdx := 0
	for sectionIdx, sec := range d.Sections {
		color := journeySecColor_(sectionIdx)
		numTasks := len(sec.Tasks)
		firstColNum := journeyTaskColNum(globalIdx)
		firstColName, _ := excelize.ColumnNumberToName(firstColNum)

		if sec.Name != "" {
			secCell := fmt.Sprintf("%s%d", firstColName, journeySectionRow)
			if err := f.AddShape(g.Sheet, &excelize.Shape{
				Cell:   secCell,
				Type:   "rect",
				Width:  uint(journeySpanPx(numTasks)),
				Height: uint(journeySectionBoxH),
				Line:   excelize.ShapeLine{Color: color.border, Width: &lw},
				Fill:   excelize.Fill{Color: []string{color.sectionFill}, Pattern: 1},
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
		globalIdx += numTasks
	}

	// --- Phase 2: Vertical connectors (behind task/score boxes and arrow) ---
	globalIdx = 0
	for sectionIdx, sec := range d.Sections {
		for taskIdx := range sec.Tasks {
			colName, _ := excelize.ColumnNumberToName(journeyTaskColNum(globalIdx))
			connCell := fmt.Sprintf("%s%d", colName, journeyTaskRow)
			if err := f.AddShape(g.Sheet, &excelize.Shape{
				Cell:   connCell,
				Type:   "rect",
				Width:  uint(journeyConnW_),
				Height: uint(connH),
				Line:   excelize.ShapeLine{Color: "BBBBBB", Width: &lwConn},
				Fill:   excelize.Fill{Color: []string{"BBBBBB"}, Pattern: 1},
				Paragraph: []excelize.RichTextRun{{
					Text: " ",
					Font: &excelize.Font{Size: 1},
				}},
				Format: excelize.GraphicOptions{
					Name:    fmt.Sprintf("conn_%d_%d", sectionIdx, taskIdx),
					OffsetX: connOffX,
					OffsetY: connOffY,
				},
			}); err != nil {
				return fmt.Errorf("connector %d_%d: %w", sectionIdx, taskIdx, err)
			}
			globalIdx++
		}
	}

	// --- Phase 3: Horizontal timeline arrow (row 5) ---
	firstTaskColName, _ := excelize.ColumnNumberToName(journeyFirstTaskColIdx)
	arrowCell := fmt.Sprintf("%s%d", firstTaskColName, journeyArrowRow)
	lwArrow := 0.25
	if err := f.AddShape(g.Sheet, &excelize.Shape{
		Cell:   arrowCell,
		Type:   "rightArrow",
		Width:  uint(journeySpanPx(totalTasks)),
		Height: uint(journeyArrowShapeH),
		Line:   excelize.ShapeLine{Color: "333333", Width: &lwArrow},
		Fill:   excelize.Fill{Color: []string{"333333"}, Pattern: 1},
		Paragraph: []excelize.RichTextRun{{
			Text: " ",
			Font: &excelize.Font{Size: 1},
		}},
		Format: excelize.GraphicOptions{
			Name:    "journey_arrow",
			OffsetY: arrowOffY,
		},
	}); err != nil {
		return fmt.Errorf("journey arrow: %w", err)
	}

	// --- Phase 4: Task name boxes (row 4) and score indicator circles (row 6) ---
	globalIdx = 0
	for sectionIdx, sec := range d.Sections {
		color := journeySecColor_(sectionIdx)
		for taskIdx, task := range sec.Tasks {
			colName, _ := excelize.ColumnNumberToName(journeyTaskColNum(globalIdx))
			boxW := uint(evtColPx - 4) // 2px margin each side

			// Task name box
			taskCell := fmt.Sprintf("%s%d", colName, journeyTaskRow)
			taskText := task.Title
			if taskText == "" {
				taskText = " "
			}
			if err := f.AddShape(g.Sheet, &excelize.Shape{
				Cell:   taskCell,
				Type:   "rect",
				Width:  boxW,
				Height: uint(journeySlotBoxH),
				Line:   excelize.ShapeLine{Color: color.border, Width: &lw},
				Fill:   excelize.Fill{Color: []string{color.taskFill}, Pattern: 1},
				Paragraph: []excelize.RichTextRun{{
					Text: taskText,
					Font: &excelize.Font{Bold: false, Size: 10, Color: color.text, Family: "Calibri"},
				}},
				Format: excelize.GraphicOptions{
					Name:    fmt.Sprintf("task_%d_%d", sectionIdx, taskIdx),
					OffsetX: 2,
					OffsetY: slotBoxOffY,
				},
			}); err != nil {
				return fmt.Errorf("task %q box: %w", task.Title, err)
			}

			// Score indicator: colored circle with score number
			scoreFill, scoreTextColor := journeyScoreStyle(task.Score)
			scoreCell := fmt.Sprintf("%s%d", colName, journeyScoreRow)
			scoreStr := strconv.FormatFloat(task.Score, 'f', -1, 64)
			scoreOffY := (scoreRowPx - journeySlotBoxH) / 2
			if err := f.AddShape(g.Sheet, &excelize.Shape{
				Cell:   scoreCell,
				Type:   "ellipse",
				Width:  boxW,
				Height: uint(journeySlotBoxH),
				Line:   excelize.ShapeLine{Color: scoreFill, Width: &lw},
				Fill:   excelize.Fill{Color: []string{scoreFill}, Pattern: 1},
				Paragraph: []excelize.RichTextRun{{
					Text: scoreStr,
					Font: &excelize.Font{Bold: true, Size: 14, Color: scoreTextColor, Family: "Calibri"},
				}},
				Format: excelize.GraphicOptions{
					Name:    fmt.Sprintf("score_%d_%d", sectionIdx, taskIdx),
					OffsetX: 2,
					OffsetY: scoreOffY,
				},
			}); err != nil {
				return fmt.Errorf("score for %q: %w", task.Title, err)
			}

			globalIdx++
		}
	}

	// --- Phase 5: Actor rows (one row per unique actor, below score row) ---
	if numActors > 0 {
		actorRowPx := journeyRowPx(journeyActorPt)
		actorBoxOffY := (actorRowPx - journeyActorBoxH) / 2

		actorLabelStyle, err := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Size: 10, Color: "444444", Family: "Calibri"},
			Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		})
		if err != nil {
			return fmt.Errorf("create actor label style: %w", err)
		}

		for ai, actor := range actorOrder {
			actorRow := firstActorRow + ai

			// Actor name label in column 1
			labelCell, _ := excelize.CoordinatesToCellName(1, actorRow)
			if err := f.SetCellValue(g.Sheet, labelCell, actor); err != nil {
				return fmt.Errorf("set actor label %q: %w", actor, err)
			}
			if err := f.SetCellStyle(g.Sheet, labelCell, labelCell, actorLabelStyle); err != nil {
				return fmt.Errorf("set actor label style: %w", err)
			}

			// Participation indicator per task
			globalIdx = 0
			for sectionIdx, sec := range d.Sections {
				color := journeySecColor_(sectionIdx)
				for taskIdx, task := range sec.Tasks {
					colName, _ := excelize.ColumnNumberToName(journeyTaskColNum(globalIdx))
					cellRef := fmt.Sprintf("%s%d", colName, actorRow)

					participates := slices.Contains(task.Actors, actor)

					if participates {
						boxW := uint(evtColPx - 4)
						if err := f.AddShape(g.Sheet, &excelize.Shape{
							Cell:   cellRef,
							Type:   "roundRect",
							Width:  boxW,
							Height: uint(journeyActorBoxH),
							Line:   excelize.ShapeLine{Color: color.border, Width: &lw},
							Fill:   excelize.Fill{Color: []string{color.taskFill}, Pattern: 1},
							Paragraph: []excelize.RichTextRun{{
								Text: " ",
								Font: &excelize.Font{Size: 1},
							}},
							Format: excelize.GraphicOptions{
								Name:    fmt.Sprintf("actor_%d_%d_%d", ai, sectionIdx, taskIdx),
								OffsetX: 2,
								OffsetY: actorBoxOffY,
							},
						}); err != nil {
							return fmt.Errorf("actor %q indicator for task %q: %w", actor, task.Title, err)
						}
					}
					globalIdx++
				}
			}
		}
	}

	return nil
}

package gen

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/v420v/mrxl/internal/ast"
	"github.com/xuri/excelize/v2"
)

// mermaidDateToGo converts a Mermaid dateFormat string to a Go time layout string.
func mermaidDateToGo(mermaidFmt string) string {
	r := mermaidFmt
	r = strings.ReplaceAll(r, "YYYY", "2006")
	r = strings.ReplaceAll(r, "YY", "06")
	r = strings.ReplaceAll(r, "MM", "01")
	r = strings.ReplaceAll(r, "DD", "02")
	r = strings.ReplaceAll(r, "HH", "15")
	r = strings.ReplaceAll(r, "mm", "04")
	r = strings.ReplaceAll(r, "ss", "05")
	return r
}

// ganttParseDuration parses "Xd", "Xw", or "Xh" into a number of days.
func ganttParseDuration(s string) (int, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	unit := s[len(s)-1]
	n, err := strconv.Atoi(s[:len(s)-1])
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	switch unit {
	case 'd':
		return n, nil
	case 'w':
		return n * 7, nil
	case 'h':
		return int(math.Ceil(float64(n) / 24.0)), nil
	default:
		return 0, fmt.Errorf("unknown duration unit %q in %q", string(unit), s)
	}
}

func ganttIsDuration(s string) bool {
	if len(s) < 2 {
		return false
	}
	unit := s[len(s)-1]
	if unit != 'd' && unit != 'w' && unit != 'h' {
		return false
	}
	_, err := strconv.Atoi(s[:len(s)-1])
	return err == nil
}

func ganttResolveEnd(start time.Time, endRaw, goFmt string) (time.Time, error) {
	if endRaw == "" {
		return start.AddDate(0, 0, 1), nil
	}
	if ganttIsDuration(endRaw) {
		days, err := ganttParseDuration(endRaw)
		if err != nil {
			return time.Time{}, err
		}
		return start.AddDate(0, 0, days), nil
	}
	t, err := time.Parse(goFmt, endRaw)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse end %q as date or duration", endRaw)
	}
	return t, nil
}

type ganttTask struct {
	name        string
	sectionName string
	sectionIdx  int
	start       time.Time
	end         time.Time
	isCrit      bool
	isDone      bool
	isActive    bool
	isMilestone bool
}

// GanttDrawing renders a GanttDiagram onto an Excel sheet.
type GanttDrawing struct {
	File    *excelize.File
	Sheet   string
	Diagram *ast.GanttDiagram
}

func (g *GanttDrawing) drawGantt() error {
	d := g.Diagram
	goFmt := mermaidDateToGo(d.DateFormat)

	// --- Date resolution ---
	type rawEntry struct {
		task        *ast.GanttTask
		sectionName string
		sectionIdx  int
	}
	var allRaw []rawEntry
	for si, sec := range d.Sections {
		for _, t := range sec.Tasks {
			allRaw = append(allRaw, rawEntry{t, sec.Name, si})
		}
	}

	resolved := make([]*ganttTask, len(allRaw))
	idMap := map[string]*ganttTask{}

	// First pass: tasks with explicit absolute start dates
	for i, re := range allRaw {
		t := re.task
		if t.StartRaw == "" || t.After != "" {
			continue
		}
		start, err := time.Parse(goFmt, t.StartRaw)
		if err != nil {
			return fmt.Errorf("task %q: parse start %q: %w", t.Name, t.StartRaw, err)
		}
		end, err := ganttResolveEnd(start, t.EndRaw, goFmt)
		if err != nil {
			return fmt.Errorf("task %q: resolve end: %w", t.Name, err)
		}
		rt := &ganttTask{
			name: t.Name, sectionName: re.sectionName, sectionIdx: re.sectionIdx,
			start: start, end: end,
			isCrit: t.IsCrit, isDone: t.IsDone, isActive: t.IsActive, isMilestone: t.IsMilestone,
		}
		resolved[i] = rt
		idMap[t.ID] = rt
	}

	// Find initial project start from absolute-date tasks
	var projectStart time.Time
	for _, rt := range resolved {
		if rt == nil {
			continue
		}
		if projectStart.IsZero() || rt.start.Before(projectStart) {
			projectStart = rt.start
		}
	}

	// Iterative pass: resolve "after" and no-start tasks, handling chains
	for range allRaw {
		progress := false
		for i, re := range allRaw {
			if resolved[i] != nil {
				continue
			}
			t := re.task
			var start time.Time
			if t.After != "" {
				pred, ok := idMap[t.After]
				if !ok {
					continue // predecessor not yet resolved
				}
				start = pred.end
			} else {
				if projectStart.IsZero() {
					continue // no project start known yet
				}
				start = projectStart
			}
			end, err := ganttResolveEnd(start, t.EndRaw, goFmt)
			if err != nil {
				end = start.AddDate(0, 0, 1)
			}
			rt := &ganttTask{
				name: t.Name, sectionName: re.sectionName, sectionIdx: re.sectionIdx,
				start: start, end: end,
				isCrit: t.IsCrit, isDone: t.IsDone, isActive: t.IsActive, isMilestone: t.IsMilestone,
			}
			resolved[i] = rt
			idMap[t.ID] = rt
			if projectStart.IsZero() || start.Before(projectStart) {
				projectStart = start
			}
			progress = true
		}
		if !progress {
			break
		}
	}

	// Fallback: any still-unresolved tasks get today as start
	today := time.Now().Truncate(24 * time.Hour)
	for i, re := range allRaw {
		if resolved[i] != nil {
			continue
		}
		start := today
		if !projectStart.IsZero() {
			start = projectStart
		}
		t := re.task
		end, err := ganttResolveEnd(start, t.EndRaw, goFmt)
		if err != nil {
			end = start.AddDate(0, 0, 1)
		}
		resolved[i] = &ganttTask{
			name: t.Name, sectionName: re.sectionName, sectionIdx: re.sectionIdx,
			start: start, end: end,
			isCrit: t.IsCrit, isDone: t.IsDone, isActive: t.IsActive, isMilestone: t.IsMilestone,
		}
	}

	// Find global date range
	var minDate, maxDate time.Time
	for _, rt := range resolved {
		if minDate.IsZero() || rt.start.Before(minDate) {
			minDate = rt.start
		}
		if maxDate.IsZero() || rt.end.After(maxDate) {
			maxDate = rt.end
		}
	}

	totalDays := int(maxDate.Sub(minDate).Hours()/24) + 1

	// Choose time granularity
	unitDays := 1
	slotColWidth := 3.5
	if totalDays > 60 {
		unitDays = 7
		slotColWidth = 5.5
	}
	numSlots := (totalDays + unitDays - 1) / unitDays

	slotLabel := func(t time.Time) string {
		if unitDays == 1 {
			return fmt.Sprintf("%d/%d", int(t.Month()), t.Day())
		}
		_, w := t.ISOWeek()
		return fmt.Sprintf("W%d\n%s", w, t.Format("Jan"))
	}

	// --- Layout constants ---
	const (
		sectionCol  = 1
		taskCol     = 2
		firstSlot   = 3
		titleRow    = 1
		headerRow   = 2
		firstTaskRow = 3
	)
	const (
		sectionColWidth = 13.0
		taskColWidth    = 22.0
		titlePt         = 24.0
		headerPt        = 32.0
		taskPt          = 18.0
	)

	f := g.File
	sheet := g.Sheet

	secColName, _ := excelize.ColumnNumberToName(sectionCol)
	taskColName, _ := excelize.ColumnNumberToName(taskCol)

	// Column widths
	if err := f.SetColWidth(sheet, secColName, secColName, sectionColWidth); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, taskColName, taskColName, taskColWidth); err != nil {
		return err
	}
	for i := range numSlots {
		col, _ := excelize.ColumnNumberToName(firstSlot + i)
		if err := f.SetColWidth(sheet, col, col, slotColWidth); err != nil {
			return err
		}
	}

	// Row heights
	if err := f.SetRowHeight(sheet, titleRow, titlePt); err != nil {
		return err
	}
	if err := f.SetRowHeight(sheet, headerRow, headerPt); err != nil {
		return err
	}
	for i := range resolved {
		if err := f.SetRowHeight(sheet, firstTaskRow+i, taskPt); err != nil {
			return err
		}
	}

	lastSlotCol, _ := excelize.ColumnNumberToName(firstSlot + numSlots - 1)

	// Title row
	if d.Title != "" {
		titleCell := fmt.Sprintf("%s%d", secColName, titleRow)
		lastTitle := fmt.Sprintf("%s%d", lastSlotCol, titleRow)
		if err := f.MergeCell(sheet, titleCell, lastTitle); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, titleCell, d.Title); err != nil {
			return err
		}
		style, err := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 14, Color: "333333", Family: "Calibri"},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{"F2F2F2"}, Pattern: 1},
		})
		if err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, titleCell, lastTitle, style); err != nil {
			return err
		}
	}

	// Header row labels
	headerLabelStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "FFFFFF", Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"404040"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "right", Color: "666666", Style: 1}},
	})
	if err != nil {
		return err
	}
	for _, cv := range []struct {
		col, val string
	}{
		{secColName, "Section"},
		{taskColName, "Task"},
	} {
		cell := fmt.Sprintf("%s%d", cv.col, headerRow)
		if err := f.SetCellValue(sheet, cell, cv.val); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cell, cell, headerLabelStyle); err != nil {
			return err
		}
	}

	// Date slot headers
	headerDateStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: false, Size: 8, Color: "333333", Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9D9D9"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "right", Color: "CCCCCC", Style: 1}},
	})
	if err != nil {
		return err
	}
	for i := range numSlots {
		slotDate := minDate.AddDate(0, 0, i*unitDays)
		col, _ := excelize.ColumnNumberToName(firstSlot + i)
		cell := fmt.Sprintf("%s%d", col, headerRow)
		if err := f.SetCellValue(sheet, cell, slotLabel(slotDate)); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cell, cell, headerDateStyle); err != nil {
			return err
		}
	}

	// Section palette
	sectionColors := [][2]string{
		{"7B7FC4", "FFFFFF"},
		{"F5A566", "333333"},
		{"7BC47B", "FFFFFF"},
		{"66B5F5", "333333"},
		{"C47B7B", "FFFFFF"},
		{"F5F566", "333333"},
	}

	// Task bar colors by status
	barColorFor := func(rt *ganttTask) string {
		switch {
		case rt.isMilestone:
			return "7030A0"
		case rt.isDone:
			return "A9A9A9"
		case rt.isActive:
			return "70AD47"
		case rt.isCrit:
			return "C0504D"
		default:
			return "4472C4"
		}
	}

	// Empty slot style
	emptyStyle, err := f.NewStyle(&excelize.Style{
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"F8F8F8"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "right", Color: "E8E8E8", Style: 1},
			{Type: "bottom", Color: "E8E8E8", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	// Task label style
	taskLabelStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "333333", Family: "Calibri"},
		Alignment: &excelize.Alignment{Vertical: "center", Indent: 1},
		Border:    []excelize.Border{{Type: "bottom", Color: "E0E0E0", Style: 1}},
	})
	if err != nil {
		return err
	}

	// Render task rows
	type secRange struct {
		name     string
		idx      int
		startRow int
		endRow   int
	}
	var secRanges []secRange
	secIdxMap := map[int]int{} // sectionIdx → index in secRanges

	for taskIdx, rt := range resolved {
		rowNum := firstTaskRow + taskIdx

		// Track section range
		si := rt.sectionIdx
		if pos, ok := secIdxMap[si]; ok {
			secRanges[pos].endRow = rowNum
		} else {
			secIdxMap[si] = len(secRanges)
			secRanges = append(secRanges, secRange{
				name: rt.sectionName, idx: si,
				startRow: rowNum, endRow: rowNum,
			})
		}

		// Task name cell
		taskCell := fmt.Sprintf("%s%d", taskColName, rowNum)
		if err := f.SetCellValue(sheet, taskCell, rt.name); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, taskCell, taskCell, taskLabelStyle); err != nil {
			return err
		}

		// Bar style for this task
		barStyle, err := f.NewStyle(&excelize.Style{
			Fill:   excelize.Fill{Type: "pattern", Color: []string{barColorFor(rt)}, Pattern: 1},
			Border: []excelize.Border{{Type: "bottom", Color: "E0E0E0", Style: 1}},
		})
		if err != nil {
			return err
		}

		startSlot := int(rt.start.Sub(minDate).Hours() / 24 / float64(unitDays))
		endSlot := int(math.Ceil(rt.end.Sub(minDate).Hours() / 24 / float64(unitDays)))
		if startSlot < 0 {
			startSlot = 0
		}
		if endSlot > numSlots {
			endSlot = numSlots
		}

		for s := range numSlots {
			col, _ := excelize.ColumnNumberToName(firstSlot + s)
			cell := fmt.Sprintf("%s%d", col, rowNum)
			if s >= startSlot && s < endSlot {
				if err := f.SetCellStyle(sheet, cell, cell, barStyle); err != nil {
					return err
				}
			} else {
				if err := f.SetCellStyle(sheet, cell, cell, emptyStyle); err != nil {
					return err
				}
			}
		}
	}

	// Render section column (merge + color)
	for _, sr := range secRanges {
		palette := sectionColors[sr.idx%len(sectionColors)]
		secStyle, err := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 10, Color: palette[1], Family: "Calibri"},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true, TextRotation: 90},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{palette[0]}, Pattern: 1},
		})
		if err != nil {
			return err
		}
		topCell := fmt.Sprintf("%s%d", secColName, sr.startRow)
		botCell := fmt.Sprintf("%s%d", secColName, sr.endRow)
		if sr.startRow != sr.endRow {
			if err := f.MergeCell(sheet, topCell, botCell); err != nil {
				return err
			}
		}
		if err := f.SetCellValue(sheet, topCell, sr.name); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, topCell, botCell, secStyle); err != nil {
			return err
		}
	}

	return nil
}

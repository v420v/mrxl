package ast

type QuadrantChart struct {
	Title     string
	XAxisLow  string
	XAxisHigh string
	YAxisLow  string
	YAxisHigh string
	Quadrant1 string // top-right  (high x, high y)
	Quadrant2 string // top-left   (low x,  high y)
	Quadrant3 string // bottom-left  (low x,  low y)
	Quadrant4 string // bottom-right (high x, low y)
	Points    []*QuadrantPoint
}

func NewQuadrantChart(title, xLow, xHigh, yLow, yHigh, q1, q2, q3, q4 string, points []*QuadrantPoint) Diagram {
	return &QuadrantChart{
		Title:     title,
		XAxisLow:  xLow,
		XAxisHigh: xHigh,
		YAxisLow:  yLow,
		YAxisHigh: yHigh,
		Quadrant1: q1,
		Quadrant2: q2,
		Quadrant3: q3,
		Quadrant4: q4,
		Points:    points,
	}
}

func (d *QuadrantChart) Type() string {
	return "quadrant"
}

type QuadrantPoint struct {
	Label string
	X     float64
	Y     float64
}

func NewQuadrantPoint(label string, x, y float64) *QuadrantPoint {
	return &QuadrantPoint{Label: label, X: x, Y: y}
}

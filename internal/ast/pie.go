package ast

type PieChart struct {
	Title  string
	Slices []*PieSlice
}

func NewPieChart(title string, slices []*PieSlice) Diagram {
	return &PieChart{Title: title, Slices: slices}
}

func (d *PieChart) Type() string {
	return "pie"
}

type PieSlice struct {
	Label string
	Value float64
}

func NewPieSlice(label string, value float64) *PieSlice {
	return &PieSlice{Label: label, Value: value}
}

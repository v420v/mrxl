package ast

type GanttDiagram struct {
	Title      string
	DateFormat string
	Sections   []*GanttSection
}

func NewGanttDiagram(title, dateFormat string, sections []*GanttSection) Diagram {
	return &GanttDiagram{Title: title, DateFormat: dateFormat, Sections: sections}
}

func (d *GanttDiagram) Type() string { return "gantt" }

type GanttSection struct {
	Name  string
	Tasks []*GanttTask
}

func NewGanttSection(name string) *GanttSection {
	return &GanttSection{Name: name, Tasks: make([]*GanttTask, 0)}
}

type GanttTask struct {
	Name        string
	ID          string
	After       string // predecessor task ID (empty if not used)
	StartRaw    string // raw date string (empty if using After or no explicit start)
	EndRaw      string // raw date or duration string like "30d", "2w"
	IsCrit      bool
	IsDone      bool
	IsActive    bool
	IsMilestone bool
}

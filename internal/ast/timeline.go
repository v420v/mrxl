package ast

type TimeDiagram struct {
	Title    string
	Sections []*TimeSection
}

func NewTimeDiagram(title string, sections []*TimeSection) Diagram {
	return &TimeDiagram{Title: title, Sections: sections}
}

func (d *TimeDiagram) Type() string {
	return "timeline"
}

type TimeSection struct {
	Name   string
	Events []*TimeEvent
}

func NewTimeSection(name string) *TimeSection {
	return &TimeSection{Name: name, Events: make([]*TimeEvent, 0)}
}

type TimeEvent struct {
	Time  string
	Texts []string
}

func NewTimeEvent(timeLabel string, text string) *TimeEvent {
	return &TimeEvent{Time: timeLabel, Texts: []string{text}}
}

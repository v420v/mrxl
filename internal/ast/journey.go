package ast

type UserJourneyDiagram struct {
	Title    string
	Sections []*JourneySection
}

func NewUserJourneyDiagram(title string, sections []*JourneySection) Diagram {
	return &UserJourneyDiagram{Title: title, Sections: sections}
}

func (d *UserJourneyDiagram) Type() string {
	return "journey"
}

type JourneySection struct {
	Name  string
	Tasks []*JourneyTask
}

func NewJourneySection(name string) *JourneySection {
	return &JourneySection{Name: name, Tasks: make([]*JourneyTask, 0)}
}

type JourneyTask struct {
	Title  string
	Score  float64
	Actors []string
}

func NewJourneyTask(title string, score float64, actors []string) *JourneyTask {
	return &JourneyTask{Title: title, Score: score, Actors: actors}
}

package ast

type SequenceDiagram struct {
	Title        string
	Autonumber   bool
	Participants []*Participant
	Events       []SequenceEvent
}

func NewSequenceDiagram(title string, autonumber bool, participants []*Participant, events []SequenceEvent) Diagram {
	return &SequenceDiagram{Title: title, Autonumber: autonumber, Participants: participants, Events: events}
}

func (d *SequenceDiagram) Type() string {
	return "sequence"
}

// SequenceEvent is either a *Message or a *Note.
type SequenceEvent interface {
	isSequenceEvent()
}

type Participant struct {
	Name string
}

func NewParticipant(name string) *Participant {
	return &Participant{Name: name}
}

// LineStyle is the style of the line
type LineStyle int

const (
	LineSolid LineStyle = iota
	LineDashed
)

// ArrowHead is the type of the arrow head
type ArrowHead int

const (
	ArrowFilled ArrowHead = iota // > or >>
	ArrowOpen                    // )
	ArrowCross                   // x
)

// MessageKind is the type of the message
type MessageKind int

const (
	KindCall MessageKind = iota
	KindReturn
	KindAsync
	KindError
)

type Message struct {
	From        *Participant
	To          *Participant
	LineStyle   LineStyle
	ArrowHead   ArrowHead
	MessageKind MessageKind
	Text        string
}

func NewMessage(from *Participant, to *Participant, lineStyle LineStyle, arrowHead ArrowHead, messageKind MessageKind, text string) *Message {
	return &Message{
		From:        from,
		To:          to,
		LineStyle:   lineStyle,
		ArrowHead:   arrowHead,
		MessageKind: messageKind,
		Text:        text,
	}
}

func (m *Message) isSequenceEvent() {}

// NotePosition indicates where the note is placed relative to its participant(s).
type NotePosition int

const (
	NoteLeft  NotePosition = iota // note left of X
	NoteRight                     // note right of X
	NoteOver                      // note over X[,Y]
)

type Note struct {
	Position NotePosition
	Left     *Participant
	Right    *Participant // same as Left for single-participant notes
	Text     string
}

func NewNote(pos NotePosition, left, right *Participant, text string) *Note {
	return &Note{Position: pos, Left: left, Right: right, Text: text}
}

func (n *Note) isSequenceEvent() {}

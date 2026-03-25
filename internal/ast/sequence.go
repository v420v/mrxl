package ast

type SequenceDiagram struct {
	Title        string
	Autonumber   bool
	Participants []*Participant
	Messages     []*Message
}

func NewSequenceDiagram(title string, autonumber bool, participants []*Participant, messages []*Message) Diagram {
	return &SequenceDiagram{Title: title, Autonumber: autonumber, Participants: participants, Messages: messages}
}

func (d *SequenceDiagram) Type() string {
	return "sequence"
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

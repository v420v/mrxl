package ast

type Diagram interface {
	// Type returns a string identifying the kind of diagram,
	// e.g. "sequence", "state", etc.
	Type() string
}

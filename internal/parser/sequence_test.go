package parser

import (
	"testing"

	"github.com/v420v/mrxl/internal/ast"
)

// mustMessage asserts that ev is a *ast.Message and returns it.
func mustMessage(t *testing.T, ev ast.SequenceEvent) *ast.Message {
	t.Helper()
	m, ok := ev.(*ast.Message)
	if !ok {
		t.Fatalf("expected *ast.Message, got %T", ev)
	}
	return m
}

// mustNote asserts that ev is a *ast.Note and returns it.
func mustNote(t *testing.T, ev ast.SequenceEvent) *ast.Note {
	t.Helper()
	n, ok := ev.(*ast.Note)
	if !ok {
		t.Fatalf("expected *ast.Note, got %T", ev)
	}
	return n
}

// mustActivation asserts that ev is a *ast.Activation and returns it.
func mustActivation(t *testing.T, ev ast.SequenceEvent) *ast.Activation {
	t.Helper()
	a, ok := ev.(*ast.Activation)
	if !ok {
		t.Fatalf("expected *ast.Activation, got %T", ev)
	}
	return a
}

// mustBlock asserts that ev is a *ast.InteractionBlock and returns it.
func mustBlock(t *testing.T, ev ast.SequenceEvent) *ast.InteractionBlock {
	t.Helper()
	b, ok := ev.(*ast.InteractionBlock)
	if !ok {
		t.Fatalf("expected *ast.InteractionBlock, got %T", ev)
	}
	return b
}

func TestParseSequenceDiagram(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, d *ast.SequenceDiagram)
	}{
		{
			name:  "basic participants and message",
			input: "participant A\nparticipant B\nA->B: hello",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Participants) != 2 {
					t.Fatalf("len(Participants) = %d, want 2", len(d.Participants))
				}
				if len(d.Events) != 1 {
					t.Fatalf("len(Events) = %d, want 1", len(d.Events))
				}
				msg := mustMessage(t, d.Events[0])
				if msg.Text != "hello" {
					t.Errorf("Text = %q, want %q", msg.Text, "hello")
				}
				if msg.LineStyle != ast.LineSolid {
					t.Errorf("LineStyle = %v, want LineSolid", msg.LineStyle)
				}
				if msg.ArrowHead != ast.ArrowOpen {
					t.Errorf("ArrowHead = %v, want ArrowOpen", msg.ArrowHead)
				}
			},
		},
		{
			name:  "arrow type ->",
			input: "participant A\nparticipant B\nA->B: msg",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				msg := mustMessage(t, d.Events[0])
				if msg.LineStyle != ast.LineSolid {
					t.Errorf("LineStyle = %v, want LineSolid", msg.LineStyle)
				}
				if msg.ArrowHead != ast.ArrowOpen {
					t.Errorf("ArrowHead = %v, want ArrowOpen", msg.ArrowHead)
				}
			},
		},
		{
			name:  "arrow type ->>",
			input: "participant A\nparticipant B\nA->>B: msg",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				msg := mustMessage(t, d.Events[0])
				if msg.LineStyle != ast.LineSolid {
					t.Errorf("LineStyle = %v, want LineSolid", msg.LineStyle)
				}
				if msg.ArrowHead != ast.ArrowFilled {
					t.Errorf("ArrowHead = %v, want ArrowFilled", msg.ArrowHead)
				}
			},
		},
		{
			name:  "arrow type -->",
			input: "participant A\nparticipant B\nA-->B: msg",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				msg := mustMessage(t, d.Events[0])
				if msg.LineStyle != ast.LineDashed {
					t.Errorf("LineStyle = %v, want LineDashed", msg.LineStyle)
				}
				if msg.ArrowHead != ast.ArrowOpen {
					t.Errorf("ArrowHead = %v, want ArrowOpen", msg.ArrowHead)
				}
			},
		},
		{
			name:  "arrow type -->>",
			input: "participant A\nparticipant B\nA-->>B: msg",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				msg := mustMessage(t, d.Events[0])
				if msg.LineStyle != ast.LineDashed {
					t.Errorf("LineStyle = %v, want LineDashed", msg.LineStyle)
				}
				if msg.ArrowHead != ast.ArrowFilled {
					t.Errorf("ArrowHead = %v, want ArrowFilled", msg.ArrowHead)
				}
			},
		},
		{
			name:  "actor keyword",
			input: "actor User\nUser->User: self",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Participants) == 0 {
					t.Fatal("expected at least 1 participant")
				}
				if d.Participants[0].Name != "User" {
					t.Errorf("Participants[0].Name = %q, want %q", d.Participants[0].Name, "User")
				}
			},
		},
		{
			name:  "title",
			input: "title My Seq\nparticipant A\nA->A: x",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if d.Title != "My Seq" {
					t.Errorf("Title = %q, want %q", d.Title, "My Seq")
				}
			},
		},
		{
			name:  "autonumber",
			input: "autonumber\nparticipant A\nA->A: x",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if !d.Autonumber {
					t.Error("Autonumber = false, want true")
				}
			},
		},
		{
			name:  "implicit participant from message",
			input: "A->B: msg",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Participants) != 2 {
					t.Fatalf("len(Participants) = %d, want 2", len(d.Participants))
				}
				names := map[string]bool{}
				for _, p := range d.Participants {
					names[p.Name] = true
				}
				if !names["A"] {
					t.Error("participant A not found")
				}
				if !names["B"] {
					t.Error("participant B not found")
				}
			},
		},
		{
			name:  "note left of",
			input: "participant A\nnote left of A: info\nA->A: x",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Events) == 0 {
					t.Fatal("no events")
				}
				note := mustNote(t, d.Events[0])
				if note.Position != ast.NoteLeft {
					t.Errorf("Position = %v, want NoteLeft", note.Position)
				}
				if note.Text != "info" {
					t.Errorf("Text = %q, want %q", note.Text, "info")
				}
			},
		},
		{
			name:  "note right of",
			input: "participant A\nnote right of A: info\nA->A: x",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Events) == 0 {
					t.Fatal("no events")
				}
				note := mustNote(t, d.Events[0])
				if note.Position != ast.NoteRight {
					t.Errorf("Position = %v, want NoteRight", note.Position)
				}
			},
		},
		{
			name:  "note over single",
			input: "participant A\nnote over A: info\nA->A: x",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				note := mustNote(t, d.Events[0])
				if note.Position != ast.NoteOver {
					t.Errorf("Position = %v, want NoteOver", note.Position)
				}
				if note.Left == nil || note.Right == nil {
					t.Fatal("Left or Right is nil")
				}
				if note.Left != note.Right {
					t.Error("expected Left == Right for single-participant note over")
				}
			},
		},
		{
			name:  "note over two participants",
			input: "participant A\nparticipant B\nnote over A,B: span\nA->B: x",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				note := mustNote(t, d.Events[0])
				if note.Position != ast.NoteOver {
					t.Errorf("Position = %v, want NoteOver", note.Position)
				}
				if note.Left == nil || note.Left.Name != "A" {
					t.Errorf("Left.Name = %q, want %q", note.Left.Name, "A")
				}
				if note.Right == nil || note.Right.Name != "B" {
					t.Errorf("Right.Name = %q, want %q", note.Right.Name, "B")
				}
			},
		},
		{
			name:  "activate deactivate",
			input: "participant A\nactivate A\nA->A: x\ndeactivate A",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Events) != 3 {
					t.Fatalf("len(Events) = %d, want 3", len(d.Events))
				}
				act0 := mustActivation(t, d.Events[0])
				if !act0.Active {
					t.Error("events[0] Active = false, want true")
				}
				if act0.Participant == nil || act0.Participant.Name != "A" {
					t.Errorf("events[0] Participant.Name = %q, want %q", act0.Participant.Name, "A")
				}
				mustMessage(t, d.Events[1])
				act2 := mustActivation(t, d.Events[2])
				if act2.Active {
					t.Error("events[2] Active = true, want false")
				}
			},
		},
		{
			name:  "activation shorthand +",
			input: "participant A\nparticipant B\nA->>+B: start",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Events) != 2 {
					t.Fatalf("len(Events) = %d, want 2", len(d.Events))
				}
				msg := mustMessage(t, d.Events[0])
				if msg.From == nil || msg.From.Name != "A" {
					t.Errorf("From.Name = %q, want %q", msg.From.Name, "A")
				}
				if msg.To == nil || msg.To.Name != "B" {
					t.Errorf("To.Name = %q, want %q", msg.To.Name, "B")
				}
				act := mustActivation(t, d.Events[1])
				if !act.Active {
					t.Error("Activation Active = false, want true")
				}
				if act.Participant == nil || act.Participant.Name != "B" {
					t.Errorf("Activation Participant.Name = %q, want %q", act.Participant.Name, "B")
				}
			},
		},
		{
			name:  "activation shorthand -",
			input: "participant A\nparticipant B\nB-->>-A: done",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Events) != 2 {
					t.Fatalf("len(Events) = %d, want 2", len(d.Events))
				}
				msg := mustMessage(t, d.Events[0])
				if msg.From == nil || msg.From.Name != "B" {
					t.Errorf("From.Name = %q, want %q", msg.From.Name, "B")
				}
				if msg.To == nil || msg.To.Name != "A" {
					t.Errorf("To.Name = %q, want %q", msg.To.Name, "A")
				}
				act := mustActivation(t, d.Events[1])
				if act.Active {
					t.Error("Activation Active = true, want false")
				}
				if act.Participant == nil || act.Participant.Name != "A" {
					t.Errorf("Activation Participant.Name = %q, want %q", act.Participant.Name, "A")
				}
			},
		},
		{
			name:  "loop block",
			input: "participant A\nparticipant B\nloop every second\nA->>B: ping\nend",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Events) != 1 {
					t.Fatalf("len(Events) = %d, want 1", len(d.Events))
				}
				block := mustBlock(t, d.Events[0])
				if block.Kind != "loop" {
					t.Errorf("Kind = %q, want %q", block.Kind, "loop")
				}
				if len(block.Branches) != 1 {
					t.Fatalf("len(Branches) = %d, want 1", len(block.Branches))
				}
				if block.Branches[0].Label != "every second" {
					t.Errorf("Branches[0].Label = %q, want %q", block.Branches[0].Label, "every second")
				}
				if len(block.Branches[0].Events) != 1 {
					t.Fatalf("len(Branches[0].Events) = %d, want 1", len(block.Branches[0].Events))
				}
				mustMessage(t, block.Branches[0].Events[0])
			},
		},
		{
			name:  "alt else block",
			input: "participant A\nparticipant B\nalt success\nA->>B: ok\nelse fail\nA->>B: error\nend",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				block := mustBlock(t, d.Events[0])
				if block.Kind != "alt" {
					t.Errorf("Kind = %q, want %q", block.Kind, "alt")
				}
				if len(block.Branches) != 2 {
					t.Fatalf("len(Branches) = %d, want 2", len(block.Branches))
				}
				if block.Branches[0].Label != "success" {
					t.Errorf("Branches[0].Label = %q, want %q", block.Branches[0].Label, "success")
				}
				if block.Branches[1].Label != "fail" {
					t.Errorf("Branches[1].Label = %q, want %q", block.Branches[1].Label, "fail")
				}
			},
		},
		{
			name:  "opt block",
			input: "participant A\nopt maybe\nA->A: x\nend",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				block := mustBlock(t, d.Events[0])
				if block.Kind != "opt" {
					t.Errorf("Kind = %q, want %q", block.Kind, "opt")
				}
			},
		},
		{
			name:  "break block",
			input: "participant A\nbreak overload\nA->A: x\nend",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				block := mustBlock(t, d.Events[0])
				if block.Kind != "break" {
					t.Errorf("Kind = %q, want %q", block.Kind, "break")
				}
			},
		},
		{
			name:  "nested block",
			input: "participant A\nparticipant B\nopt outer\nloop inner\nA->>B: msg\nend\nend",
			check: func(t *testing.T, d *ast.SequenceDiagram) {
				t.Helper()
				if len(d.Events) != 1 {
					t.Fatalf("len(Events) = %d, want 1", len(d.Events))
				}
				outer := mustBlock(t, d.Events[0])
				if outer.Kind != "opt" {
					t.Errorf("outer.Kind = %q, want %q", outer.Kind, "opt")
				}
				if len(outer.Branches) == 0 || len(outer.Branches[0].Events) == 0 {
					t.Fatal("outer block has no inner events")
				}
				inner := mustBlock(t, outer.Branches[0].Events[0])
				if inner.Kind != "loop" {
					t.Errorf("inner.Kind = %q, want %q", inner.Kind, "loop")
				}
			},
		},
		{
			name:    "unclosed block",
			input:   "participant A\nloop\nA->A: x",
			wantErr: true,
		},
		{
			name:    "else outside block",
			input:   "participant A\nelse\nA->A: x",
			wantErr: true,
		},
		{
			name:    "end outside block",
			input:   "participant A\nend",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSequenceDiagram(normalizedLines(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseSequenceDiagram() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			d, ok := got.(*ast.SequenceDiagram)
			if !ok {
				t.Fatalf("expected *ast.SequenceDiagram, got %T", got)
			}
			if tt.check != nil {
				tt.check(t, d)
			}
		})
	}
}

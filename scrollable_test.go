package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// scrollableMockPanel is a test double implementing Panel.
type scrollableMockPanel struct {
	view string
	msgs []tea.Msg
}

func (m *scrollableMockPanel) View(width, height int) string {
	return m.view
}

func (m *scrollableMockPanel) Update(msg tea.Msg) tea.Cmd {
	m.msgs = append(m.msgs, msg)
	return nil
}

func TestNewScrollable(t *testing.T) {
	p := &scrollableMockPanel{view: "hello"}
	s := NewScrollable(p)
	if s == nil {
		t.Fatal("NewScrollable returned nil")
	}
	if s.Content != p {
		t.Errorf("expected content to be set, got %v", s.Content)
	}
	if s.Offset != 0 {
		t.Errorf("expected initial offset 0, got %d", s.Offset)
	}
}

func TestScrollableView_NilContent(t *testing.T) {
	s := &Scrollable{Content: nil}
	got := s.View(10, 3)
	want := strings.Repeat("\n", 3)
	if got != want {
		t.Errorf("nil content View mismatch:\nwant %q\ngot  %q", want, got)
	}
}

func TestScrollableView_ShortContent(t *testing.T) {
	p := &scrollableMockPanel{view: "line1\nline2"}
	s := NewScrollable(p)
	got := s.View(6, 5)
	want := "line1 \nline2 \n      \n      \n      "
	if got != want {
		t.Errorf("short content View mismatch:\nwant %q\ngot  %q", want, got)
	}
}

func TestScrollableView_TallContentWithOffset(t *testing.T) {
	p := &scrollableMockPanel{view: "0\n1\n2\n3\n4\n5\n6\n7\n8\n9"}
	s := NewScrollable(p)
	s.Offset = 3
	got := s.View(1, 4)
	want := "3\n4\n5\n6"
	if got != want {
		t.Errorf("tall content View mismatch:\nwant %q\ngot  %q", want, got)
	}
}

func TestScrollableView_OffsetClamping(t *testing.T) {
	p := &scrollableMockPanel{view: "0\n1\n2\n3"}
	s := NewScrollable(p)

	s.Offset = -5
	got := s.View(1, 3)
	want := "0\n1\n2"
	if got != want {
		t.Errorf("negative offset clamping mismatch:\nwant %q\ngot  %q", want, got)
	}

	s.Offset = 100
	got = s.View(1, 3)
	want = "1\n2\n3"
	if got != want {
		t.Errorf("excess offset clamping mismatch:\nwant %q\ngot  %q", want, got)
	}
}

func TestScrollableUpdate_MouseWheel(t *testing.T) {
	p := &scrollableMockPanel{view: "0\n1\n2\n3\n4\n5\n6\n7\n8\n9"}
	s := NewScrollable(p)

	s.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	if s.Offset != 3 {
		t.Errorf("wheel down offset = %d, want 3", s.Offset)
	}

	s.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	if s.Offset != 0 {
		t.Errorf("wheel up offset = %d, want 0", s.Offset)
	}

	// Wheel up should not go below zero.
	s.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	if s.Offset != 0 {
		t.Errorf("wheel up below zero offset = %d, want 0", s.Offset)
	}

	if len(p.msgs) != 3 {
		t.Errorf("expected 3 forwarded messages, got %d", len(p.msgs))
	}
}

func TestScrollableUpdate_Keys(t *testing.T) {
	p := &scrollableMockPanel{view: "0\n1\n2\n3\n4\n5\n6\n7\n8\n9"}
	s := NewScrollable(p)

	s.Update(tea.KeyMsg{Type: tea.KeyDown})
	if s.Offset != 1 {
		t.Errorf("down offset = %d, want 1", s.Offset)
	}

	s.Update(tea.KeyMsg{Type: tea.KeyUp})
	if s.Offset != 0 {
		t.Errorf("up offset = %d, want 0", s.Offset)
	}

	s.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if s.Offset != 10 {
		t.Errorf("pgdown offset = %d, want 10", s.Offset)
	}

	s.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if s.Offset != 0 {
		t.Errorf("pgup offset = %d, want 0", s.Offset)
	}

	// Up below zero should clamp.
	s.Update(tea.KeyMsg{Type: tea.KeyUp})
	if s.Offset != 0 {
		t.Errorf("up below zero offset = %d, want 0", s.Offset)
	}

	if len(p.msgs) != 5 {
		t.Errorf("expected 5 forwarded messages, got %d", len(p.msgs))
	}
}

func TestScrollableUpdate_ForwardsToNilContent(t *testing.T) {
	s := &Scrollable{Content: nil}
	cmd := s.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil {
		t.Errorf("expected nil cmd for nil content, got %v", cmd)
	}
}

func TestPadLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		w    int
		want string
	}{
		{"pad", "hi", 5, "hi   "},
		{"exact", "hello", 5, "hello"},
		{"truncate", "hello world", 5, "hello"},
		{"empty", "", 3, "   "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := padLine(tt.line, tt.w)
			if got != tt.want {
				t.Errorf("padLine(%q, %d) = %q, want %q", tt.line, tt.w, got, tt.want)
			}
		})
	}
}

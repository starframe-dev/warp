package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// staticPanel is a Panel implementation that returns a fixed block of content.
type staticPanel struct {
	content string
	lastMsg tea.Msg
}

func (p *staticPanel) View(width, height int) string {
	lines := strings.Split(p.content, "\n")
	for i := range lines {
		if len(lines[i]) > width {
			lines[i] = lines[i][:width]
		} else {
			lines[i] = lines[i] + strings.Repeat(" ", width-len(lines[i]))
		}
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return strings.Join(lines[:height], "\n")
}

func (p *staticPanel) Update(msg tea.Msg) tea.Cmd {
	p.lastMsg = msg
	return nil
}

func TestNewSelectable(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)
	if s == nil {
		t.Fatal("NewSelectable returned nil")
	}
	if s.Content != p {
		t.Error("NewSelectable did not store content panel")
	}
	if s.HasSelection {
		t.Error("new selectable should have no selection")
	}
}

func TestSelectableViewNoSelection(t *testing.T) {
	p := &staticPanel{content: "hello\nworld"}
	s := NewSelectable(p)

	got := s.View(5, 2)
	want := "hello\nworld"
	if got != want {
		t.Errorf("View without selection: got %q, want %q", got, want)
	}
	if s.lastW != 5 || s.lastH != 2 {
		t.Errorf("View did not remember dimensions: got %dx%d", s.lastW, s.lastH)
	}
}

func TestSelectableViewNilContent(t *testing.T) {
	s := NewSelectable(nil)
	got := s.View(5, 2)
	want := "\n\n"
	if got != want {
		t.Errorf("View nil content: got %q, want %q", got, want)
	}
}

func TestSelectableSelectAll(t *testing.T) {
	p := &staticPanel{content: "abcd\nefgh"}
	s := NewSelectable(p)
	s.SelectAll(4, 2)

	if !s.HasSelection {
		t.Fatal("SelectAll should create selection")
	}
	if s.AnchorX != 0 || s.AnchorY != 0 {
		t.Errorf("SelectAll anchor: got (%d,%d), want (0,0)", s.AnchorX, s.AnchorY)
	}
	if s.CursorX != 3 || s.CursorY != 1 {
		t.Errorf("SelectAll cursor: got (%d,%d), want (3,1)", s.CursorX, s.CursorY)
	}
}

func TestSelectableViewWithSelection(t *testing.T) {
	p := &staticPanel{content: "hello\nworld"}
	s := NewSelectable(p)
	s.AnchorX, s.AnchorY = 0, 0
	s.CursorX, s.CursorY = 4, 0
	s.HasSelection = true

	got := s.View(5, 2)
	if !strings.Contains(got, selectionStyleANSI) {
		t.Errorf("View with selection did not apply highlight: %q", got)
	}
	if !strings.Contains(got, resetStyle) {
		t.Errorf("View with selection did not reset style: %q", got)
	}
}

func TestSelectableSelectedText(t *testing.T) {
	p := &staticPanel{content: "hello\nworld"}
	s := NewSelectable(p)
	// Selection is half-open in cell coordinates: cursor marks the boundary.
	s.AnchorX, s.AnchorY = 0, 0
	s.CursorX, s.CursorY = 4, 0
	s.HasSelection = true

	// Trigger View so lastLines is populated.
	s.View(5, 2)

	got := s.SelectedText()
	want := "hell"
	if got != want {
		t.Errorf("SelectedText: got %q, want %q", got, want)
	}
}

func TestSelectableSelectedTextNilContent(t *testing.T) {
	s := NewSelectable(nil)
	s.HasSelection = true
	if s.SelectedText() != "" {
		t.Error("SelectedText with nil content should be empty")
	}
}

func TestSelectableSelectedTextNoSelection(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)
	if s.SelectedText() != "" {
		t.Error("SelectedText without selection should be empty")
	}
}

func TestSelectableSelectedTextReverse(t *testing.T) {
	p := &staticPanel{content: "hello\nworld"}
	s := NewSelectable(p)
	// Cursor before anchor; selection extracts the half-open range.
	s.AnchorX, s.AnchorY = 3, 0
	s.CursorX, s.CursorY = 1, 0
	s.HasSelection = true

	s.View(5, 2)
	got := s.SelectedText()
	want := "el"
	if got != want {
		t.Errorf("SelectedText reverse: got %q, want %q", got, want)
	}
}

func TestSelectableSelectedTextWithoutView(t *testing.T) {
	p := &staticPanel{content: "hello\nworld"}
	s := NewSelectable(p)
	s.AnchorX, s.AnchorY = 0, 0
	s.CursorX, s.CursorY = 4, 0
	s.HasSelection = true

	got := s.SelectedText()
	want := "hell"
	if got != want {
		t.Errorf("SelectedText without View: got %q, want %q", got, want)
	}
}

func TestSelectableClearSelection(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)
	s.SelectAll(5, 1)
	s.ClearSelection()
	if s.HasSelection || s.Selecting {
		t.Error("ClearSelection should reset selection flags")
	}
}

func TestSelectableCopy(t *testing.T) {
	p := &staticPanel{content: "hello\nworld"}
	s := NewSelectable(p)
	s.SelectAll(5, 2)
	s.View(5, 2)

	cmd := s.Copy()
	if cmd == nil {
		t.Fatal("Copy should return a command when text is selected")
	}
	msg := cmd()
	if msg != nil {
		t.Errorf("Copy command returned non-nil msg: %v", msg)
	}
}

func TestSelectableCopyEmpty(t *testing.T) {
	s := NewSelectable(&staticPanel{content: "hello"})
	if s.Copy() != nil {
		t.Error("Copy with no selection should return nil")
	}
}

func TestSelectableUpdateMousePressAndRelease(t *testing.T) {
	p := &staticPanel{content: "hello\nworld"}
	s := NewSelectable(p)

	press := tea.MouseMsg{X: 1, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	cmd := s.Update(press)
	if cmd != nil {
		t.Error("Mouse press should not return command")
	}
	if s.Selecting != true {
		t.Error("Mouse press should set Selecting")
	}
	if s.HasSelection {
		t.Error("Mouse press should not create selection yet")
	}

	// Add a motion so the drag creates a selection; press-then-release alone does not.
	motion := tea.MouseMsg{X: 3, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion}
	s.Update(motion)

	release := tea.MouseMsg{X: 3, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease}
	s.Update(release)
	if s.Selecting {
		t.Error("Mouse release should clear Selecting")
	}
	if !s.HasSelection {
		t.Error("Mouse release should create selection")
	}
	if s.SelectedText() != "el" {
		t.Errorf("Mouse selection text: got %q, want %q", s.SelectedText(), "el")
	}
}

func TestSelectableUpdateMouseReleaseWithoutMotion(t *testing.T) {
	p := &staticPanel{content: "hello\nworld"}
	s := NewSelectable(p)

	press := tea.MouseMsg{X: 1, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	s.Update(press)
	release := tea.MouseMsg{X: 3, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease}
	s.Update(release)
	if s.HasSelection {
		t.Error("Press and release without motion should not select")
	}
}

func TestSelectableUpdateMouseReleaseSameCell(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)

	press := tea.MouseMsg{X: 1, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	s.Update(press)
	release := tea.MouseMsg{X: 1, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease}
	s.Update(release)
	if s.HasSelection {
		t.Error("Press and release at same cell should not select")
	}
}

func TestSelectableUpdateMouseMotion(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)

	press := tea.MouseMsg{X: 0, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	s.Update(press)
	motion := tea.MouseMsg{X: 4, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion}
	s.Update(motion)
	if !s.HasSelection {
		t.Error("Mouse motion during drag should create selection")
	}
	if s.SelectedText() != "hell" {
		t.Errorf("Mouse motion selection: got %q, want %q", s.SelectedText(), "hell")
	}
}

func TestSelectableUpdateMouseMotionWithoutSelecting(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)
	motion := tea.MouseMsg{X: 4, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion}
	s.Update(motion)
	if s.HasSelection {
		t.Error("Mouse motion without prior press should not select")
	}
}

func TestSelectableUpdateShiftArrows(t *testing.T) {
	p := &staticPanel{content: "hello\nworld\nfoo"}
	s := NewSelectable(p)

	// Shift+arrow keys move the cursor one cell per press. The selection is
	// half-open from the anchor to the cursor.
	s.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	if !s.HasSelection {
		t.Fatal("Shift+right should create selection")
	}
	if s.CursorX != 1 {
		t.Errorf("Shift+right cursor: got %d, want 1", s.CursorX)
	}
	s.View(5, 3)
	if s.SelectedText() != "h" {
		t.Errorf("Shift+right selection: got %q", s.SelectedText())
	}

	s.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	if s.CursorX != 2 {
		t.Errorf("Second shift+right cursor: got %d, want 2", s.CursorX)
	}
	s.View(5, 3)
	if s.SelectedText() != "he" {
		t.Errorf("Second shift+right selection: got %q", s.SelectedText())
	}

	s.Update(tea.KeyMsg{Type: tea.KeyShiftLeft})
	if s.CursorX != 1 {
		t.Errorf("Shift+left cursor: got %d, want 1", s.CursorX)
	}
	s.View(5, 3)
	if s.SelectedText() != "h" {
		t.Errorf("Shift+left selection: got %q", s.SelectedText())
	}

	s.Update(tea.KeyMsg{Type: tea.KeyShiftUp})
	if s.CursorY != 0 {
		t.Errorf("Shift+up should clamp cursor Y to 0, got %d", s.CursorY)
	}

	s.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	if s.CursorY != 1 {
		t.Errorf("Shift+down cursor Y: got %d, want 1", s.CursorY)
	}
}

func TestSelectableUpdateCtrlA(t *testing.T) {
	p := &staticPanel{content: "hello\nworld"}
	s := NewSelectable(p)
	s.lastW, s.lastH = 5, 2

	cmd := s.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if cmd != nil {
		t.Error("Ctrl+A should not return a command")
	}
	if !s.HasSelection {
		t.Fatal("Ctrl+A should select all")
	}
	s.View(5, 2)
	// SelectAll sets the cursor to the last cell, so the first selected line
	// is fully included and only the last line is truncated by the half-open range.
	if s.SelectedText() != "hello\nworl" {
		t.Errorf("Ctrl+A selection: got %q", s.SelectedText())
	}
}

func TestSelectableUpdateEsc(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)
	s.SelectAll(5, 1)

	cmd := s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Error("Esc with selection should not return command")
	}
	if s.HasSelection {
		t.Error("Esc should clear selection")
	}
}

func TestSelectableUpdateEscNoSelection(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)

	cmd := s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Error("Esc without selection should not return command")
	}
	if p.lastMsg == nil {
		t.Error("Esc without selection should be forwarded to content")
	}
}

func TestSelectableUpdateShiftTabForwarded(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)

	cmd := s.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if cmd != nil {
		t.Error("Shift+Tab should not return command from Selectable")
	}
	if p.lastMsg == nil {
		t.Fatal("Shift+Tab should be forwarded to content")
	}
	if key, ok := p.lastMsg.(tea.KeyMsg); !ok || key.String() != "shift+tab" {
		t.Errorf("Shift+Tab forwarded wrong message: %v", p.lastMsg)
	}
}

func TestSelectableUpdateUnmodifiedKeyForwarded(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	s.Update(msg)
	if p.lastMsg == nil {
		t.Fatal("Unmodified key should be forwarded to content")
	}
}

func TestSelectableUpdateResize(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)

	cmd := s.Update(ResizeMsg{Width: 10, Height: 5})
	if cmd != nil {
		t.Error("ResizeMsg should not return command from Selectable")
	}
	if p.lastMsg == nil {
		t.Fatal("ResizeMsg should be forwarded to content")
	}
}

func TestSelectableUpdateNilContent(t *testing.T) {
	s := NewSelectable(nil)

	press := tea.MouseMsg{X: 0, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	if cmd := s.Update(press); cmd != nil {
		t.Error("Update with nil content should return nil")
	}

	cmd := s.Update(ResizeMsg{Width: 10, Height: 5})
	if cmd != nil {
		t.Error("ResizeMsg with nil content should return nil")
	}
}

func TestSelectableSelectionClampBounds(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)
	s.SelectAll(5, 1)
	// Move cursor outside the view bounds.
	s.CursorX = 10
	s.CursorY = 5
	s.View(5, 1)
	if s.CursorX != 4 {
		t.Errorf("CursorX should be clamped to width-1, got %d", s.CursorX)
	}
	if s.CursorY != 0 {
		t.Errorf("CursorY should be clamped to height-1, got %d", s.CursorY)
	}
}

func TestSelectableSelectionClearedWhenCollapsed(t *testing.T) {
	p := &staticPanel{content: "hello"}
	s := NewSelectable(p)
	s.AnchorX, s.AnchorY = 2, 0
	s.CursorX, s.CursorY = 2, 0
	s.HasSelection = true
	s.View(5, 1)
	if s.HasSelection {
		t.Error("Selection should be cleared when anchor and cursor coincide")
	}
}

func TestSelectableHighlightRange(t *testing.T) {
	line := "hello world"
	got := highlightRange(line, 0, 5)
	if !strings.Contains(got, selectionStyleANSI) {
		t.Errorf("highlightRange should wrap selection: %q", got)
	}
	if !strings.Contains(got, resetStyle) {
		t.Errorf("highlightRange should reset style: %q", got)
	}
}

func TestSelectableHighlightRangeNoSelection(t *testing.T) {
	line := "hello world"
	got := highlightRange(line, 3, 3)
	if got != line {
		t.Errorf("highlightRange with empty range should return original line: got %q", got)
	}
}

func TestSelectableExtractVisRange(t *testing.T) {
	line := "hello world"
	got := extractVisRange(line, 6, 11)
	if got != "world" {
		t.Errorf("extractVisRange: got %q, want %q", got, "world")
	}
}

func TestSelectableExtractVisRangeANSI(t *testing.T) {
	line := "\x1b[31mhello\x1b[0m world"
	got := extractVisRange(line, 6, 11)
	if got != "world" {
		t.Errorf("extractVisRange ANSI: got %q, want %q", got, "world")
	}
}
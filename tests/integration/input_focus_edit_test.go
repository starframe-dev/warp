package integration_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

func TestInputFocusRoutingGraphemeEditingResizeAndSemanticBounds(t *testing.T) {
	group := warp.NewTabGroup(warp.TabNone)
	tab := group.ActiveTab()
	first := warp.NewInput("First")
	second := warp.NewInput("Second")
	tab.SetRootPanel(first)
	tab.SplitVertical(first, 0.5, second)

	tab.FocusFirst()
	if tab.Focus() != first || !first.Focused() || second.Focused() {
		t.Fatal("FocusFirst did not focus only the first input")
	}
	tab.FocusNext()
	if tab.Focus() != second || first.Focused() || !second.Focused() {
		t.Fatal("FocusNext did not route focus to the second input")
	}
	tab.FocusPrev()
	if tab.Focus() != first || !first.Focused() || second.Focused() {
		t.Fatal("FocusPrev did not route focus back to the first input")
	}

	// Input cursor positions are expressed in runes, but movement and deletion
	// must treat the woman-technologist sequence as one grapheme cluster.
	first.SetValue("a👩\u200d💻b")
	first.SetCursor(2) // inside the emoji cluster; normalize to its right edge.
	if first.Cursor != 4 {
		t.Fatalf("cursor inside grapheme normalized to %d, want 4", first.Cursor)
	}
	first.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if first.Cursor != 1 {
		t.Fatalf("left moved cursor to rune %d, want grapheme boundary 1", first.Cursor)
	}
	first.Update(tea.KeyMsg{Type: tea.KeyRight})
	if first.Cursor != 4 {
		t.Fatalf("right moved cursor to rune %d, want grapheme boundary 4", first.Cursor)
	}
	first.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if first.Value != "ab" || first.Cursor != 1 {
		t.Fatalf("backspace split grapheme: value=%q cursor=%d", first.Value, first.Cursor)
	}
	first.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("界")})
	if first.Value != "a界b" || first.Cursor != 2 {
		t.Fatalf("insert produced value=%q cursor=%d, want %q at 2", first.Value, first.Cursor, "a界b")
	}
	first.SetCursor(1)
	first.Update(tea.KeyMsg{Type: tea.KeyDelete})
	if first.Value != "ab" || first.Cursor != 1 {
		t.Fatalf("delete produced value=%q cursor=%d, want ab at 1", first.Value, first.Cursor)
	}

	// Focus gates keyboard editing.
	tab.FocusNext()
	before := first.Value
	first.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if first.Value != before || !second.Focused() {
		t.Fatal("unfocused input accepted keyboard input")
	}

	// Layout geometry, semantic bounds, and rendering all follow the current
	// viewport size; each input's textbox occupies its allocated leaf bounds.
	const width, height = 41, 12
	_ = group.View(width, height)
	elements := group.Elements(width, height)
	firstBox, ok := warp.FindElement(elements, "textbox", "First", "focus")
	if !ok {
		t.Fatalf("first textbox missing from semantic elements: %+v", elements)
	}
	secondBox, ok := warp.FindElement(elements, "textbox", "Second", "focus")
	if !ok {
		t.Fatalf("second textbox missing from semantic elements: %+v", elements)
	}
	if want := (warp.Bounds{W: 20, H: height}); firstBox.Bounds != want {
		t.Fatalf("first textbox bounds=%+v, want %+v", firstBox.Bounds, want)
	}
	if want := (warp.Bounds{X: 21, W: 20, H: height}); secondBox.Bounds != want {
		t.Fatalf("second textbox bounds=%+v, want %+v", secondBox.Bounds, want)
	}

	const resizedWidth, resizedHeight = 30, 8
	_ = group.View(resizedWidth, resizedHeight)
	resized := group.Elements(resizedWidth, resizedHeight)
	firstBox, _ = warp.FindElement(resized, "textbox", "First", "focus")
	secondBox, _ = warp.FindElement(resized, "textbox", "Second", "focus")
	if want := (warp.Bounds{W: 14, H: resizedHeight}); firstBox.Bounds != want {
		t.Fatalf("resized first textbox bounds=%+v, want %+v", firstBox.Bounds, want)
	}
	if want := (warp.Bounds{X: 15, W: 15, H: resizedHeight}); secondBox.Bounds != want {
		t.Fatalf("resized second textbox bounds=%+v, want %+v", secondBox.Bounds, want)
	}
}

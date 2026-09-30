package e2e_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

func TestFocusTraversalAndInputEditingRoutesKeysOnlyToFocusedPanel(t *testing.T) {
	first := &tabE2EPanel{}
	input := warp.NewInput("Text")
	last := &tabE2EPanel{}

	tab := warp.NewTab("focus-input")
	tab.SetRootPanel(first)
	tab.FlexRow(first, []warp.FlexItemSpec{
		{Panel: first, Grow: 1},
		{Panel: input, Grow: 2},
		{Panel: last, Grow: 1},
	})

	tab.FocusFirst()
	if tab.Focus() != first || !first.Focused() || input.Focused() || last.Focused() {
		t.Fatal("FocusFirst did not focus the first panel")
	}

	// Tab traversal follows visual order, wraps at either end, and blurs the
	// panel that no longer owns focus.
	tab.FocusNext()
	if tab.Focus() != input || !input.Focused() || first.Focused() || last.Focused() {
		t.Fatal("FocusNext did not transfer focus to the input")
	}
	tab.FocusNext()
	if tab.Focus() != last || !last.Focused() || input.Focused() {
		t.Fatal("FocusNext did not transfer focus to the last panel")
	}
	tab.FocusNext()
	if tab.Focus() != first || !first.Focused() || last.Focused() {
		t.Fatal("FocusNext did not wrap to the first panel")
	}
	tab.FocusPrev()
	if tab.Focus() != last || !last.Focused() || first.Focused() {
		t.Fatal("FocusPrev did not wrap to the last panel")
	}
	tab.FocusPrev()
	if tab.Focus() != input || !input.Focused() || last.Focused() {
		t.Fatal("FocusPrev did not transfer focus to the input")
	}

	// Keys dispatched through the tab reach only the focused input; the input
	// itself ignores editing messages while blurred.
	input.Blur()
	input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ignored")})
	if input.Value != "" {
		t.Fatalf("blurred input accepted text: %q", input.Value)
	}
	tab.FocusPanel(first)
	firstFocuses, lastFocuses := first.focuses, last.focuses
	first.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ignored")})
	if input.Value != "" {
		t.Fatalf("typing sent to an unfocused panel changed the input: %q", input.Value)
	}
	tab.FocusPanel(input)
	input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	if input.Value != "hello" || first.focuses != firstFocuses || last.focuses != lastFocuses {
		t.Fatalf("typing was not confined to the focused input: value=%q", input.Value)
	}

	// Move the rune cursor left twice, insert in the middle, then exercise
	// backspace and right-cursor movement.
	input.Update(tea.KeyMsg{Type: tea.KeyLeft})
	input.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if input.Cursor != 3 {
		t.Fatalf("cursor after two left keys = %d; want 3", input.Cursor)
	}
	input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("X")})
	if input.Value != "helXlo" || input.Cursor != 4 {
		t.Fatalf("middle insertion produced value=%q cursor=%d; want helXlo at 4", input.Value, input.Cursor)
	}
	input.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if input.Value != "hello" || input.Cursor != 3 {
		t.Fatalf("backspace produced value=%q cursor=%d; want hello at 3", input.Value, input.Cursor)
	}
	input.Update(tea.KeyMsg{Type: tea.KeyRight})
	input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")})
	if input.Value != "hell!o" || input.Cursor != 5 {
		t.Fatalf("right movement and insertion produced value=%q cursor=%d; want hell!o at 5", input.Value, input.Cursor)
	}
}

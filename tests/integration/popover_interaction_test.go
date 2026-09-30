package integration_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	warp "github.com/starframe-dev/warp"
)

func popoverMouse(x, y int, action tea.MouseAction) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: action}
}

func popoverKey(key tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: key} }

func popoverBackground(width, height int) []string {
	lines := make([]string, height)
	for i := range lines {
		lines[i] = strings.Repeat("x", width)
	}
	return lines
}

func TestPopoverOverlaysAndClampsToViewport(t *testing.T) {
	const width, height = 60, 10
	background := popoverBackground(width, height)
	background[0] = "terminal header content"
	background[9] = "terminal footer content"
	menu := &warp.Popover{
		Items: []warp.PopoverItem{{Name: "Open"}, {Name: "Save"}, {Name: "Quit"}},
		X: 58, Y: 9, Width: 12,
	}

	lines := menu.Overlay(background, width, height)
	if len(lines) != height {
		t.Fatalf("overlay line count=%d, want %d", len(lines), height)
	}
	if !strings.Contains(warp.StripANSI(lines[0]), "terminal header content") {
		t.Fatalf("uncovered header content was lost: %q", warp.StripANSI(lines[0]))
	}
	if !strings.Contains(warp.StripANSI(lines[9]), "terminal footer content") {
		t.Fatalf("content outside menu width was not preserved: %q", warp.StripANSI(lines[9]))
	}
	// The framed menu is 14 columns wide and five rows high. Its requested
	// bottom-right position is therefore clamped to (46, 5).
	const clampedX, clampedY = 46, 5
	for row := clampedY; row < height; row++ {
		plain := warp.StripANSI(lines[row])
		if got := strings.Index(plain, "┌"); row == clampedY && got != clampedX {
			t.Fatalf("top border begins at x=%d, want clamped x=%d: %q", got, clampedX, plain)
		}
		if got := ansi.StringWidth(lines[row]); got < clampedX+14 {
			t.Fatalf("overlaid row %d width=%d, menu should reach at least %d columns", row, got, clampedX+14)
		}
	}
	plainMenu := warp.StripANSI(strings.Join(lines[clampedY:height], "\n"))
	for _, label := range []string{"Open", "Save", "Quit"} {
		if !strings.Contains(plainMenu, label) {
			t.Errorf("clamped menu missing %q: %q", label, plainMenu)
		}
	}
}

func TestPopoverKeyboardSelectionAndCallbacks(t *testing.T) {
	calls := make([]string, 0, 1)
	closes := 0
	menu := &warp.Popover{
		Items: []warp.PopoverItem{
			{Name: "First", Action: func() { calls = append(calls, "First") }},
			{Name: "Second", Action: func() { calls = append(calls, "Second") }},
			{Name: "Third", Action: func() { calls = append(calls, "Third") }},
		},
		Width: 14,
		OnClose: func() { closes++ },
	}
	menu.Overlay(popoverBackground(40, 12), 40, 12)

	if !menu.HandleKey(popoverKey(tea.KeyDown)) || !menu.HandleKey(popoverKey(tea.KeyDown)) {
		t.Fatal("down-arrow navigation was not consumed")
	}
	if !menu.HandleKey(popoverKey(tea.KeyUp)) {
		t.Fatal("up-arrow navigation was not consumed")
	}
	if !menu.HandleKey(popoverKey(tea.KeyEnter)) {
		t.Fatal("enter was not consumed")
	}
	if len(calls) != 1 || calls[0] != "Second" || closes != 1 {
		t.Fatalf("keyboard action/close callbacks = %v/%d, want [Second]/1", calls, closes)
	}
	if menu.HandleKey(popoverKey(tea.KeyRunes)) {
		t.Fatal("unhandled key was consumed")
	}

	// Selection is bounded at both ends, and Esc closes without running an action.
	menu.Overlay(popoverBackground(40, 12), 40, 12)
	menu.HandleKey(popoverKey(tea.KeyUp))
	menu.HandleKey(popoverKey(tea.KeyUp))
	menu.HandleKey(popoverKey(tea.KeyEsc))
	if len(calls) != 1 || closes != 2 {
		t.Fatalf("boundary navigation/Esc callbacks = %v/%d, want [Second]/2", calls, closes)
	}
}

func TestPopoverMouseSelectionAndOutsideClickClose(t *testing.T) {
	calls, closes := 0, 0
	menu := &warp.Popover{
		Items: []warp.PopoverItem{
			{Name: "One", Action: func() { calls++ }},
			{Name: "Two", Action: func() { calls++ }},
		},
		X: 50, Y: 8, Width: 10,
		OnClose: func() { closes++ },
	}
	menu.Overlay(popoverBackground(40, 8), 40, 8)
	// The 12x4 framed menu clamps to (28,4); its second item is at row 6.
	if !menu.HandleMouse(popoverMouse(30, 6, tea.MouseActionPress)) {
		t.Fatal("click on second menu item was not consumed")
	}
	if calls != 1 || closes != 1 {
		t.Fatalf("mouse item callbacks=%d close callbacks=%d, want 1/1", calls, closes)
	}

	if !menu.HandleMouse(popoverMouse(0, 0, tea.MouseActionPress)) {
		t.Fatal("outside click was not consumed")
	}
	if calls != 1 || closes != 2 {
		t.Fatalf("outside click callbacks=%d close callbacks=%d, want 1/2", calls, closes)
	}
	if !menu.HandleMouse(popoverMouse(0, 0, tea.MouseActionRelease)) {
		t.Fatal("release was not consumed")
	}
	if menu.HandleMouse(popoverMouse(0, 0, tea.MouseActionMotion)) == false {
		t.Fatal("motion over the active popover was not consumed")
	}
}

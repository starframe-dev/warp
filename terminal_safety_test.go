package warp

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestFrameworkChromeSurvivesMalformedTerminalFragments(t *testing.T) {
	malformed := string([]byte{0x1b, '[', '3', '1'})
	invalid := string([]byte{0xff, 'x'})

	t.Run("tab", func(t *testing.T) {
		group := NewTabGroup(TabTop)
		group.ActiveTab().name = malformed
		view := group.View(30, 4)
		row := strings.Split(view, "\n")[0]
		if !strings.Contains(row, "×") {
			t.Fatalf("tab close glyph swallowed by malformed name: %q", row)
		}
		if got := ansi.StringWidth(row); got != 30 {
			t.Fatalf("tab row width=%d, want 30: %q", got, row)
		}
	})

	t.Run("float", func(t *testing.T) {
		pane := &FloatPane{Panel: BasePanel{}, Width: 16, Height: 4, Title: malformed}
		lines := pane.render(40, 10)
		if len(lines) == 0 || !strings.Contains(lines[0], "×") {
			t.Fatalf("float close glyph swallowed by malformed title: %+v", lines)
		}
		if got := ansi.StringWidth(lines[0]); got != 16 {
			t.Fatalf("float title width=%d, want 16: %q", got, lines[0])
		}
	})

	t.Run("collapsible", func(t *testing.T) {
		panel := NewCollapsible(malformed, BasePanel{})
		row := panel.View(18, 1)
		if !strings.Contains(row, "┐") || !strings.Contains(row, "▼") {
			t.Fatalf("collapsible chrome swallowed: %q", row)
		}
		if got := ansi.StringWidth(row); got != 18 {
			t.Fatalf("collapsible width=%d, want 18: %q", got, row)
		}
	})

	t.Run("dropdown", func(t *testing.T) {
		menu := NewDropdownMenu(malformed, []DropdownItem{{Label: malformed}})
		row := menu.View(18, 2)
		if !strings.Contains(row, "▼") {
			t.Fatalf("dropdown arrow swallowed: %q", row)
		}
		menu.Open = true
		open := strings.Split(menu.View(18, 2), "\n")
		if len(open) != 2 || !strings.Contains(open[0], "▲") {
			t.Fatalf("open dropdown chrome swallowed: %q", open)
		}
		for _, line := range open {
			if got := ansi.StringWidth(line); got != 18 {
				t.Fatalf("dropdown line width=%d, want 18: %q", got, line)
			}
		}
	})

	t.Run("modal", func(t *testing.T) {
		modal := NewModal(malformed, invalid, []ModalButton{{Label: malformed}}, nil)
		lines := make([]string, 10)
		for i := range lines {
			lines[i] = strings.Repeat(" ", 40)
		}
		out := modal.Overlay(lines, 40, 10)
		foundClose := false
		foundButtonBracket := false
		for _, line := range out {
			foundClose = foundClose || strings.Contains(line, "✕")
			foundButtonBracket = foundButtonBracket || strings.Contains(line, "]")
			if got := ansi.StringWidth(line); got != 40 {
				t.Fatalf("modal output width=%d, want 40: %q", got, line)
			}
		}
		if !foundClose || !foundButtonBracket {
			t.Fatalf("modal chrome swallowed: close=%v bracket=%v", foundClose, foundButtonBracket)
		}
	})

	t.Run("input prompt", func(t *testing.T) {
		input := NewInput(malformed)
		view := input.View(20, 1)
		if got := ansi.StringWidth(view); got != 20 {
			t.Fatalf("input width=%d, want 20: %q", got, view)
		}
	})

	t.Run("popover", func(t *testing.T) {
		popover := &Popover{X: 1, Y: 1, Width: 12, Items: []PopoverItem{{Name: malformed}}}
		lines := make([]string, 6)
		for i := range lines {
			lines[i] = strings.Repeat(" ", 30)
		}
		out := popover.Overlay(lines, 30, 6)
		for _, line := range out {
			if got := ansi.StringWidth(line); got != 30 {
				t.Fatalf("popover output width=%d, want 30: %q", got, line)
			}
		}
	})
}

func TestFrameworkLabelsStripTerminalControlSequences(t *testing.T) {
	styled := string([]byte{0x1b, '[', '3', '1', 'm', 'D', 'a', 'n', 'g', 'e', 'r', 0x1b, '[', '0', 'm'})
	if got := sanitizeFrameworkLabel(styled); got != "Danger" {
		t.Fatalf("sanitizeFrameworkLabel styled=%q, want %q", got, "Danger")
	}
	osc := string([]byte{0x1b, ']', '0', ';', 't', 'i', 't', 'l', 'e', 0x07, 'V', 'i', 's', 'i', 'b', 'l', 'e'})
	if got := sanitizeFrameworkLabel(osc); got != "Visible" {
		t.Fatalf("sanitizeFrameworkLabel OSC=%q, want %q", got, "Visible")
	}
}

func TestSemanticNamesUseSanitizedFrameworkLabels(t *testing.T) {
	raw := string([]byte{0x1b, '[', '3', '1', 'm', 'N', 'a', 'm', 'e', 0x1b, '[', '0', 'm'})

	group := NewTabGroup(TabTop)
	group.ActiveTab().name = raw
	if _, ok := FindElement(group.Elements(40, 5), "tab", "Name", "activate-tab"); !ok {
		t.Fatal("tab semantic name was not sanitized")
	}

	collapsible := NewCollapsible(raw, BasePanel{})
	if _, ok := FindElement(collapsible.Elements(20, 3), "button", "Name", "toggle-collapse"); !ok {
		t.Fatal("collapsible semantic name was not sanitized")
	}

	input := NewInput(raw)
	if _, ok := FindElement(input.Elements(20, 1), "textbox", "Name", "focus"); !ok {
		t.Fatal("input semantic name was not sanitized")
	}

	dropdown := NewDropdownMenu(raw, []DropdownItem{{Label: raw}})
	dropdown.Open = true
	elements := dropdown.Elements(20, 3)
	if _, ok := FindElement(elements, "combobox", "Name", "toggle"); !ok {
		t.Fatal("dropdown combobox semantic name was not sanitized")
	}
	if _, ok := FindElement(elements, "option", "Name", "select"); !ok {
		t.Fatal("dropdown option semantic name was not sanitized")
	}
}

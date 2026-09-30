package integration_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	warp "github.com/starframe-dev/warp"
)

const unsafeFrameworkLabel = "wide 界🙂\x1b[31mRED\x1b[0m\tline\x00\x7f\n\r" + "\xff"

func assertTerminalGeometry(t *testing.T, rendered string, width, height int) {
	t.Helper()
	lines := strings.Split(rendered, "\n")
	if len(lines) != height {
		t.Fatalf("rendered %d lines, want %d: %q", len(lines), height, rendered)
	}
	for i, line := range lines {
		if got := ansi.StringWidth(line); got != width {
			t.Errorf("line %d has terminal width %d, want %d: %q", i, got, width, line)
		}
		if !utf8.ValidString(line) {
			t.Errorf("line %d contains invalid UTF-8: %q", i, line)
		}
	}
}

func assertFrameworkLabelSafe(t *testing.T, rendered string) {
	t.Helper()
	plain := ansi.Strip(rendered)
	for _, line := range strings.Split(plain, "\n") {
		for _, control := range []byte{0x00, 0x07, 0x1b, 0x7f, '\r', '\t'} {
			if strings.ContainsRune(line, rune(control)) {
				t.Errorf("plain output contains control character %#x: %q", control, line)
			}
		}
	}
	if strings.Contains(plain, "[31m") || strings.Contains(plain, "[0m") {
		t.Errorf("ANSI sequence from framework label leaked into output: %q", plain)
	}
	if !strings.ContainsRune(plain, '\uFFFD') {
		t.Errorf("invalid UTF-8 was not replaced safely: %q", plain)
	}
	if !strings.Contains(plain, "界🙂") {
		t.Errorf("wide Unicode label was lost: %q", plain)
	}
}

func TestFrameworkOwnedLabelsAreSafeAcrossRenderers(t *testing.T) {
	const width, height = 32, 7

	t.Run("tab labels", func(t *testing.T) {
		group := warp.NewTabGroup(warp.TabTop)
		group.NewTab(unsafeFrameworkLabel)
		rendered := group.View(width, height)
		assertTerminalGeometry(t, rendered, width, height)
		assertFrameworkLabelSafe(t, rendered)
	})

	t.Run("float title", func(t *testing.T) {
		tab := warp.NewTab("float")
		tab.Float(&renderSafetyPanel{}, 1, 1, width-2, height-1)
		lines := strings.Split(tab.View(width, height), "\n")
		if len(lines) != height {
			t.Fatalf("tab rendered %d lines, want %d", len(lines), height)
		}
		for i, line := range lines {
			if got := ansi.StringWidth(line); got != width {
				t.Errorf("line %d has terminal width %d, want %d", i, got, width)
			}
		}
		for _, line := range lines {
			if !utf8.ValidString(line) {
				t.Errorf("float line contains invalid UTF-8: %q", line)
			}
		}
	})

	t.Run("dropdown", func(t *testing.T) {
		menu := warp.NewDropdownMenu(unsafeFrameworkLabel, []warp.DropdownItem{{Label: unsafeFrameworkLabel}})
		menu.Open = true
		rendered := menu.View(width, 3)
		assertTerminalGeometry(t, rendered, width, 2)
		assertFrameworkLabelSafe(t, rendered)
	})

	t.Run("input prompt", func(t *testing.T) {
		input := warp.NewInput(unsafeFrameworkLabel)
		rendered := input.View(width, 3)
		assertTerminalGeometry(t, rendered, width, 3)
		assertFrameworkLabelSafe(t, rendered)
	})

	t.Run("collapsible title", func(t *testing.T) {
		panel := warp.NewCollapsible(unsafeFrameworkLabel, &renderSafetyPanel{})
		rendered := panel.View(width, 4)
		assertTerminalGeometry(t, rendered, width, 4)
		assertFrameworkLabelSafe(t, rendered)
	})

	t.Run("modal title and buttons", func(t *testing.T) {
		modal := warp.NewModal(unsafeFrameworkLabel, "dialog", []warp.ModalButton{{Label: unsafeFrameworkLabel}}, nil)
		lines := make([]string, height)
		for i := range lines {
			lines[i] = strings.Repeat(".", width)
		}
		rendered := modal.Overlay(lines, width, height)
		assertTerminalGeometry(t, strings.Join(rendered, "\n"), width, height)
		assertFrameworkLabelSafe(t, strings.Join(rendered, "\n"))
	})

	t.Run("popover item", func(t *testing.T) {
		popover := &warp.Popover{X: 2, Y: 1, Width: 20, Items: []warp.PopoverItem{{Name: unsafeFrameworkLabel}}}
		lines := make([]string, height)
		for i := range lines {
			lines[i] = strings.Repeat(".", width)
		}
		rendered := popover.Overlay(lines, width, height)
		assertTerminalGeometry(t, strings.Join(rendered, "\n"), width, height)
		assertFrameworkLabelSafe(t, strings.Join(rendered, "\n"))
	})
}

type renderSafetyPanel struct{}

func (*renderSafetyPanel) View(_, height int) string {
	return strings.Repeat("content\n", max(0, height-1)) + "content"
}
func (*renderSafetyPanel) Update(tea.Msg) tea.Cmd { return nil }

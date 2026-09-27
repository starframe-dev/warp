package warp

import (
	"math"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func FuzzPadVisualLine(f *testing.F) {
	for _, seed := range []string{"", "abc", "界🙂", "\x1b[31mred\x1b[0m", "a\u0301"} {
		f.Add(seed, 20)
	}
	f.Fuzz(func(t *testing.T, input string, width int) {
		width = width % 129
		if width < 0 {
			width = -width
		}
		input = strings.ReplaceAll(input, "\n", "")
		got := padVisualLine(input, width)
		if visual := ansi.StringWidth(got); visual != width {
			t.Fatalf("padVisualLine width=%d, want %d for %q", visual, width, input)
		}
	})
}

func FuzzInputEditing(f *testing.F) {
	f.Add("á漢🙂", 3, byte(0))
	f.Add("", 0, byte(4))
	f.Fuzz(func(t *testing.T, value string, cursor int, op byte) {
		input := NewInput("> ")
		input.Focus()
		input.SetValue(value)
		input.SetCursor(cursor)
		switch op % 6 {
		case 0:
			input.Update(tea.KeyMsg{Type: tea.KeyLeft})
		case 1:
			input.Update(tea.KeyMsg{Type: tea.KeyRight})
		case 2:
			input.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		case 3:
			input.Update(tea.KeyMsg{Type: tea.KeyDelete})
		case 4:
			input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'界'}})
		case 5:
			input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\u0301'}})
		}
		if input.Cursor < 0 || input.Cursor > len([]rune(input.Value)) {
			t.Fatalf("cursor=%d outside value rune length=%d", input.Cursor, len([]rune(input.Value)))
		}
		_ = input.View(32, 1)
	})
}

func FuzzSplitGeometry(f *testing.F) {
	for _, seed := range []struct {
		width, height             int
		fraction                  float64
		horizontal, first, second bool
	}{
		{80, 24, 0.5, false, false, false},
		{1, 1, math.NaN(), false, true, false},
		{20, 5, math.Inf(1), true, false, true},
		{20, 5, math.Inf(-1), true, true, true},
	} {
		f.Add(seed.width, seed.height, seed.fraction, seed.horizontal, seed.first, seed.second)
	}
	f.Fuzz(func(t *testing.T, width, height int, fraction float64, horizontal, firstCollapsed, secondCollapsed bool) {
		width %= 257
		height %= 129
		if width < 0 {
			width = -width
		}
		if height < 0 {
			height = -height
		}
		direction := Vertical
		if horizontal {
			direction = Horizontal
		}
		first := &Node{Panel: BasePanel{}}
		second := &Node{Panel: BasePanel{}}
		split := &SplitConfig{
			Direction: direction,
			Fraction:  fraction,
			First:     first,
			Second:    second,
		}
		if firstCollapsed {
			first.Collapse = &NodeCollapse{Active: true}
		}
		if secondCollapsed {
			second.Collapse = &NodeCollapse{Active: true}
		}
		layout := newLayout(&Node{Split: split}, layoutRect{w: width, h: height})
		var walk func(*layoutNode)
		walk = func(node *layoutNode) {
			if node == nil {
				return
			}
			if node.bounds.w < 0 || node.bounds.h < 0 {
				t.Fatalf("negative bounds: %+v", node.bounds)
			}
			for _, child := range node.children {
				if child.bounds.x < node.bounds.x || child.bounds.y < node.bounds.y ||
					child.bounds.x+child.bounds.w > node.bounds.x+node.bounds.w ||
					child.bounds.y+child.bounds.h > node.bounds.y+node.bounds.h {
					t.Fatalf("child %+v outside parent %+v", child.bounds, node.bounds)
				}
				walk(child)
			}
		}
		walk(layout)
		_ = renderLayout(layout)
	})
}

func FuzzFrameworkLabels(f *testing.F) {
	for _, seed := range []string{
		"plain",
		"界🙂",
		"\t",
		"00000",
		string([]byte{'0', 0x0f, '0', '0', '0', '0', '0', '0', 0x1f, '0', '0', '0', '0', '0', 0x86, 0xe6, '0', 0x1b, 0x82, '0'}),
		string([]byte{0x1b, '[', '3', '1', 'm', 'r', 'e', 'd', 0x1b, '[', '0', 'm'}),
		string([]byte{0x1b, '[', '3', '1'}),
		string([]byte{0xff, 'x'}),
	} {
		f.Add(seed, uint8(20))
	}
	f.Add("\t", uint8(3))
	f.Add("00000", uint8(1))
	f.Add(string([]byte{'0', 0x0f, '0', '0', '0', '0', '0', '0', 0x1f, '0', '0', '0', '0', '0', 0x86, 0xe6, '0', 0x1b, 0x82, '0'}), uint8('C'))
	f.Fuzz(func(t *testing.T, label string, rawWidth uint8) {
		width := 10 + int(rawWidth%31)

		collapsible := NewCollapsible(label, BasePanel{})
		if got := ansi.StringWidth(collapsible.View(width, 1)); got != width {
			t.Fatalf("Collapsible width=%d, want %d for %q", got, width, label)
		}

		dropdown := NewDropdownMenu(label, []DropdownItem{{Label: label}})
		if got := ansi.StringWidth(dropdown.View(width, 2)); got != width {
			t.Fatalf("Dropdown width=%d, want %d for %q", got, width, label)
		}

		pane := &FloatPane{Panel: BasePanel{}, Width: width, Height: 3, Title: label}
		for i, line := range pane.render(width, 3) {
			if got := ansi.StringWidth(line); got != width {
				t.Fatalf("Float line %d width=%d, want %d for %q", i, got, width, label)
			}
		}

		group := NewTabGroup(TabTop)
		group.ActiveTab().name = label
		rows := strings.Split(group.View(width, 3), "\n")
		if len(rows) == 0 || ansi.StringWidth(rows[0]) != width {
			t.Fatalf("Tab row width=%d, want %d for %q", ansi.StringWidth(rows[0]), width, label)
		}
	})
}

func FuzzModalContent(f *testing.F) {
	for _, seed := range []string{
		"plain",
		"界🙂",
		string([]byte{0x1b, '[', '3', '1', 'm', 'r', 'e', 'd', 0x1b, '[', '0', 'm'}),
		string([]byte{0x1b, '[', '3', '1'}),
		string([]byte{0xff, 'x'}),
	} {
		f.Add(seed, uint8(40), uint8(10))
	}
	f.Fuzz(func(t *testing.T, content string, rawWidth, rawHeight uint8) {
		width := 1 + int(rawWidth%80)
		height := 1 + int(rawHeight%24)
		modal := NewModal("Title", content, []ModalButton{{Label: "OK"}}, nil)
		lines := make([]string, height)
		for i := range lines {
			lines[i] = strings.Repeat(" ", width)
		}
		out := modal.Overlay(lines, width, height)
		if len(out) != len(lines) {
			t.Fatalf("Modal output lines=%d, want %d", len(out), len(lines))
		}
		for i, line := range out {
			if got := ansi.StringWidth(line); got != width {
				t.Fatalf("Modal line %d width=%d, want %d for content %q; line=%q", i, got, width, content, line)
			}
		}
	})
}

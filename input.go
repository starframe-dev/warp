package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Input is a single-line text input component.
type Input struct {
	Value   string
	Cursor  int // cursor position in runes
	Prompt  string
	Width   int // desired width (0 = auto from View w)
	focused bool
}

// NewInput creates a new empty input with the given prompt.
func NewInput(prompt string) *Input {
	return &Input{
		Value:  "",
		Cursor: 0,
		Prompt: prompt,
	}
}

// SetValue replaces the input value and places the cursor at the end.
func (in *Input) SetValue(v string) {
	in.Value = v
	in.Cursor = len([]rune(v))
	in.clampCursor()
}

// SetCursor sets the cursor position in runes.
func (in *Input) SetCursor(pos int) {
	in.Cursor = pos
	in.clampCursor()
}

// Focused reports whether the input has focus.
func (in *Input) Focused() bool {
	return in.focused
}

// Focus gives the input focus.
func (in *Input) Focus() {
	in.focused = true
}

// Blur removes focus from the input.
func (in *Input) Blur() {
	in.focused = false
}

// View renders the input. If height >= 3 it draws a bordered box.
func (in *Input) View(w, h int) string {
	if h >= 3 {
		return in.viewBoxed(w, h)
	}
	return in.viewInline(w, h)
}

func (in *Input) viewBoxed(w, h int) string {
	borderStyle := inputBorderStyle
	if in.focused {
		borderStyle = inputFocusBorderStyle
	}

	innerW := w - 2
	if innerW < 1 {
		innerW = 1
	}
	innerH := h - 2
	if innerH < 1 {
		innerH = 1
	}

	contentLine := in.renderLine(innerW)
	contentLine = padRight(contentLine, innerW)

	lines := make([]string, h)
	top := "╭" + strings.Repeat("─", innerW) + "╮"
	bottom := "╰" + strings.Repeat("─", innerW) + "╯"

	lines[0] = borderStyle.Render(top)
	for i := 1; i < h-1; i++ {
		if i == (h-2)/2+1 || i == 1 {
			lines[i] = borderStyle.Render("│") + inputStyle.Render(contentLine) + borderStyle.Render("│")
		} else {
			lines[i] = borderStyle.Render("│") + strings.Repeat(" ", innerW) + borderStyle.Render("│")
		}
	}
	lines[h-1] = borderStyle.Render(bottom)

	return strings.Join(lines, "\n")
}

func (in *Input) viewInline(w, h int) string {
	line := in.renderLine(w)
	line = padRight(line, w)
	lines := make([]string, h)
	for i := range lines {
		lines[i] = inputStyle.Render(line)
	}
	return strings.Join(lines, "\n")
}

// renderLine builds the prompt + value line with cursor highlight.
func (in *Input) renderLine(maxW int) string {
	prefix := in.Prompt
	visPrefix := lipgloss.Width(prefix)

	value := in.Value
	if visPrefix+lipgloss.Width(value) > maxW && maxW > visPrefix {
		// Truncate value from the left so cursor stays visible
		avail := maxW - visPrefix
		value = truncateTailToWidth(value, avail, in.Cursor)
	}

	// Build value with cursor highlight
	var result strings.Builder
	result.WriteString(prefix)

	curPos := 0
	for _, r := range value {
		if curPos == in.Cursor {
			result.WriteString("\x1b[7m")
			result.WriteRune(r)
			result.WriteString("\x1b[0m")
		} else {
			result.WriteRune(r)
		}
		curPos++
	}
	if in.Cursor >= curPos {
		result.WriteString("\x1b[7m \x1b[0m")
	}

	return result.String()
}

// truncateTailToWidth keeps the value visible around the cursor.
func truncateTailToWidth(s string, maxW, cursor int) string {
	if maxW <= 0 {
		return ""
	}
	// Simple strategy: shift start so cursor is near the end of visible area
	var runes []rune
	for _, r := range s {
		runes = append(runes, r)
	}
	if cursor < maxW {
		return string(runes[:min(maxW, len(runes))])
	}
	start := cursor - maxW + 1
	if start < 0 {
		start = 0
	}
	if start > len(runes) {
		start = len(runes)
	}
	end := start + maxW
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

// Update handles keyboard input.
func (in *Input) Update(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !in.focused {
		return nil
	}

	if in.Cursor < 0 {
		in.Cursor = 0
	}
	if in.Cursor > len([]rune(in.Value)) {
		in.Cursor = len([]rune(in.Value))
	}

	switch key.String() {
	case "backspace":
		in.deleteBeforeCursor()
	case "delete":
		in.deleteAtCursor()
	case "left":
		if in.Cursor > 0 {
			in.Cursor--
		}
	case "right":
		if in.Cursor < len([]rune(in.Value)) {
			in.Cursor++
		}
	case "home":
		in.Cursor = 0
	case "end":
		in.Cursor = len([]rune(in.Value))
	case "tab", "shift+tab":
		// Handled by parent focus traversal
	case "enter":
		// Submit — could return a custom message; for now no-op
	default:
		if len(key.String()) == 1 || len([]rune(key.String())) == 1 {
			in.insertAtCursor(key.String())
		}
	}
	in.clampCursor()
	return nil
}

func (in *Input) insertAtCursor(s string) {
	if s == "" {
		return
	}
	runes := []rune(in.Value)
	if in.Cursor > len(runes) {
		in.Cursor = len(runes)
	}
	runes = append(runes[:in.Cursor], append([]rune(s), runes[in.Cursor:]...)...)
	in.Value = string(runes)
	in.Cursor += len([]rune(s))
}

func (in *Input) deleteBeforeCursor() {
	runes := []rune(in.Value)
	if in.Cursor <= 0 || len(runes) == 0 {
		return
	}
	runes = append(runes[:in.Cursor-1], runes[in.Cursor:]...)
	in.Value = string(runes)
	in.Cursor--
}

func (in *Input) deleteAtCursor() {
	runes := []rune(in.Value)
	if in.Cursor < 0 || in.Cursor >= len(runes) {
		return
	}
	runes = append(runes[:in.Cursor], runes[in.Cursor+1:]...)
	in.Value = string(runes)
}

func (in *Input) clampCursor() {
	n := len([]rune(in.Value))
	if in.Cursor < 0 {
		in.Cursor = 0
	}
	if in.Cursor > n {
		in.Cursor = n
	}
}

var (
	inputStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color(gbLight1))
	inputBorderStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(gbDark4))
	inputFocusBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(gbBlue))
)

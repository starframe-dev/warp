package warp

import (
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
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
	in.Cursor = utf8.RuneCountInString(v)
	in.clampCursor()
}

// SetCursor sets the cursor position in runes. Positions inside a grapheme cluster
// are normalized to that cluster's end.
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
	w = max(0, w)
	h = max(0, h)
	if h >= 3 && w >= 3 {
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

	contentLine := in.renderLine(innerW)
	contentLine = padRight(contentLine, innerW)

	lines := make([]string, h)
	top := "╭" + strings.Repeat("─", innerW) + "╮"
	bottom := "╰" + strings.Repeat("─", innerW) + "╯"

	lines[0] = borderStyle.Render(top)
	contentRow := 1 + (h-3)/2
	for i := 1; i < h-1; i++ {
		if i == contentRow {
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

// renderLine builds the prompt and value, keeping the cursor visible in terminal cells.
func (in *Input) renderLine(maxW int) string {
	maxW = max(0, maxW)
	in.clampCursor()
	if maxW == 0 {
		return ""
	}
	prefix := ansi.Truncate(in.Prompt, maxW, "")
	prefixWidth := ansi.StringWidth(prefix)
	if prefixWidth >= maxW {
		return prefix
	}

	value, cursor := truncateInputAtCursor(in.Value, maxW-prefixWidth, in.Cursor)
	var result strings.Builder
	result.WriteString(prefix)

	runePos := 0
	graphemes := uniseg.NewGraphemes(value)
	for graphemes.Next() {
		cluster := graphemes.Str()
		runeCount := utf8.RuneCountInString(cluster)
		if cursor >= runePos && cursor < runePos+runeCount {
			result.WriteString("\x1b[7m")
			result.WriteString(cluster)
			result.WriteString("\x1b[0m")
		} else {
			result.WriteString(cluster)
		}
		runePos += runeCount
	}
	if cursor >= runePos && ansi.StringWidth(result.String()) < maxW {
		result.WriteString("\x1b[7m \x1b[0m")
	}
	return result.String()
}

func truncateInputAtCursor(value string, maxCells, cursor int) (string, int) {
	if maxCells <= 0 {
		return "", 0
	}
	type cluster struct {
		text                 string
		runeStart, runeEnd   int
		cellStart, cellWidth int
	}
	var clusters []cluster
	runePos, cellPos := 0, 0
	graphemes := uniseg.NewGraphemes(value)
	for graphemes.Next() {
		text := graphemes.Str()
		runeCount := utf8.RuneCountInString(text)
		width := ansi.StringWidth(text)
		clusters = append(clusters, cluster{
			text: text, runeStart: runePos, runeEnd: runePos + runeCount,
			cellStart: cellPos, cellWidth: width,
		})
		runePos += runeCount
		cellPos += width
	}
	cursor = min(max(0, cursor), runePos)
	if cellPos <= maxCells {
		return value, cursor
	}

	cursorCell := cellPos
	for _, current := range clusters {
		if cursor <= current.runeStart {
			cursorCell = current.cellStart
			break
		}
		if cursor < current.runeEnd {
			cursorCell = current.cellStart + current.cellWidth
			break
		}
		cursorCell = current.cellStart + current.cellWidth
	}
	startCell := max(0, cursorCell-maxCells+1)
	endCell := startCell + maxCells
	var result strings.Builder
	resultRunes := 0
	visibleCursor := -1
	for _, current := range clusters {
		cellEnd := current.cellStart + current.cellWidth
		if current.cellWidth == 0 {
			continue
		}
		start := max(startCell, current.cellStart)
		end := min(endCell, cellEnd)
		if start >= end {
			continue
		}
		before := resultRunes
		if start == current.cellStart && end == cellEnd {
			result.WriteString(current.text)
			resultRunes += current.runeEnd - current.runeStart
		} else {
			result.WriteString(strings.Repeat(" ", end-start))
			resultRunes += end - start
		}
		if cursor >= current.runeStart && cursor <= current.runeEnd {
			visibleCursor = before
			if cursor > current.runeStart {
				visibleCursor = resultRunes
			}
		}
	}
	if visibleCursor < 0 {
		if cursorCell <= startCell {
			visibleCursor = 0
		} else {
			visibleCursor = resultRunes
		}
	}
	return result.String(), visibleCursor
}

// Update handles keyboard input.
func (in *Input) Update(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !in.focused {
		return nil
	}

	in.clampCursor()

	switch key.String() {
	case "backspace":
		in.deleteBeforeCursor()
	case "delete":
		in.deleteAtCursor()
	case "left":
		in.Cursor = previousGraphemeBoundary(in.Value, in.Cursor)
	case "right":
		in.Cursor = nextGraphemeBoundary(in.Value, in.Cursor)
	case "home":
		in.Cursor = 0
	case "end":
		in.Cursor = len([]rune(in.Value))
	case "tab", "shift+tab":
		// Handled by parent focus traversal
	case "enter":
		// Submit — could return a custom message; for now no-op
	default:
		if key.Type == tea.KeyRunes {
			in.insertAtCursor(string(key.Runes))
		}
	}
	in.clampCursor()
	return nil
}

func (in *Input) insertAtCursor(s string) {
	if s == "" {
		return
	}
	in.clampCursor()
	runes := []rune(in.Value)
	runes = append(runes[:in.Cursor], append([]rune(s), runes[in.Cursor:]...)...)
	in.Value = string(runes)
	in.Cursor += utf8.RuneCountInString(s)
	in.clampCursor()
}

func (in *Input) deleteBeforeCursor() {
	in.clampCursor()
	start := previousGraphemeBoundary(in.Value, in.Cursor)
	if start == in.Cursor {
		return
	}
	runes := []rune(in.Value)
	runes = append(runes[:start], runes[in.Cursor:]...)
	in.Value = string(runes)
	in.Cursor = start
	in.clampCursor()
}

func (in *Input) deleteAtCursor() {
	in.clampCursor()
	end := nextGraphemeBoundary(in.Value, in.Cursor)
	if end == in.Cursor {
		return
	}
	runes := []rune(in.Value)
	runes = append(runes[:in.Cursor], runes[end:]...)
	in.Value = string(runes)
	in.clampCursor()
}

func (in *Input) clampCursor() {
	in.Cursor = normalizeGraphemeCursor(in.Value, in.Cursor)
}

func normalizeGraphemeCursor(value string, runePos int) int {
	runePos = max(0, min(runePos, utf8.RuneCountInString(value)))
	if runePos == 0 {
		return 0
	}

	boundary := 0
	graphemes := uniseg.NewGraphemes(value)
	for graphemes.Next() {
		boundary += utf8.RuneCountInString(graphemes.Str())
		if runePos <= boundary {
			return boundary
		}
	}
	return boundary
}

func previousGraphemeBoundary(value string, runePos int) int {
	runePos = max(0, min(runePos, utf8.RuneCountInString(value)))
	boundary := 0
	graphemes := uniseg.NewGraphemes(value)
	for graphemes.Next() {
		if runePos <= boundary {
			return boundary
		}
		end := boundary + utf8.RuneCountInString(graphemes.Str())
		if runePos <= end {
			return boundary
		}
		boundary = end
	}
	return boundary
}

func nextGraphemeBoundary(value string, runePos int) int {
	runePos = max(0, min(runePos, utf8.RuneCountInString(value)))
	boundary := 0
	graphemes := uniseg.NewGraphemes(value)
	for graphemes.Next() {
		boundary += utf8.RuneCountInString(graphemes.Str())
		if runePos < boundary {
			return boundary
		}
	}
	return boundary
}

var (
	inputStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color(gbLight1))
	inputBorderStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(gbDark4))
	inputFocusBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(gbBlue))
)

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

	graphemeValue  string
	graphemes      []inputGrapheme
	graphemesValid bool
}

type inputGrapheme struct {
	byteStart, byteEnd int
	runeStart, runeEnd int
	cellStart          int
	cellWidth          int
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
	in.Cursor = in.graphemeRuneCount()
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

	value, cursor := truncateInputAtCursor(in.Value, maxW-prefixWidth, in.Cursor, in.graphemeLayout())
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

func (in *Input) graphemeLayout() []inputGrapheme {
	if in.graphemesValid && in.graphemeValue == in.Value {
		return in.graphemes
	}

	in.graphemes = in.graphemes[:0]
	bytePos, runePos, cellPos := 0, 0, 0
	graphemes := uniseg.NewGraphemes(in.Value)
	for graphemes.Next() {
		text := graphemes.Str()
		runeCount := utf8.RuneCountInString(text)
		width := ansi.StringWidth(text)
		in.graphemes = append(in.graphemes, inputGrapheme{
			byteStart: bytePos, byteEnd: bytePos + len(text),
			runeStart: runePos, runeEnd: runePos + runeCount,
			cellStart: cellPos, cellWidth: width,
		})
		bytePos += len(text)
		runePos += runeCount
		cellPos += width
	}
	in.graphemeValue = in.Value
	in.graphemesValid = true
	return in.graphemes
}

func (in *Input) graphemeRuneCount() int {
	graphemes := in.graphemeLayout()
	if len(graphemes) == 0 {
		return 0
	}
	return graphemes[len(graphemes)-1].runeEnd
}

func truncateInputAtCursor(value string, maxCells, cursor int, clusters []inputGrapheme) (string, int) {
	if maxCells <= 0 {
		return "", 0
	}
	runeCount := 0
	cellWidth := 0
	if len(clusters) > 0 {
		last := clusters[len(clusters)-1]
		runeCount = last.runeEnd
		cellWidth = last.cellStart + last.cellWidth
	}
	cursor = min(max(0, cursor), runeCount)
	if cellWidth <= maxCells {
		return value, cursor
	}

	cursorCell := cellWidth
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
			result.WriteString(value[current.byteStart:current.byteEnd])
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
		in.Cursor = previousGraphemeBoundary(in.graphemeLayout(), in.Cursor)
	case "right":
		in.Cursor = nextGraphemeBoundary(in.graphemeLayout(), in.Cursor)
	case "home":
		in.Cursor = 0
	case "end":
		in.Cursor = in.graphemeRuneCount()
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
	start := previousGraphemeBoundary(in.graphemeLayout(), in.Cursor)
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
	end := nextGraphemeBoundary(in.graphemeLayout(), in.Cursor)
	if end == in.Cursor {
		return
	}
	runes := []rune(in.Value)
	runes = append(runes[:in.Cursor], runes[end:]...)
	in.Value = string(runes)
	in.clampCursor()
}

func (in *Input) clampCursor() {
	in.Cursor = normalizeGraphemeCursor(in.graphemeLayout(), in.Cursor)
}

func normalizeGraphemeCursor(graphemes []inputGrapheme, runePos int) int {
	runeCount := 0
	if len(graphemes) > 0 {
		runeCount = graphemes[len(graphemes)-1].runeEnd
	}
	runePos = max(0, min(runePos, runeCount))
	if runePos == 0 {
		return 0
	}
	for _, current := range graphemes {
		if runePos <= current.runeEnd {
			return current.runeEnd
		}
	}
	return runeCount
}

func previousGraphemeBoundary(graphemes []inputGrapheme, runePos int) int {
	runeCount := 0
	if len(graphemes) > 0 {
		runeCount = graphemes[len(graphemes)-1].runeEnd
	}
	runePos = max(0, min(runePos, runeCount))
	boundary := 0
	for _, current := range graphemes {
		if runePos <= boundary {
			return boundary
		}
		if runePos <= current.runeEnd {
			return boundary
		}
		boundary = current.runeEnd
	}
	return boundary
}

func nextGraphemeBoundary(graphemes []inputGrapheme, runePos int) int {
	runeCount := 0
	if len(graphemes) > 0 {
		runeCount = graphemes[len(graphemes)-1].runeEnd
	}
	runePos = max(0, min(runePos, runeCount))
	for _, current := range graphemes {
		if runePos < current.runeEnd {
			return current.runeEnd
		}
	}
	return runeCount
}

var (
	inputStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color(gbLight1))
	inputBorderStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(gbDark4))
	inputFocusBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(gbBlue))
)

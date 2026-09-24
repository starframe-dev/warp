package warp

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

// WordWrap wraps text at word boundaries so no line exceeds width.
// Words longer than width are broken mid-word.
func WordWrap(text string, width int) []string {
	if width <= 0 {
		return nil
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, wrapLine(line, width)...)
	}
	return result
}

// SpaceWrap wraps text at spaces so no line exceeds width.
// Unlike WordWrap, this does NOT break words — words longer than width overflow.
func SpaceWrap(text string, width int) []string {
	if width <= 0 {
		return nil
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, wrapAtSpaces(line, width)...)
	}
	return result
}

func wrapLine(line string, width int) []string {
	if width <= 0 {
		return nil
	}
	lines := strings.Split(ansi.Wrap(line, width, " "), "\n")
	for i, wrapped := range lines {
		if ansi.StringWidth(wrapped) > width {
			lines[i] = ansi.Truncate(wrapped, width, "")
		}
	}
	return lines
}

func wrapAtSpaces(line string, width int) []string {
	if width <= 0 {
		return nil
	}
	return strings.Split(ansi.Wordwrap(line, width, " "), "\n")
}

func isWordBreak(b byte) bool {
	return unicode.IsSpace(rune(b))
}

// WrapToString joins wrapped lines with "\n".
func WrapToString(text string, width int, useSpaceWrap bool) string {
	var lines []string
	if useSpaceWrap {
		lines = SpaceWrap(text, width)
	} else {
		lines = WordWrap(text, width)
	}
	return strings.Join(lines, "\n")
}

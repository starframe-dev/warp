package warp

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func makeLine(prefix string, width int) string {
	pw := ansi.StringWidth(prefix)
	if pw >= width {
		return prefix
	}
	return prefix + strings.Repeat(" ", width-pw)
}

func TestModalOverlayExactContent(t *testing.T) {
	tests := []struct {
		name       string
		totalW     int
		content    []string
		wantLeft   string
		wantRight  string
	}{
		{
			name:   "80x24 delete chat",
			totalW: 80,
			content: []string{
				strings.Repeat(" ", 80),
				makeLine("  - \U0001f4c2 My Projects", 80),
				makeLine("    \U0001f4ac notes", 80),
				makeLine("  - \U0001f4c2 Work", 80),
				makeLine("    \U0001f4ac meetings", 80),
				strings.Repeat(" ", 80),
				strings.Repeat(" ", 80),
				strings.Repeat(" ", 80),
			},
			wantLeft:  " - 📂",
			wantRight: "   💬",
		},
		{
			name:   "40x10 narrow",
			totalW: 40,
			content: []string{
				strings.Repeat(" ", 40),
				makeLine("  - \U0001f4c2 My Projects", 40),
				makeLine("    \U0001f4ac notes", 40),
				makeLine("  - \U0001f4c2 Work", 40),
				makeLine("    \U0001f4ac meetings", 40),
				strings.Repeat(" ", 40),
				strings.Repeat(" ", 40),
				strings.Repeat(" ", 40),
			},
			wantLeft:  " - ",
			wantRight: "     ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modal := NewModal("Delete chat", `"notes"?`,
				[]ModalButton{
					{Label: "Del", Action: func() {}},
					{Label: "Esc", Action: func() {}},
				},
				nil,
			)
			result := modal.Overlay(tt.content, tt.totalW, len(tt.content))

			t.Logf("Modal: startX=%d, startY=%d, boxWidth=%d, boxHeight=%d",
				modal.startX, modal.startY, modal.boxWidth, modal.boxHeight)
			for i, line := range result {
				t.Logf("line %3d (w=%d): %q", i, ansi.StringWidth(line), line)
			}
		})
	}
}

func TestModalOverlayWithSelectedItem(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#458588")).
		Foreground(lipgloss.Color("#282828"))

	lines := []string{
		strings.Repeat(" ", 80),
		selectedStyle.Render(makeLine("  - \U0001f4c2 My Projects", 80)),
		makeLine("    \U0001f4ac notes", 80),
		makeLine("  - \U0001f4c2 Work", 80),
		makeLine("    \U0001f4ac meetings", 80),
		strings.Repeat(" ", 80),
		strings.Repeat(" ", 80),
		strings.Repeat(" ", 80),
	}

	modal := NewModal("Delete chat", `"notes"?`,
		[]ModalButton{
			{Label: "Del", Action: func() {}},
			{Label: "Esc", Action: func() {}},
		},
		nil,
	)
	result := modal.Overlay(lines, 80, 24)

	t.Logf("Modal: startX=%d, startY=%d, boxWidth=%d, boxHeight=%d",
		modal.startX, modal.startY, modal.boxWidth, modal.boxHeight)
	for i, line := range result {
		visWidth := ansi.StringWidth(line)
		if visWidth != 80 {
			t.Errorf("line %d width = %d, want 80", i, visWidth)
		}
		t.Logf("line %3d (w=%d): %q", i, visWidth, line)
	}
}
package warp

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestModalOverlayPreservesContent(t *testing.T) {
	totalW, totalH := 40, 10

	modal := NewModal(
		"Delete chat", `"notes"?`,
		[]ModalButton{
			{Label: "Del", Action: func() {}},
			{Label: "Esc", Action: func() {}},
		},
		nil,
	)

	lines := []string{
		strings.Repeat(" ", totalW),
		"  - " + "📂 My Projects" + strings.Repeat(" ", totalW-17),
		"    💬 notes" + strings.Repeat(" ", totalW-20),
		"  - 📂 Work" + strings.Repeat(" ", totalW-18),
		"    💬 meetings" + strings.Repeat(" ", totalW-22),
		strings.Repeat(" ", totalW),
		strings.Repeat(" ", totalW),
		strings.Repeat(" ", totalW),
	}
	result := modal.Overlay(lines, totalW, totalH)

	t.Logf("Modal: startX=%d, startY=%d, boxWidth=%d, boxHeight=%d", modal.startX, modal.startY, modal.boxWidth, modal.boxHeight)
	for i, line := range result {
		visWidth := ansi.StringWidth(line)
		want := totalW
		if visWidth != want {
			t.Logf("  line %d (w=%d): %q", i, visWidth, line)
		} else {
			t.Logf("  line %d (w=%d): ok", i, visWidth)
		}
	}

	if ansi.StringWidth(result[0]) != totalW {
		t.Errorf("line 0 width = %d, want %d", ansi.StringWidth(result[0]), totalW)
	}
}

func TestModalOverlayWithANSIContent(t *testing.T) {
	totalW, totalH := 80, 24

	modal := NewModal(
		"Delete chat", `"notes"?`,
		[]ModalButton{
			{Label: "Del", Action: func() {}},
			{Label: "Esc", Action: func() {}},
		},
		nil,
	)

	lines := make([]string, totalH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", totalW)
	}
	result := modal.Overlay(lines, totalW, totalH)

	t.Logf("Modal: startX=%d, startY=%d, boxWidth=%d, boxHeight=%d", modal.startX, modal.startY, modal.boxWidth, modal.boxHeight)
	for i, line := range result {
		visWidth := ansi.StringWidth(line)
		if visWidth != totalW {
			t.Errorf("line %d width = %d, want %d", i, visWidth, totalW)
		}
	}
}
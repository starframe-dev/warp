package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestModalMouseButtons(t *testing.T) {
	delClicked := false
	escClicked := false
	closeClicked := false

	modal := NewModal("Delete chat", `"notes"?`,
		[]ModalButton{
			{Label: "Del", Action: func() { delClicked = true }},
			{Label: "Esc", Action: func() { escClicked = true }},
		},
		func() { closeClicked = true },
	)
	modal.Width = 48

	totalW, totalH := 80, 24
	lines := make([]string, totalH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", totalW)
	}
	lines = modal.Overlay(lines, totalW, totalH)

	t.Logf("Rendered modal:")
	for i, l := range lines {
		t.Logf("  %3d: %q (w=%d)", i, l, ansi.StringWidth(l))
	}
	t.Logf("startX=%d startY=%d boxWidth=%d boxHeight=%d",
		modal.startX, modal.startY, modal.boxWidth, modal.boxHeight)

	startY := modal.startY
	xBtnY := startY + 2
	btnY := startY + 4
	titleY := startY + 1
	xBtnX := modal.startX + modal.boxWidth - 4

	tests := []struct {
		name   string
		x, y   int
		action tea.MouseAction
		check  func() bool
		fail   string
	}{
		{
			name:   "click ✕ close button",
			x:      xBtnX,
			y:      xBtnY,
			action: tea.MouseActionPress,
			check:  func() bool { return closeClicked },
			fail:   "✕ not triggered",
		},
		{
			name:   "click [Del] button",
			x:      modal.startX + 3,
			y:      btnY,
			action: tea.MouseActionPress,
			check:  func() bool { return delClicked },
			fail:   "[Del] not triggered",
		},
		{
			name:   "click [Esc] button",
			x:      modal.startX + 3 + 7,
			y:      btnY,
			action: tea.MouseActionPress,
			check:  func() bool { return escClicked },
			fail:   "[Esc] not triggered",
		},
		{
			name:   "drag on title strip",
			x:      modal.startX + 10,
			y:      titleY,
			action: tea.MouseActionPress,
			check:  func() bool { return modal.dragging },
			fail:   "drag not started",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delClicked = false
			escClicked = false
			closeClicked = false

			modal.HandleMouse(tea.MouseMsg{
				X:      tt.x,
				Y:      tt.y,
				Action: tt.action,
				Button: tea.MouseButtonLeft,
			})

			if !tt.check() {
				t.Errorf("%s: mouse at (%d,%d) action=%v", tt.fail, tt.x, tt.y, tt.action)
			}
		})
	}
}
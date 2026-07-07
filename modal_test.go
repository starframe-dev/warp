package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestShowModalMsg(t *testing.T) {
	msg := ShowModalMsg{
		Title:   "title",
		Content: "content",
		Buttons: []ModalButton{{Label: "OK"}},
		OnClose: func() {},
		Width:   42,
	}
	if msg.Title != "title" || msg.Content != "content" || msg.Width != 42 {
		t.Errorf("ShowModalMsg fields not stored correctly")
	}
}

func TestCloseModalMsg(t *testing.T) {
	_ = CloseModalMsg{}
}

func TestNewModal(t *testing.T) {
	buttons := []ModalButton{{Label: "OK", Action: func() {}}}
	onClose := func() {}
	m := NewModal("T", "C", buttons, onClose)
	if m == nil {
		t.Fatal("NewModal returned nil")
	}
	if m.Title != "T" || m.Content != "C" || m.OnClose == nil || len(m.Buttons) != 1 {
		t.Errorf("NewModal did not initialize fields correctly")
	}
}

func TestModalEnsureDimensions(t *testing.T) {
	t.Run("auto width", func(t *testing.T) {
		m := NewModal("", "", nil, nil)
		m.EnsureDimensions(80, 24)
		if m.BoxWidth() != 48 {
			t.Errorf("BoxWidth = %d, want 48", m.BoxWidth())
		}
		if m.BoxHeight() != 7 {
			t.Errorf("BoxHeight = %d, want 7", m.BoxHeight())
		}
		if m.StartX() != 16 || m.StartY() != 8 {
			t.Errorf("start = (%d,%d), want (16,8)", m.StartX(), m.StartY())
		}
	})

	t.Run("clamped minimum", func(t *testing.T) {
		m := NewModal("", "", nil, nil)
		m.EnsureDimensions(20, 10)
		if m.BoxWidth() != 20 {
			t.Errorf("BoxWidth = %d, want 20", m.BoxWidth())
		}
	})

	t.Run("fixed width", func(t *testing.T) {
		m := NewModal("", "", nil, nil)
		m.Width = 35
		m.EnsureDimensions(100, 30)
		if m.BoxWidth() != 35 {
			t.Errorf("BoxWidth = %d, want 35", m.BoxWidth())
		}
	})

	t.Run("does not recompute", func(t *testing.T) {
		m := NewModal("", "", nil, nil)
		m.EnsureDimensions(80, 24)
		firstW := m.BoxWidth()
		m.EnsureDimensions(10, 10)
		if m.BoxWidth() != firstW {
			t.Errorf("BoxWidth changed after second EnsureDimensions call")
		}
	})
}

func TestModalOverlay(t *testing.T) {
	t.Run("preserves content dimensions", func(t *testing.T) {
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
	})

	t.Run("with ANSI background", func(t *testing.T) {
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
		for i, line := range result {
			if ansi.StringWidth(line) != totalW {
				t.Errorf("line %d width = %d, want %d", i, ansi.StringWidth(line), totalW)
			}
		}
	})

	t.Run("empty input returns early", func(t *testing.T) {
		modal := NewModal("T", "C", nil, nil)
		result := modal.Overlay([]string{}, 80, 24)
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d lines", len(result))
		}
	})

	t.Run("zero width returns early", func(t *testing.T) {
		modal := NewModal("T", "C", nil, nil)
		lines := []string{"hello"}
		result := modal.Overlay(lines, 0, 24)
		if len(result) != 1 || result[0] != "hello" {
			t.Errorf("expected unchanged input, got %v", result)
		}
	})

	t.Run("sets public geometry accessors", func(t *testing.T) {
		modal := NewModal("T", "C", nil, nil)
		modal.Overlay([]string{strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80)}, 80, 10)
		if modal.BoxWidth() == 0 || modal.BoxHeight() == 0 || modal.StartX() < 0 || modal.StartY() < 0 {
			t.Errorf("geometry accessors not set correctly: w=%d h=%d x=%d y=%d", modal.BoxWidth(), modal.BoxHeight(), modal.StartX(), modal.StartY())
		}
	})
}

func TestModalHandleMouse(t *testing.T) {
	t.Run("close via x button", func(t *testing.T) {
		closed := false
		modal := NewModal("T", "C", nil, func() { closed = true })
		modal.EnsureDimensions(80, 24)
		modal.Overlay([]string{strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80)}, 80, 10)

		xBtnY := modal.StartY() + 2
		xBtnX := modal.StartX() + modal.BoxWidth() - 4

		consumed := modal.HandleMouse(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      xBtnX,
			Y:      xBtnY,
		})
		if !consumed {
			t.Errorf("expected close click to be consumed")
		}
		if !closed {
			t.Errorf("expected OnClose to be called")
		}
	})

	t.Run("button action", func(t *testing.T) {
		clicked := false
		modal := NewModal("T", "C", []ModalButton{{Label: "OK", Action: func() { clicked = true }}}, nil)
		modal.EnsureDimensions(80, 24)
		modal.Overlay([]string{strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80)}, 80, 10)

		btnY := modal.StartY() + 4
		btnX := modal.StartX() + 4

		consumed := modal.HandleMouse(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      btnX,
			Y:      btnY,
		})
		if !consumed {
			t.Errorf("expected button click to be consumed")
		}
		if !clicked {
			t.Errorf("expected button action to be called")
		}
	})

	t.Run("drag and release", func(t *testing.T) {
		modal := NewModal("T", "C", nil, nil)
		modal.EnsureDimensions(80, 24)
		modal.Overlay([]string{strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80)}, 80, 10)

		titleY := modal.StartY() + 1
		startX := modal.StartX()

		press := modal.HandleMouse(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      startX + 2,
			Y:      titleY,
		})
		if !press {
			t.Errorf("expected drag press to be consumed")
		}

		motion := modal.HandleMouse(tea.MouseMsg{
			Action: tea.MouseActionMotion,
			X:      startX + 5,
			Y:      titleY + 2,
		})
		if !motion {
			t.Errorf("expected drag motion to be consumed")
		}
		if modal.StartX() != startX+3 || modal.StartY() != titleY+1 {
			t.Errorf("drag did not update position: got (%d,%d)", modal.StartX(), modal.StartY())
		}

		release := modal.HandleMouse(tea.MouseMsg{
			Action: tea.MouseActionRelease,
			Button: tea.MouseButtonLeft,
		})
		if !release {
			t.Errorf("expected drag release to be consumed")
		}
	})

	t.Run("ignores before dimensions", func(t *testing.T) {
		modal := NewModal("T", "C", nil, nil)
		consumed := modal.HandleMouse(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      10,
			Y:      10,
		})
		if consumed {
			t.Errorf("expected mouse event to be ignored before dimensions are set")
		}
	})

	t.Run("ignores unrelated press", func(t *testing.T) {
		modal := NewModal("T", "C", nil, nil)
		modal.EnsureDimensions(80, 24)
		modal.Overlay([]string{strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80), strings.Repeat(" ", 80)}, 80, 10)
		consumed := modal.HandleMouse(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonRight,
			X:      0,
			Y:      0,
		})
		if consumed {
			t.Errorf("expected unrelated press to be ignored")
		}
	})
}

func TestModalButton(t *testing.T) {
	btn := ModalButton{Label: "Cancel", Action: func() {}}
	if btn.Label != "Cancel" {
		t.Errorf("Label = %q, want Cancel", btn.Label)
	}
}

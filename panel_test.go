package warp

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBasePanelView(t *testing.T) {
	bp := BasePanel{}
	got := bp.View(80, 24)
	if got != "" {
		t.Errorf("BasePanel.View(80, 24) = %q; want empty string", got)
	}
}

func TestBasePanelUpdate(t *testing.T) {
	bp := BasePanel{}
	cmd := bp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd != nil {
		t.Errorf("BasePanel.Update(...) = %v; want nil", cmd)
	}
}

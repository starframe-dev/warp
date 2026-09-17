package warp

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestBasePanelView verifies that BasePanel.View returns an empty string.
func TestBasePanelView(t *testing.T) {
	bp := BasePanel{}
	got := bp.View(80, 24)
	if got != "" {
		t.Errorf("BasePanel.View(80, 24) = %q; want empty string", got)
	}
}

// TestBasePanelUpdate verifies that BasePanel.Update returns a nil tea.Cmd.
func TestBasePanelUpdate(t *testing.T) {
	bp := BasePanel{}
	cmd := bp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd != nil {
		t.Errorf("BasePanel.Update(...) = %v; want nil", cmd)
	}
}

// TestBasePanelViewZeroSize verifies that View works with zero-sized inputs
// and still returns an empty string.
func TestBasePanelViewZeroSize(t *testing.T) {
	bp := BasePanel{}
	if got := bp.View(0, 0); got != "" {
		t.Fatalf("BasePanel.View(0, 0) = %q; want empty string", got)
	}
}

// TestBasePanelUpdateDifferentMsgs verifies that Update returns nil for
// different kinds of messages and does not panic.
func TestBasePanelUpdateDifferentMsgs(t *testing.T) {
	bp := BasePanel{}
	msgs := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyCtrlC},
		tea.WindowSizeMsg{Width: 100, Height: 40},
		tea.MouseMsg{X: 5, Y: 7},
	}
	for i := range msgs {
		if got := bp.Update(msgs[i]); got != nil {
			t.Errorf("BasePanel.Update(msgs[%d]) = %v; want nil", i, got)
		}
	}
}

// TestPanelInterfaceSatisfied verifies that *BasePanel satisfies the Panel
// interface by using it as a Panel value.
func TestPanelInterfaceSatisfied(t *testing.T) {
	var p Panel = BasePanel{}
	if p == nil {
		t.Fatalf("expected non-nil Panel")
	}
	if got := p.View(60, 30); got != "" {
		t.Errorf("Panel.View(60, 30) = %q; want empty string", got)
	}
	if cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Errorf("Panel.Update(tea.KeyEnter) = %v; want nil", cmd)
	}
}

package integration_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type floatInteractionPanel struct {
	name      string
	lastMouse tea.MouseMsg
	mouseSeen bool
	unmounts  int
}

func (p *floatInteractionPanel) View(_, height int) string {
	if height <= 0 {
		return ""
	}
	lines := make([]string, height)
	for i := range lines {
		lines[i] = p.name
	}
	return strings.Join(lines, "\n")
}

func (p *floatInteractionPanel) Update(msg tea.Msg) tea.Cmd {
	if mouse, ok := msg.(tea.MouseMsg); ok {
		p.lastMouse, p.mouseSeen = mouse, true
	}
	return nil
}

func (p *floatInteractionPanel) Unmount() { p.unmounts++ }

func floatMouse(x, y int, action tea.MouseAction) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: action}
}

func TestTabManagedFloatsRenderInZOrderAndClampToViewport(t *testing.T) {
	tab := warp.NewTab("floats")
	low := &floatInteractionPanel{name: "LOW"}
	high := &floatInteractionPanel{name: "HIGH"}
	tab.Float(low, 8, 2, 18, 7)
	tab.Float(high, 12, 4, 18, 7)

	view := strings.Split(warp.StripANSI(tab.View(32, 10)), "\n")
	if len(view) <= 5 || !strings.Contains(view[5], "LOW") || !strings.Contains(view[5], "HIGH") || strings.Index(view[5], "LOW") > strings.Index(view[5], "HIGH") {
		t.Fatalf("later float did not cover the lower float in their overlap: %v", view)
	}

	// Resizing the viewport clamps the visible dimensions and position.
	tab.View(14, 5)
	clamped := tab.Elements(14, 5)
	close := requireFloatElement(t, clamped, "button", "Close Float", "close-float")
	if close.Bounds.X < 0 || close.Bounds.Y < 0 || close.Bounds.X >= 14 || close.Bounds.Y >= 5 {
		t.Fatalf("clamped float close affordance outside viewport: %+v", close.Bounds)
	}
	view = strings.Split(warp.StripANSI(tab.View(32, 10)), "\n")
	if len(view) <= 5 || !strings.Contains(view[5], "HIGH") {
		t.Fatalf("float failed to render after viewport growth: %v", view)
	}
}

func TestFloatMouseForwardingDragResizeCloseAndSemanticChrome(t *testing.T) {
	tab := warp.NewTab("interaction")
	panel := &floatInteractionPanel{name: "payload"}
	tab.Float(panel, 5, 2, 16, 7)
	tab.Update(tea.WindowSizeMsg{Width: 40, Height: 15})

	fpElements := tab.Elements(40, 15)
	close := requireFloatElement(t, fpElements, "button", "Close Float", "close-float")
	title := requireFloatElement(t, fpElements, "titlebar", "Float", "move-float")
	if close.Bounds != (warp.Bounds{X: 19, Y: 2, W: 1, H: 1}) {
		t.Fatalf("close chrome bounds=%+v", close.Bounds)
	}
	if title.Bounds.X != 6 || title.Bounds.Y != 2 {
		t.Fatalf("titlebar bounds=%+v", title.Bounds)
	}

	// The content begins one cell inside the frame, in both axes.
	tab.HandleMouse(floatMouse(8, 5, tea.MouseActionPress))
	if !panel.mouseSeen || panel.lastMouse.X != 2 || panel.lastMouse.Y != 2 {
		t.Fatalf("forwarded mouse=%+v seen=%v, want content-relative (2,2)", panel.lastMouse, panel.mouseSeen)
	}

	// Drag from the title and verify viewport clamping at the top-left edge.
	tab.HandleMouse(floatMouse(8, 2, tea.MouseActionPress))
	tab.HandleMouse(floatMouse(-20, -20, tea.MouseActionMotion))
	tab.HandleMouse(floatMouse(-20, -20, tea.MouseActionRelease))
	close = requireFloatElement(t, tab.Elements(40, 15), "button", "Close Float", "close-float")
	if close.Bounds.X != 14 || close.Bounds.Y != 0 {
		t.Fatalf("dragged float close bounds=%+v, want x=14 y=0", close.Bounds)
	}

	// Resize the east edge; the opposite edge remains fixed and width grows.
	before := close.Bounds.X
	tab.HandleMouse(floatMouse(15, 3, tea.MouseActionPress))
	tab.HandleMouse(floatMouse(20, 3, tea.MouseActionMotion))
	tab.HandleMouse(floatMouse(20, 3, tea.MouseActionRelease))
	after := requireFloatElement(t, tab.Elements(40, 15), "button", "Close Float", "close-float").Bounds.X
	if after <= before {
		t.Fatalf("east-edge resize did not increase float width: close x before=%d after=%d", before, after)
	}

	// Clicking the semantic × location closes the float and unmounts its panel.
	close = requireFloatElement(t, tab.Elements(40, 15), "button", "Close Float", "close-float")
	tab.HandleMouse(floatMouse(close.Bounds.X, close.Bounds.Y, tea.MouseActionPress))
	if elements := tab.Elements(40, 15); hasFloatElement(elements) {
		t.Fatalf("float chrome remained after close request: %+v", elements)
	}
	if panel.unmounts != 1 {
		t.Fatalf("float panel Unmount calls=%d, want 1", panel.unmounts)
	}
	tab.CloseFloat(nil)
	if panel.unmounts != 1 {
		t.Fatalf("closing nil float changed lifecycle count: %d", panel.unmounts)
	}
}

func requireFloatElement(t *testing.T, elements []warp.Element, role, name, action string) warp.Element {
	t.Helper()
	el, ok := warp.FindElement(elements, role, name, action)
	if !ok {
		t.Fatalf("float element role=%q name=%q action=%q missing in %+v", role, name, action, elements)
	}
	return el
}

func hasFloatElement(elements []warp.Element) bool {
	for _, el := range elements {
		if el.Action == "close-float" || el.Action == "move-float" || el.Name == "payload" {
			return true
		}
	}
	return false
}

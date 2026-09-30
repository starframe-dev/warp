package e2e_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type floatE2EPanel struct {
	content string
}

func (p *floatE2EPanel) View(_, height int) string {
	return strings.Repeat(p.content+"\n", height)
}

func (*floatE2EPanel) Update(tea.Msg) tea.Cmd { return nil }

func findFloatCloseButton(tab *warp.Tab) (warp.Element, bool) {
	for _, element := range tab.Elements(70, 20) {
		if element.Action == "close-float" {
			return element, true
		}
	}
	return warp.Element{}, false
}

func TestFloatBringToFrontDragResizeClampAndClose(t *testing.T) {
	const width, height = 70, 20
	tab := warp.NewTab("floats")
	tab.SetRootPanel(&floatE2EPanel{content: "background"})
	tab.Float(&floatE2EPanel{content: "back"}, 8, 4, 30, 10)
	tab.Float(&floatE2EPanel{content: "front"}, 20, 7, 28, 9)
	tab.Update(tea.WindowSizeMsg{Width: width, Height: height})
	tab.View(width, height)

	// The later float is initially topmost. Click an exposed cell of the
	// earlier float to raise it, then verify its close control is first.
	controls := tab.Elements(width, height)
	if len(controls) == 0 || controls[0].Action != "close-float" || controls[0].Bounds.X != 46 {
		t.Fatalf("initial top float close control = %#v; want the float at x=20 on top", controls)
	}
	tab.HandleMouse(tea.MouseMsg{X: 10, Y: 8, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	controls = tab.Elements(width, height)
	if len(controls) == 0 || controls[0].Action != "close-float" || controls[0].Bounds.X != 36 {
		t.Fatalf("clicking the exposed back float did not raise it: %#v", controls)
	}

	// Drag the raised float by its title bar.
	tab.HandleMouse(tea.MouseMsg{X: 15, Y: 4, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	tab.HandleMouse(tea.MouseMsg{X: 22, Y: 6, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	tab.HandleMouse(tea.MouseMsg{X: 22, Y: 6, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	tab.View(width, height)
	controls = tab.Elements(width, height)
	found := false
	for _, element := range controls {
		if element.Action == "close-float" && element.Bounds.X == 43 && element.Bounds.Y == 6 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("dragged float close control not found at expected location: %#v", controls)
	}

	// Resize from the southeast corner, then drag beyond the viewport; the
	// resulting rectangle must remain fully visible.
	tab.HandleMouse(tea.MouseMsg{X: 44, Y: 15, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	tab.HandleMouse(tea.MouseMsg{X: 100, Y: 40, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	tab.HandleMouse(tea.MouseMsg{X: 100, Y: 40, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	tab.View(width, height)
	controls = tab.Elements(width, height)
	var resized warp.Element
	found = false
	for _, element := range controls {
		if element.Action == "close-float" && element.Bounds.X >= 0 {
			// The resized topmost float is listed first in z-order.
			resized, found = element, true
			break
		}
	}
	if !found {
		t.Fatal("resized float has no close control")
	}
	// The close control is always two cells from the right edge; its
	// position also reveals that the resized rectangle was viewport-clamped.
	if resized.Bounds.X != width-2 || resized.Bounds.Y > height-3 {
		t.Fatalf("resized float close control at (%d,%d), outside or not clamped to viewport", resized.Bounds.X, resized.Bounds.Y)
	}

	tab.HandleMouse(tea.MouseMsg{X: resized.Bounds.X, Y: resized.Bounds.Y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	tab.View(width, height)
	remaining := tab.Elements(width, height)
	for _, element := range remaining {
		if element.Action == "close-float" && element.Bounds.X == resized.Bounds.X && element.Bounds.Y == resized.Bounds.Y {
			t.Fatal("clicking the close control did not remove the resized float")
		}
	}
}

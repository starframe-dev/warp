package warp

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestSetTheme verifies that SetTheme propagates a ThemeColors palette
// into every package-level color and style variable.
func TestSetTheme(t *testing.T) {
	originals := map[string]lipgloss.Color{
		"gbDark0":     gbDark0,
		"gbDark1":     gbDark1,
		"gbDark2":     gbDark2,
		"gbDark3":     gbDark3,
		"gbDark4":     gbDark4,
		"gbGray":      gbGray,
		"gbLight1":    gbLight1,
		"gbRed":       gbRed,
		"gbGreen":     gbGreen,
		"gbYellow":    gbYellow,
		"gbBlue":      gbBlue,
		"tabBarBg":    tabBarBg,
		"activeTabBg": activeTabBg,
		"inactiveFg":   inactiveTabFg,
	}

	defer func() {
		for k, v := range originals {
			switch k {
			case "gbDark0":
				gbDark0 = v
			case "gbDark1":
				gbDark1 = v
			case "gbDark2":
				gbDark2 = v
			case "gbDark3":
				gbDark3 = v
			case "gbDark4":
				gbDark4 = v
			case "gbGray":
				gbGray = v
			case "gbLight1":
				gbLight1 = v
			case "gbRed":
				gbRed = v
			case "gbGreen":
				gbGreen = v
			case "gbYellow":
				gbYellow = v
			case "gbBlue":
				gbBlue = v
			case "tabBarBg":
				tabBarBg = v
			case "activeTabBg":
				activeTabBg = v
			case "inactiveFg":
				inactiveTabFg = v
			}
		}
	}()

	styles := lipgloss.Color("#000001")
	SetTheme(ThemeColors{
		Background:          "#000001",
		Surface:             "#000001",
		Raised:              "#000001",
		Border:              "#000001",
		BorderMuted:         "#000001",
		Text:                "#000001",
		TextMuted:           "#000001",
		TextStrong:          "#000001",
		Accent:              "#000001",
		AccentMuted:         "#000001",
		Error:               "#000001",
		Success:             "#000001",
		Warning:             "#000001",
		SelectionBackground: "#000001",
		SelectionForeground: "#000001",
	})

	if gbDark0 != styles {
		t.Errorf("expected gbDark0 to be %v, got %v", styles, gbDark0)
	}
	if gbDark1 != styles {
		t.Errorf("expected gbDark1 to be %v, got %v", styles, gbDark1)
	}
	if gbDark2 != styles {
		t.Errorf("expected gbDark2 to be %v, got %v", styles, gbDark2)
	}
	if gbDark3 != styles {
		t.Errorf("expected gbDark3 to be %v, got %v", styles, gbDark3)
	}
	if gbDark4 != styles {
		t.Errorf("expected gbDark4 to be %v, got %v", styles, gbDark4)
	}
	if gbGray != styles {
		t.Errorf("expected gbGray to be %v, got %v", styles, gbGray)
	}
	if gbLight1 != styles {
		t.Errorf("expected gbLight1 to be %v, got %v", styles, gbLight1)
	}
	if gbRed != styles {
		t.Errorf("expected gbRed to be %v, got %v", styles, gbRed)
	}
	if gbGreen != styles {
		t.Errorf("expected gbGreen to be %v, got %v", styles, gbGreen)
	}
	if gbYellow != styles {
		t.Errorf("expected gbYellow to be %v, got %v", styles, gbYellow)
	}
	if gbBlue != styles {
		t.Errorf("expected gbBlue to be %v, got %v", styles, gbBlue)
	}

	if tabBarBg != styles {
		t.Errorf("expected tabBarBg to be %v, got %v", styles, tabBarBg)
	}
	if activeTabBg != styles {
		t.Errorf("expected activeTabBg to be %v, got %v", styles, activeTabBg)
	}
	if inactiveTabFg != styles {
		t.Errorf("expected inactiveTabFg to be %v, got %v", styles, inactiveTabFg)
	}
	if newTabFg != styles {
		t.Errorf("expected newTabFg to be %v, got %v", styles, newTabFg)
	}
	if closeTabFg != styles {
		t.Errorf("expected closeTabFg to be %v, got %v", styles, closeTabFg)
	}

	if borderStyle.GetForeground() != styles {
		t.Errorf("expected borderStyle fg %v, got %v", styles, borderStyle.GetForeground())
	}
	if borderHoverStyle.GetForeground() != styles {
		t.Errorf("expected borderHoverStyle fg %v, got %v", styles, borderHoverStyle.GetForeground())
	}
	if borderDragStyle.GetForeground() != styles {
		t.Errorf("expected borderDragStyle fg %v, got %v", styles, borderDragStyle.GetForeground())
	}
	if floatTitleStyle.GetForeground() != styles || floatTitleStyle.GetBackground() != styles {
		t.Errorf("floatTitleStyle mismatch: got bg %v, fg %v", floatTitleStyle.GetBackground(), floatTitleStyle.GetForeground())
	}
	if dimStyle.GetForeground() != styles || dimStyle.GetBackground() != styles {
		t.Errorf("dimStyle mismatch: got bg %v, fg %v", dimStyle.GetBackground(), dimStyle.GetForeground())
	}
	if popoverBaseStyle.GetForeground() != styles || popoverBaseStyle.GetBackground() != styles {
		t.Errorf("popoverBaseStyle mismatch: got bg %v, fg %v", popoverBaseStyle.GetBackground(), popoverBaseStyle.GetForeground())
	}
	b, _, _, _, _ := modalBorderStyle.GetBorder()
	if b == lipgloss.NormalBorder() {
		t.Errorf("expected modalBorderStyle border")
	}
}

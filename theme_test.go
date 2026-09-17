package warp

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestSetTheme(t *testing.T) {
	originals := []lipgloss.Color{
		gbDark0, gbDark1, gbDark2, gbDark3, gbDark4, gbGray,
		gbLight1, gbRed, gbGreen, gbYellow, gbBlue,
	}

	defer func() {
		gbDark0 = originals[0]
		gbDark1 = originals[1]
		gbDark2 = originals[2]
		gbDark3 = originals[3]
		gbDark4 = originals[4]
		gbGray = originals[5]
		gbLight1 = originals[6]
		gbRed = originals[7]
		gbGreen = originals[8]
		gbYellow = originals[9]
		gbBlue = originals[10]
	}()

	c := lipgloss.Color("#123456")
	SetTheme(ThemeColors{
		Background:          "#123456",
		Surface:             "#123456",
		Raised:              "#123456",
		Border:              "#123456",
		BorderMuted:         "#123456",
		Text:                "#123456",
		TextMuted:           "#123456",
		TextStrong:          "#123456",
		Accent:              "#123456",
		AccentMuted:         "#123456",
		Error:               "#123456",
		Success:             "#123456",
		Warning:             "#123456",
		SelectionBackground: "#123456",
		SelectionForeground: "#123456",
	})

	if gbDark0 != c {
		t.Errorf("expected gbDark0 to be %v, got %v", c, gbDark0)
	}
	if gbDark1 != c {
		t.Errorf("expected gbDark1 to be %v, got %v", c, gbDark1)
	}
	if gbDark2 != c {
		t.Errorf("expected gbDark2 to be %v, got %v", c, gbDark2)
	}
	if gbDark3 != c {
		t.Errorf("expected gbDark3 to be %v, got %v", c, gbDark3)
	}
	if gbDark4 != c {
		t.Errorf("expected gbDark4 to be %v, got %v", c, gbDark4)
	}
	if gbGray != c {
		t.Errorf("expected gbGray to be %v, got %v", c, gbGray)
	}
	if gbLight1 != c {
		t.Errorf("expected gbLight1 to be %v, got %v", c, gbLight1)
	}
	if gbRed != c {
		t.Errorf("expected gbRed to be %v, got %v", c, gbRed)
	}
	if gbGreen != c {
		t.Errorf("expected gbGreen to be %v, got %v", c, gbGreen)
	}
	if gbYellow != c {
		t.Errorf("expected gbYellow to be %v, got %v", c, gbYellow)
	}
	if gbBlue != c {
		t.Errorf("expected gbBlue to be %v, got %v", c, gbBlue)
	}

	if tabBarBg != c {
		t.Errorf("expected tabBarBg to be %v, got %v", c, tabBarBg)
	}
	if activeTabBg != c {
		t.Errorf("expected activeTabBg to be %v, got %v", c, activeTabBg)
	}
	if inactiveTabFg != c {
		t.Errorf("expected inactiveTabFg to be %v, got %v", c, inactiveTabFg)
	}
	if newTabFg != c {
		t.Errorf("expected newTabFg to be %v, got %v", c, newTabFg)
	}
	if closeTabFg != c {
		t.Errorf("expected closeTabFg to be %v, got %v", c, closeTabFg)
	}

	if borderStyle.GetForeground() != c {
		t.Errorf("borderStyle fg mismatch: got %v", borderStyle.GetForeground())
	}
	if borderHoverStyle.GetForeground() != c {
		t.Errorf("borderHoverStyle fg mismatch: got %v", borderHoverStyle.GetForeground())
	}
	if borderDragStyle.GetForeground() != c {
		t.Errorf("borderDragStyle fg mismatch: got %v", borderDragStyle.GetForeground())
	}
	if collapsibleStyle.GetForeground() != c {
		t.Errorf("collapsibleStyle fg mismatch: got %v", collapsibleStyle.GetForeground())
	}
	if floatBorderStyle.GetForeground() != c {
		t.Errorf("floatBorderStyle fg mismatch: got %v", floatBorderStyle.GetForeground())
	}
	if floatTitleStyle.GetForeground() != c || floatTitleStyle.GetBackground() != c {
		t.Errorf("floatTitleStyle mismatch: got bg %v, fg %v", floatTitleStyle.GetBackground(), floatTitleStyle.GetForeground())
	}
	if floatCloseStyle.GetForeground() != c {
		t.Errorf("floatCloseStyle fg mismatch: got %v", floatCloseStyle.GetForeground())
	}
	if floatBgStyle.GetBackground() != c {
		t.Errorf("floatBgStyle bg mismatch: got %v", floatBgStyle.GetBackground())
	}
	if dropdownButtonStyle.GetBackground() != c || dropdownButtonStyle.GetForeground() != c {
		t.Errorf("dropdownButtonStyle mismatch: got bg %v, fg %v", dropdownButtonStyle.GetBackground(), dropdownButtonStyle.GetForeground())
	}
	if dropdownItemStyle.GetBackground() != c || dropdownItemStyle.GetForeground() != c {
		t.Errorf("dropdownItemStyle mismatch: got bg %v, fg %v", dropdownItemStyle.GetBackground(), dropdownItemStyle.GetForeground())
	}
	if popoverBaseStyle.GetBackground() != c || popoverBaseStyle.GetForeground() != c {
		t.Errorf("popoverBaseStyle mismatch: got bg %v, fg %v", popoverBaseStyle.GetBackground(), popoverBaseStyle.GetForeground())
	}
	if popoverSelectedStyle.GetBackground() != c || popoverSelectedStyle.GetForeground() != c {
		t.Errorf("popoverSelectedStyle mismatch: got bg %v, fg %v", popoverSelectedStyle.GetBackground(), popoverSelectedStyle.GetForeground())
	}
	if dimStyle.GetBackground() != c || dimStyle.GetForeground() != c {
		t.Errorf("dimStyle mismatch: got bg %v, fg %v", dimStyle.GetBackground(), dimStyle.GetForeground())
	}
	if inputStyle.GetForeground() != c {
		t.Errorf("inputStyle fg mismatch: got %v", inputStyle.GetForeground())
	}
	if inputBorderStyle.GetForeground() != c {
		t.Errorf("inputBorderStyle fg mismatch: got %v", inputBorderStyle.GetForeground())
	}
	if inputFocusBorderStyle.GetForeground() != c {
		t.Errorf("inputFocusBorderStyle fg mismatch: got %v", inputFocusBorderStyle.GetForeground())
	}
}

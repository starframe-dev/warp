package warp

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBorderStyle(t *testing.T) {
	style := BorderStyle()
	if style.GetForeground() != lipgloss.Color(borderColor) {
		t.Fatalf("expected BorderStyle foreground %v, got %v", borderColor, style.GetForeground())
	}
}

func TestBorderDragStyle(t *testing.T) {
	style := BorderDragStyle()
	if style.GetForeground() != lipgloss.Color(borderDragColor) {
		t.Fatalf("expected BorderDragStyle foreground %v, got %v", borderDragColor, style.GetForeground())
	}
}

func TestBorderHoverStyle(t *testing.T) {
	style := BorderHoverStyle()
	if style.GetForeground() != lipgloss.Color(borderHoverColor) {
		t.Fatalf("expected BorderHoverStyle foreground %v, got %v", borderHoverColor, style.GetForeground())
	}
}

func TestTabBarColors(t *testing.T) {
	if tabBarBg != gbDark0 {
		t.Fatalf("expected tabBarBg %v, got %v", gbDark0, tabBarBg)
	}
	if activeTabBg != gbDark2 {
		t.Fatalf("expected activeTabBg %v, got %v", gbDark2, activeTabBg)
	}
	if activeTabFg != gbLight1 {
		t.Fatalf("expected activeTabFg %v, got %v", gbLight1, activeTabFg)
	}
	if inactiveTabFg != gbGray {
		t.Fatalf("expected inactiveTabFg %v, got %v", gbGray, inactiveTabFg)
	}
	if newTabFg != gbGreen {
		t.Fatalf("expected newTabFg %v, got %v", gbGreen, newTabFg)
	}
	if closeTabFg != gbRed {
		t.Fatalf("expected closeTabFg %v, got %v", gbRed, closeTabFg)
	}
}

func TestSplitBorderColors(t *testing.T) {
	if borderColor != gbDark1 {
		t.Fatalf("expected borderColor %v, got %v", gbDark1, borderColor)
	}
	if borderDragColor != gbYellow {
		t.Fatalf("expected borderDragColor %v, got %v", gbYellow, borderDragColor)
	}
	if borderHoverColor != gbDark3 {
		t.Fatalf("expected borderHoverColor %v, got %v", gbDark3, borderHoverColor)
	}
}

func TestFloatPaneColors(t *testing.T) {
	if floatBorderColor != gbGray {
		t.Fatalf("expected floatBorderColor %v, got %v", gbGray, floatBorderColor)
	}
	if floatTitleBg != gbDark1 {
		t.Fatalf("expected floatTitleBg %v, got %v", gbDark1, floatTitleBg)
	}
	if floatTitleFg != gbLight1 {
		t.Fatalf("expected floatTitleFg %v, got %v", gbLight1, floatTitleFg)
	}
	if floatBg != gbDark0 {
		t.Fatalf("expected floatBg %v, got %v", gbDark0, floatBg)
	}
	if floatCloseFg != gbRed {
		t.Fatalf("expected floatCloseFg %v, got %v", gbRed, floatCloseFg)
	}
}

func TestTabBarStyles(t *testing.T) {
	if tabBarStyle.GetBackground() != lipgloss.Color(tabBarBg) {
		t.Fatalf("expected tabBarStyle bg %v, got %v", tabBarBg, tabBarStyle.GetBackground())
	}
	if activeTabStyle.GetBackground() != lipgloss.Color(activeTabBg) {
		t.Fatalf("expected activeTabStyle bg %v, got %v", activeTabBg, activeTabStyle.GetBackground())
	}
	if activeTabStyle.GetForeground() != lipgloss.Color(activeTabFg) {
		t.Fatalf("expected activeTabStyle fg %v, got %v", activeTabFg, activeTabStyle.GetForeground())
	}
	if !activeTabStyle.GetBold() {
		t.Fatalf("expected activeTabStyle bold")
	}
	if inactiveTabStyle.GetBackground() != lipgloss.Color(tabBarBg) {
		t.Fatalf("expected inactiveTabStyle bg %v, got %v", tabBarBg, inactiveTabStyle.GetBackground())
	}
	if inactiveTabStyle.GetForeground() != lipgloss.Color(inactiveTabFg) {
		t.Fatalf("expected inactiveTabStyle fg %v, got %v", inactiveTabFg, inactiveTabStyle.GetForeground())
	}
	if newTabStyle.GetForeground() != lipgloss.Color(newTabFg) {
		t.Fatalf("expected newTabStyle fg %v, got %v", newTabFg, newTabStyle.GetForeground())
	}
	if closeTabStyle.GetForeground() != lipgloss.Color(closeTabFg) {
		t.Fatalf("expected closeTabStyle fg %v, got %v", closeTabFg, closeTabStyle.GetForeground())
	}
}

func TestBorderStyles(t *testing.T) {
	if borderStyle.GetForeground() != lipgloss.Color(borderColor) {
		t.Fatalf("expected borderStyle fg %v, got %v", borderColor, borderStyle.GetForeground())
	}
	if borderHoverStyle.GetForeground() != lipgloss.Color(borderHoverColor) {
		t.Fatalf("expected borderHoverStyle fg %v, got %v", borderHoverColor, borderHoverStyle.GetForeground())
	}
	if borderDragStyle.GetForeground() != lipgloss.Color(borderDragColor) {
		t.Fatalf("expected borderDragStyle fg %v, got %v", borderDragColor, borderDragStyle.GetForeground())
	}
	if collapseStyle.GetForeground() != lipgloss.Color(borderColor) {
		t.Fatalf("expected collapseStyle fg %v, got %v", borderColor, collapseStyle.GetForeground())
	}
}

func TestFloatPaneStyles(t *testing.T) {
	if floatBorderStyle.GetForeground() != lipgloss.Color(floatBorderColor) {
		t.Fatalf("expected floatBorderStyle fg %v, got %v", floatBorderColor, floatBorderStyle.GetForeground())
	}
	if floatTitleStyle.GetBackground() != lipgloss.Color(floatTitleBg) {
		t.Fatalf("expected floatTitleStyle bg %v, got %v", floatTitleBg, floatTitleStyle.GetBackground())
	}
	if floatTitleStyle.GetForeground() != lipgloss.Color(floatTitleFg) {
		t.Fatalf("expected floatTitleStyle fg %v, got %v", floatTitleFg, floatTitleStyle.GetForeground())
	}
	if !floatTitleStyle.GetBold() {
		t.Fatalf("expected floatTitleStyle bold")
	}
	if floatCloseStyle.GetForeground() != lipgloss.Color(floatCloseFg) {
		t.Fatalf("expected floatCloseStyle fg %v, got %v", floatCloseFg, floatCloseStyle.GetForeground())
	}
	if !floatCloseStyle.GetBold() {
		t.Fatalf("expected floatCloseStyle bold")
	}
	if floatBgStyle.GetBackground() != lipgloss.Color(floatBg) {
		t.Fatalf("expected floatBgStyle bg %v, got %v", floatBg, floatBgStyle.GetBackground())
	}
}

func TestCollapsibleStyles(t *testing.T) {
	if collapsibleStyle.GetForeground() != lipgloss.Color(gbLight1) {
		t.Fatalf("expected collapsibleStyle fg %v, got %v", gbLight1, collapsibleStyle.GetForeground())
	}
	if collapsibleStyle.GetBackground() != lipgloss.Color(gbDark1) {
		t.Fatalf("expected collapsibleStyle bg %v, got %v", gbDark1, collapsibleStyle.GetBackground())
	}
	if collapsibleBorderStyle.GetForeground() != lipgloss.Color(gbDark4) {
		t.Fatalf("expected collapsibleBorderStyle fg %v, got %v", gbDark4, collapsibleBorderStyle.GetForeground())
	}
}

func TestDropdownStyles(t *testing.T) {
	if dropdownButtonStyle.GetBackground() != lipgloss.Color(gbDark2) {
		t.Fatalf("expected dropdownButtonStyle bg %v, got %v", gbDark2, dropdownButtonStyle.GetBackground())
	}
	if dropdownButtonStyle.GetForeground() != lipgloss.Color(gbLight1) {
		t.Fatalf("expected dropdownButtonStyle fg %v, got %v", gbLight1, dropdownButtonStyle.GetForeground())
	}
	if dropdownItemStyle.GetBackground() != lipgloss.Color(gbDark0) {
		t.Fatalf("expected dropdownItemStyle bg %v, got %v", gbDark0, dropdownItemStyle.GetBackground())
	}
	if dropdownItemStyle.GetForeground() != lipgloss.Color(gbLight1) {
		t.Fatalf("expected dropdownItemStyle fg %v, got %v", gbLight1, dropdownItemStyle.GetForeground())
	}
	if dropdownItemHoverStyle.GetBackground() != lipgloss.Color(gbDark2) {
		t.Fatalf("expected dropdownItemHoverStyle bg %v, got %v", gbDark2, dropdownItemHoverStyle.GetBackground())
	}
	if dropdownItemHoverStyle.GetForeground() != lipgloss.Color(gbYellow) {
		t.Fatalf("expected dropdownItemHoverStyle fg %v, got %v", gbYellow, dropdownItemHoverStyle.GetForeground())
	}
	if dropdownItemSelectedStyle.GetBackground() != lipgloss.Color(gbDark2) {
		t.Fatalf("expected dropdownItemSelectedStyle bg %v, got %v", gbDark2, dropdownItemSelectedStyle.GetBackground())
	}
	if dropdownItemSelectedStyle.GetForeground() != lipgloss.Color(gbGreen) {
		t.Fatalf("expected dropdownItemSelectedStyle fg %v, got %v", gbGreen, dropdownItemSelectedStyle.GetForeground())
	}
	if !dropdownItemSelectedStyle.GetBold() {
		t.Fatalf("expected dropdownItemSelectedStyle bold")
	}
}

func TestGruvboxPalette(t *testing.T) {
	if gbDark0 != lipgloss.Color("#282828") {
		t.Fatalf("expected gbDark0, got %v", gbDark0)
	}
	if gbDark1 != lipgloss.Color("#3c3836") {
		t.Fatalf("expected gbDark1, got %v", gbDark1)
	}
	if gbDark2 != lipgloss.Color("#504945") {
		t.Fatalf("expected gbDark2, got %v", gbDark2)
	}
	if gbDark3 != lipgloss.Color("#665c54") {
		t.Fatalf("expected gbDark3, got %v", gbDark3)
	}
	if gbDark4 != lipgloss.Color("#7c6f64") {
		t.Fatalf("expected gbDark4, got %v", gbDark4)
	}
	if gbGray != lipgloss.Color("#928374") {
		t.Fatalf("expected gbGray, got %v", gbGray)
	}
	if gbLight1 != lipgloss.Color("#ebdbb2") {
		t.Fatalf("expected gbLight1, got %v", gbLight1)
	}
	if gbRed != lipgloss.Color("#fb4934") {
		t.Fatalf("expected gbRed, got %v", gbRed)
	}
	if gbGreen != lipgloss.Color("#b8bb26") {
		t.Fatalf("expected gbGreen, got %v", gbGreen)
	}
	if gbYellow != lipgloss.Color("#fabd2f") {
		t.Fatalf("expected gbYellow, got %v", gbYellow)
	}
	if gbBlue != lipgloss.Color("#83a598") {
		t.Fatalf("expected gbBlue, got %v", gbBlue)
	}
}

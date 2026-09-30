package e2e_test

import (
	"fmt"
	"strings"
	"testing"

	warp "github.com/starframe-dev/warp"
)

func TestRuntimeThemeUpdatesExistingTabsControlsAndOverlays(t *testing.T) {
	// Construct every component before changing the package-level styles.
	tabs := warp.NewTabGroup(warp.TabTop)
	tabs.NewTab("editor")
	input := warp.NewInput("Search")
	input.Focus()
	dropdown := warp.NewDropdownMenu("Palette", []warp.DropdownItem{
		{Label: "Ocean", Selected: true},
		{Label: "Forest"},
	})
	dropdown.Open = true
	modal := warp.NewModal("Theme applied", "Existing content", nil, nil)

	colors := warp.ThemeColors{
		Background:          "#102030",
		Surface:             "#213243",
		Raised:              "#324354",
		Border:              "#435465",
		BorderMuted:         "#546576",
		Text:                "#657687",
		TextMuted:           "#768798",
		TextStrong:          "#8798a9",
		Accent:              "#98a9ba",
		AccentMuted:         "#a9bacb",
		Error:               "#bacbdc",
		Success:             "#cbdced",
		Warning:             "#dcedfe",
		SelectionBackground: "#1a2b3c",
		SelectionForeground: "#f1e2d3",
	}
	warp.SetTheme(colors)
	defer warp.SetTheme(warp.ThemeColors{
		Background: "#282828", Surface: "#3c3836", Raised: "#504945",
		Border: "#7c6f64", BorderMuted: "#665c54", Text: "#ebdbb2",
		TextMuted: "#928374", TextStrong: "#ebdbb2", Accent: "#83a598",
		AccentMuted: "#83a598", Error: "#fb4934", Success: "#b8bb26",
		Warning: "#fabd2f", SelectionBackground: "#504945", SelectionForeground: "#ebdbb2",
	})

	tabView := warp.StripANSI(tabs.View(48, 8))
	if !strings.Contains(tabView, "editor") {
		t.Fatalf("existing tab did not render after applying theme: %q", tabView)
	}
	if got := fmt.Sprint(warp.BorderStyle().GetForeground()); got != colors.Surface {
		t.Fatalf("split border foreground = %q, want themed surface %q", got, colors.Surface)
	}

	inputView := warp.StripANSI(input.View(24, 3))
	if !strings.Contains(inputView, "Search") {
		t.Fatalf("existing focused input did not render after applying theme: %q", inputView)
	}

	dropdownView := warp.StripANSI(dropdown.View(24, 4))
	if !strings.Contains(dropdownView, "Palette") || !strings.Contains(dropdownView, "Ocean") || !strings.Contains(dropdownView, "Forest") {
		t.Fatalf("existing dropdown did not render all visible rows after applying theme: %q", dropdownView)
	}

	background := make([]string, 12)
	for i := range background {
		background[i] = strings.Repeat("background ", 8)
	}
	modalView := warp.StripANSI(strings.Join(modal.Overlay(background, 64, len(background)), "\n"))
	for _, content := range []string{"Theme applied", "Existing content", "background"} {
		if !strings.Contains(modalView, content) {
			t.Errorf("existing modal overlay does not contain %q after applying theme: %q", content, modalView)
		}
	}
}

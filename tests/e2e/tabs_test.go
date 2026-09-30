package e2e_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type tabE2EPanel struct {
	focused bool
	focuses int
	blurs   int
}

func (*tabE2EPanel) View(_, _ int) string   { return "tab content" }
func (*tabE2EPanel) Update(tea.Msg) tea.Cmd { return nil }
func (p *tabE2EPanel) Focus()               { p.focused = true; p.focuses++ }
func (p *tabE2EPanel) Blur()                { p.focused = false; p.blurs++ }
func (p *tabE2EPanel) Focused() bool        { return p.focused }

func waitForTabState(t *testing.T, condition func() bool, description string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}

func TestTabsKeyboardMouseCreationCloseAndFocusRetention(t *testing.T) {
	const width, height = 72, 18
	app := warp.New()
	group, ok := app.Root().(*warp.TabGroup)
	if !ok {
		t.Fatal("Warp root is not a TabGroup")
	}

	mainTab := app.ActiveTab()
	mainPanel := &tabE2EPanel{}
	mainTab.SetRootPanel(mainPanel)
	mainTab.FocusFirst()
	secondTab := app.NewTab("second")
	secondPanel := &tabE2EPanel{}
	secondTab.SetRootPanel(secondPanel)
	secondTab.FocusFirst()
	app.PrevTab()
	if app.ActiveTab() != mainTab || !mainPanel.Focused() {
		t.Fatal("initial tab focus was not restored")
	}

	var output bytes.Buffer
	program := tea.NewProgram(app,
		tea.WithInput(strings.NewReader("")),
		tea.WithOutput(&output),
	)
	done := make(chan error, 1)
	go func() {
		_, err := program.Run()
		done <- err
	}()
	defer func() {
		program.Quit()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			program.Kill()
			t.Error("Bubble Tea program did not shut down")
		}
	}()

	program.Send(tea.WindowSizeMsg{Width: width, Height: height})
	waitForTabState(t, func() bool { return app.Width() == width && app.Height() == height }, "terminal window size")

	// Ctrl+Tab and Ctrl+Shift+Tab navigate the tabs in either direction.
	program.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+tab")})
	waitForTabState(t, func() bool { return app.ActiveTab() == secondTab }, "Ctrl+Tab to activate second tab")
	if mainPanel.Focused() || !secondPanel.Focused() {
		t.Fatal("switching tabs did not suspend old focus and focus the new tab")
	}
	program.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+shift+tab")})
	waitForTabState(t, func() bool { return app.ActiveTab() == mainTab }, "Ctrl+Shift+Tab to return to main tab")
	if !mainPanel.Focused() || secondPanel.Focused() {
		t.Fatal("returning to main did not restore its retained focus")
	}

	// Click the second tab's semantic tab-bar target using its rendered bounds.
	group.View(width, height)
	var secondElement warp.Element
	found := false
	for _, element := range group.Elements(width, height) {
		if element.Role == "tab" && element.Name == "second" && element.Action == "activate-tab" {
			secondElement, found = element, true
			break
		}
	}
	if !found {
		t.Fatal("second tab has no semantic mouse target")
	}
	x, y := secondElement.Bounds.Center()
	program.Send(tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	waitForTabState(t, func() bool { return app.ActiveTab() == secondTab }, "mouse click to activate second tab")
	if !secondPanel.Focused() || mainPanel.Focused() {
		t.Fatal("mouse tab switch did not restore the second tab's focus")
	}

	// The keyboard shortcut creates and activates a fresh tab; Ctrl+W closes it.
	program.Send(tea.KeyMsg{Type: tea.KeyCtrlT})
	waitForTabState(t, func() bool { return app.ActiveTab() != secondTab }, "Ctrl+T to create and activate a tab")
	createdTab := app.ActiveTab()
	if createdTab == mainTab {
		t.Fatal("Ctrl+T did not create a distinct tab")
	}
	program.Send(tea.KeyMsg{Type: tea.KeyCtrlW})
	waitForTabState(t, func() bool { return app.ActiveTab() == secondTab }, "Ctrl+W to close the created tab")
	if !secondPanel.Focused() {
		t.Fatal("closing the active tab did not restore focus in the selected tab")
	}

	// Exercise mouse creation and closing through the visible + and × controls.
	group.View(width, height)
	var newButton warp.Element
	found = false
	for _, element := range group.Elements(width, height) {
		if element.Role == "button" && element.Action == "new-tab" {
			newButton, found = element, true
			break
		}
	}
	if !found {
		t.Fatal("tab bar has no new-tab mouse target")
	}
	x, y = newButton.Bounds.Center()
	program.Send(tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	waitForTabState(t, func() bool { return app.ActiveTab() != secondTab }, "mouse click to create a tab")
	mouseCreatedTab := app.ActiveTab()
	group.View(width, height)
	var closeButton warp.Element
	found = false
	for _, element := range group.Elements(width, height) {
		for _, child := range element.Children {
			if child.Role == "button" && child.Action == "close-tab" {
				closeButton, found = child, true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("active tab has no close mouse target")
	}
	x, y = closeButton.Bounds.Center()
	program.Send(tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	waitForTabState(t, func() bool { return app.ActiveTab() == secondTab }, "mouse click to close the active tab")
	if mouseCreatedTab == app.ActiveTab() || !secondPanel.Focused() {
		t.Fatal("mouse close did not select the remaining tab and restore its focus")
	}
}

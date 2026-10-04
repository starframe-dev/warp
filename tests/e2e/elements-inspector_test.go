package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

func TestHTTPInspectorTracksTabsAndInteractiveElements(t *testing.T) {
	const width, height = 64, 18

	app := warp.New()
	if err := app.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("start HTTP inspector: %v", err)
	}
	defer func() { _ = app.Close() }()

	menu := warp.NewDropdownMenu("Color", []warp.DropdownItem{
		{Label: "Red", Selected: true},
		{Label: "Green"},
	})
	mainTab := app.ActiveTab()
	mainTab.SetRootPanel(menu)

	settingsTab := app.NewTab("settings")
	input := warp.NewInput("Name")
	input.SetValue("Ada")
	settingsTab.SetRootPanel(input)
	settingsTab.FocusFirst()
	app.PrevTab()

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

	baseURL := "http://" + app.HTTPAddr()
	waitForInspector(t, func() bool {
		response, body, err := inspectorGet(baseURL + "/healthz")
		return err == nil && response.StatusCode == http.StatusOK && strings.TrimSpace(string(body)) == "ok"
	}, "healthy inspector endpoint")

	program.Send(tea.WindowSizeMsg{Width: width, Height: height})
	waitForInspector(t, func() bool { return app.Width() == width && app.Height() == height }, "terminal dimensions")
	program.Send(tea.MouseMsg{X: 2, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease}) // harmless release
	waitForInspector(t, func() bool {
		elements, err := inspectorElements(baseURL)
		if err != nil {
			return false
		}
		_, tabOK := findInspectorElement(elements, "tab", "main", "activate-tab")
		_, comboOK := findInspectorElement(elements, "combobox", "Color", "toggle")
		return tabOK && comboOK
	}, "initial tab and collapsed dropdown snapshot")

	elements := mustInspectorElements(t, baseURL)
	mainSemantic, ok := findInspectorElement(elements, "tab", "main", "activate-tab")
	if !ok {
		t.Fatal("main tab is missing from semantic snapshot")
	}
	assertInspectorBounds(t, mainSemantic.Bounds, width, height)
	combo, ok := findInspectorElement(elements, "combobox", "Color", "toggle")
	if !ok {
		t.Fatal("collapsed dropdown combobox is missing")
	}
	assertInspectorBounds(t, combo.Bounds, width, height)
	if _, ok := findInspectorElement(elements, "option", "Red", "select"); ok {
		t.Fatal("closed dropdown unexpectedly exposes option rows")
	}

	// Open the dropdown using the cell-coordinate target published by the inspector.
	x, y := combo.Bounds.Center()
	program.Send(tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	waitForInspector(t, func() bool {
		elements, err := inspectorElements(baseURL)
		if err != nil {
			return false
		}
		_, redOK := findInspectorElement(elements, "option", "Red", "select")
		_, greenOK := findInspectorElement(elements, "option", "Green", "select")
		return redOK && greenOK
	}, "open dropdown options")
	elements = mustInspectorElements(t, baseURL)
	red, ok := findInspectorElement(elements, "option", "Red", "select")
	if !ok {
		t.Fatal("Red option is missing while dropdown is open")
	}
	assertInspectorBounds(t, red.Bounds, width, height)
	green, ok := findInspectorElement(elements, "option", "Green", "select")
	if !ok {
		t.Fatal("Green option is missing while dropdown is open")
	}
	assertInspectorBounds(t, green.Bounds, width, height)
	if red.Bounds.Y == combo.Bounds.Y || green.Bounds.Y == red.Bounds.Y {
		t.Fatalf("dropdown option cell bounds did not reflect distinct visible rows: button=%+v red=%+v green=%+v", combo.Bounds, red.Bounds, green.Bounds)
	}

	// Selecting Green closes the menu and updates the semantic tree.
	x, y = green.Bounds.Center()
	program.Send(tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	waitForInspector(t, func() bool {
		elements, err := inspectorElements(baseURL)
		if err != nil {
			return false
		}
		_, comboOK := findInspectorElement(elements, "combobox", "Color", "toggle")
		_, optionOK := findInspectorElement(elements, "option", "Green", "select")
		return comboOK && !optionOK
	}, "dropdown selection and collapsed snapshot")

	// Activate the other tab from its semantic tab hit target.
	elements = mustInspectorElements(t, baseURL)
	settingsSemantic, ok := findInspectorElement(elements, "tab", "settings", "activate-tab")
	if !ok {
		t.Fatal("settings tab is missing from semantic snapshot")
	}
	assertInspectorBounds(t, settingsSemantic.Bounds, width, height)
	x, y = settingsSemantic.Bounds.Center()
	program.Send(tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	waitForInspector(t, func() bool {
		elements, err := inspectorElements(baseURL)
		if err != nil {
			return false
		}
		_, inputOK := findInspectorElement(elements, "textbox", "Name", "focus")
		_, comboOK := findInspectorElement(elements, "combobox", "Color", "toggle")
		return inputOK && !comboOK
	}, "settings tab and textbox snapshot")
	elements = mustInspectorElements(t, baseURL)
	textbox, ok := findInspectorElement(elements, "textbox", "Name", "focus")
	if !ok {
		t.Fatal("Name textbox is missing from settings tab snapshot")
	}
	assertInspectorBounds(t, textbox.Bounds, width, height)

	// Editing is covered by the input integration tests. Here the inspector
	// only needs to remain coherent while those events pass through the live
	// Bubble Tea program before switching back to the first tab.
	program.Send(tea.KeyMsg{Type: tea.KeyEnd})
	program.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")})
	program.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+shift+tab")})
	waitForInspector(t, func() bool {
		elements, err := inspectorElements(baseURL)
		if err != nil {
			return false
		}
		_, comboOK := findInspectorElement(elements, "combobox", "Color", "toggle")
		_, inputOK := findInspectorElement(elements, "textbox", "Name", "focus")
		return comboOK && !inputOK
	}, "restored main tab snapshot")
}

func waitForInspector(t *testing.T, condition func() bool, description string) {
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

func inspectorGet(url string) (*http.Response, []byte, error) {
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	return response, body, err
}

func inspectorElements(baseURL string) ([]warp.Element, error) {
	response, body, err := inspectorGet(baseURL + "/elements")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("/elements returned HTTP %d: %s", response.StatusCode, body)
	}
	var elements []warp.Element
	if err := json.Unmarshal(body, &elements); err != nil {
		return nil, err
	}
	return elements, nil
}

func mustInspectorElements(t *testing.T, baseURL string) []warp.Element {
	t.Helper()
	elements, err := inspectorElements(baseURL)
	if err != nil {
		t.Fatalf("read semantic snapshot: %v", err)
	}
	return elements
}

func findInspectorElement(elements []warp.Element, role, name, action string) (warp.Element, bool) {
	for _, element := range elements {
		if (role == "" || element.Role == role) && (name == "" || element.Name == name) && (action == "" || element.Action == action) {
			return element, true
		}
		if match, ok := findInspectorElement(element.Children, role, name, action); ok {
			return match, true
		}
	}
	return warp.Element{}, false
}

func assertInspectorBounds(t *testing.T, bounds warp.Bounds, width, height int) {
	t.Helper()
	if bounds.X < 0 || bounds.Y < 0 || bounds.W <= 0 || bounds.H <= 0 || bounds.X+bounds.W > width || bounds.Y+bounds.H > height {
		t.Fatalf("semantic bounds %+v are outside %dx%d cell-coordinate viewport", bounds, width, height)
	}
}

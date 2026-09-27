package warp

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type semanticElementPanel struct {
	BasePanel
	content       string
	elements      []Element
	requestedSize Bounds
}

func (p *semanticElementPanel) View(width, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(p.content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	for i := range lines {
		lines[i] = padVisualLine(lines[i], max(0, width))
	}
	return strings.Join(lines, "\n")
}

func (p *semanticElementPanel) Elements(width, height int) []Element {
	p.requestedSize = Bounds{W: width, H: height}
	return p.elements
}

func newSemanticElementPanel() (*semanticElementPanel, []Element) {
	elements := []Element{
		{Role: "text", Name: "top", Bounds: Bounds{X: 1, Y: 0, W: 2, H: 1}},
		{Role: "text", Name: "left-clipped", Bounds: Bounds{X: -1, Y: 1, W: 3, H: 1}},
		{
			Role:   "group",
			Name:   "group",
			Bounds: Bounds{X: 1, Y: 2, W: 5, H: 2},
			Children: []Element{{
				Role:   "button",
				Name:   "child-row-3",
				Bounds: Bounds{X: 2, Y: 3, W: 3, H: 1},
			}},
		},
		{Role: "text", Name: "partial-bottom", Bounds: Bounds{X: 2, Y: 3, W: 2, H: 2}},
		{Role: "text", Name: "bottom", Bounds: Bounds{X: 0, Y: 4, W: 3, H: 1}},
		{Role: "text", Name: "offscreen", Bounds: Bounds{X: 0, Y: 6, W: 1, H: 1}},
	}
	return &semanticElementPanel{
		content:  "top\nrow1\nrow2\nrow3\nrow4\nrow5\nrow6",
		elements: elements,
	}, cloneElements(elements)
}

func TestElementWrappersForwardAndClipSemanticBounds(t *testing.T) {
	panel, original := newSemanticElementPanel()

	if got := collectElements(panel, 8, 3); !reflect.DeepEqual(got, original) {
		t.Fatalf("direct Elements = %+v, want %+v", got, original)
	}

	selectable := NewSelectable(panel)
	selectableElements := collectElements(selectable, 8, 3)
	if !reflect.DeepEqual(selectableElements, original) {
		t.Fatalf("Selectable Elements = %+v, want unchanged %+v", selectableElements, original)
	}
	selectableElements[0].Bounds.X++
	if !reflect.DeepEqual(panel.elements, original) {
		t.Fatal("Selectable.Elements exposed provider-owned element storage")
	}
	if panel.requestedSize != (Bounds{W: 8, H: 3}) {
		t.Fatalf("Selectable requested size %+v, want 8x3", panel.requestedSize)
	}

	panel, original = newSemanticElementPanel()
	scrollable := &Scrollable{Content: NewSelectable(panel), Offset: 1}
	got := collectElements(scrollable, 8, 3)
	want := []Element{
		{Role: "text", Name: "left-clipped", Bounds: Bounds{X: 0, Y: 0, W: 2, H: 1}},
		{
			Role:   "group",
			Name:   "group",
			Bounds: Bounds{X: 1, Y: 1, W: 5, H: 2},
			Children: []Element{{
				Role:   "button",
				Name:   "child-row-3",
				Bounds: Bounds{X: 2, Y: 2, W: 3, H: 1},
			}},
		},
		{Role: "text", Name: "partial-bottom", Bounds: Bounds{X: 2, Y: 2, W: 2, H: 1}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Scrollable Elements = %+v, want %+v", got, want)
	}
	if panel.requestedSize != (Bounds{W: 8, H: 4}) {
		t.Fatalf("Scrollable content requested size %+v, want 8x4", panel.requestedSize)
	}
	if !reflect.DeepEqual(panel.elements, original) {
		t.Fatal("Scrollable.Elements mutated provider-owned element bounds")
	}
}

func TestCollapsibleElementsFollowVisibleContent(t *testing.T) {
	panel, _ := newSemanticElementPanel()
	collapsible := NewCollapsible("Section", NewSelectable(panel))

	got := collectElements(collapsible, 8, 4)
	want := []Element{
		{Role: "button", Name: "Section", Action: "toggle-collapse", Bounds: Bounds{W: 8, H: 1}},
		{Role: "text", Name: "top", Bounds: Bounds{X: 1, Y: 1, W: 2, H: 1}},
		{Role: "text", Name: "left-clipped", Bounds: Bounds{X: 0, Y: 2, W: 2, H: 1}},
		{Role: "group", Name: "group", Bounds: Bounds{X: 1, Y: 3, W: 5, H: 1}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expanded Collapsible Elements = %+v, want %+v", got, want)
	}
	if panel.requestedSize != (Bounds{W: 8, H: 3}) {
		t.Fatalf("expanded content requested size %+v, want 8x3", panel.requestedSize)
	}

	collapsible.Collapsed = true
	collapsed := collectElements(collapsible, 8, 4)
	if len(collapsed) != 1 || collapsed[0].Action != "toggle-collapse" {
		t.Fatalf("collapsed Collapsible semantic controls = %+v, want title toggle only", collapsed)
	}
}

func TestNestedWrappersPreserveElementsInInspectorCoordinates(t *testing.T) {
	panel, _ := newSemanticElementPanel()
	wrapped := NewCollapsible("Section", &Scrollable{
		Content: NewSelectable(panel),
		Offset:  1,
	})
	warp := New()
	warp.SetRoot(wrapped)
	if err := warp.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}
	defer func() { _ = warp.CloseHTTP() }()
	warp.Update(tea.WindowSizeMsg{Width: 8, Height: 3})

	view := strings.Split(ansi.Strip(warp.View()), "\n")
	if len(view) != 3 || !strings.Contains(view[1], "row1") || !strings.Contains(view[2], "row2") {
		t.Fatalf("nested wrapper View = %q, want title, row1, row2", view)
	}

	recorder := httptest.NewRecorder()
	warp.handleElements(recorder, httptest.NewRequest("GET", "/elements", nil))
	var got []Element
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode /elements snapshot: %v", err)
	}
	want := []Element{
		{Role: "button", Name: "Section", Action: "toggle-collapse", Bounds: Bounds{W: 8, H: 1}},
		{Role: "text", Name: "left-clipped", Bounds: Bounds{X: 0, Y: 1, W: 2, H: 1}},
		{Role: "group", Name: "group", Bounds: Bounds{X: 1, Y: 2, W: 5, H: 1}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("nested /elements = %+v, want %+v", got, want)
	}
}

func TestCollectElementsTypedNilPanel(t *testing.T) {
	var panel *semanticElementPanel
	if got := collectElements(panel, 8, 3); got != nil {
		t.Fatalf("collectElements(typed nil) = %+v, want nil", got)
	}
}

func TestTabElementsDoesNotMutateProviderOwnedBounds(t *testing.T) {
	panel, original := newSemanticElementPanel()
	tab := NewTab("semantic-provider")
	left := tab.RootPanel()
	tab.SplitVertical(left, 0.5, panel)

	first := tab.Elements(20, 8)
	second := tab.Elements(20, 8)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated Tab.Elements drifted: first=%+v second=%+v", first, second)
	}
	if !reflect.DeepEqual(panel.elements, original) {
		t.Fatalf("Tab.Elements mutated provider-owned bounds: got=%+v want=%+v", panel.elements, original)
	}
}

func TestTabElementsIncludeFloatsTopmostFirst(t *testing.T) {
	root := &semanticElementPanel{elements: []Element{{Role: "root", Name: "root", Bounds: Bounds{W: 1, H: 1}}}}
	lower := &semanticElementPanel{elements: []Element{{Role: "button", Name: "lower", Bounds: Bounds{W: 1, H: 1}}}}
	upper := &semanticElementPanel{elements: []Element{{Role: "button", Name: "upper", Bounds: Bounds{W: 1, H: 1}}}}

	tab := NewTab("float-elements")
	tab.SetRootPanel(root)
	tab.Float(lower, 2, 1, 10, 4)
	tab.floats[0].Title = "Lower"
	tab.Float(upper, 4, 2, 10, 4)
	tab.floats[1].Title = "Upper"

	elements := tab.Elements(30, 10)
	if len(elements) != 7 {
		t.Fatalf("Tab.Elements returned %d elements, want 7: %+v", len(elements), elements)
	}
	want := []struct {
		role, name, action string
		bounds             Bounds
	}{
		{"button", "Close Upper", "close-float", Bounds{X: 12, Y: 2, W: 1, H: 1}},
		{"titlebar", "Upper", "move-float", Bounds{X: 5, Y: 2, W: 7, H: 1}},
		{"button", "upper", "", Bounds{X: 5, Y: 3, W: 1, H: 1}},
		{"button", "Close Lower", "close-float", Bounds{X: 10, Y: 1, W: 1, H: 1}},
		{"titlebar", "Lower", "move-float", Bounds{X: 3, Y: 1, W: 7, H: 1}},
		{"button", "lower", "", Bounds{X: 3, Y: 2, W: 1, H: 1}},
		{"root", "root", "", Bounds{W: 1, H: 1}},
	}
	for i, expected := range want {
		got := elements[i]
		if got.Role != expected.role || got.Name != expected.name || got.Action != expected.action || got.Bounds != expected.bounds {
			t.Fatalf("element[%d]=%+v, want role=%q name=%q action=%q bounds=%+v", i, got, expected.role, expected.name, expected.action, expected.bounds)
		}
	}
}

func TestTabElementsExposeSplitCollapseControl(t *testing.T) {
	tab := NewTab("split-control")
	left := &semanticElementPanel{}
	right := &semanticElementPanel{}
	tab.SetRootPanel(left)
	tab.SplitVertical(left, 0.5, right)
	tab.SetSplitCollapse(left, 2, nil)

	const width, height = 20, 8
	layout := newLayout(tab.root, layoutRect{w: width, h: height})
	borders := collectLayoutBorders(layout)
	if len(borders) == 0 {
		t.Fatal("test setup did not create a split border")
	}
	expected := Bounds{X: borders[0].X, Y: borders[0].Y + 2, W: 1, H: 1}

	elements := tab.Elements(width, height)
	control, ok := FindElement(elements, "button", "Toggle split", "toggle-collapse")
	if !ok {
		t.Fatalf("split collapse semantic control missing: %+v", elements)
	}
	if control.Bounds != expected {
		t.Fatalf("split collapse bounds=%+v, want %+v", control.Bounds, expected)
	}
}

func TestCollapsibleTitleMouseToggleIsSelfContained(t *testing.T) {
	collapsible := NewCollapsible("Section", nil)
	if collapsible.Collapsed {
		t.Fatal("new Collapsible unexpectedly collapsed")
	}
	collapsible.Update(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, Y: 0})
	if !collapsible.Collapsed {
		t.Fatal("title click did not collapse a standalone/wrapped Collapsible")
	}
}

func TestTabRootElementsDoNotExposeProviderStorage(t *testing.T) {
	panel, original := newSemanticElementPanel()
	tab := NewTab("root-provider")
	tab.SetRootPanel(panel)

	got := tab.Elements(8, 4)
	if len(got) == 0 {
		t.Fatal("Tab.Elements returned no root elements")
	}
	got[0].Bounds.X += 100
	if !reflect.DeepEqual(panel.elements, original) {
		t.Fatalf("Tab.Elements exposed provider storage: got=%+v want=%+v", panel.elements, original)
	}
}

func TestCollapsibleElementsDoNotMutateProviderStorage(t *testing.T) {
	panel, original := newSemanticElementPanel()
	collapsible := NewCollapsible("Section", panel)

	got := collapsible.Elements(8, 4)
	if len(got) == 0 {
		t.Fatal("Collapsible.Elements returned no visible elements")
	}
	if !reflect.DeepEqual(panel.elements, original) {
		t.Fatalf("Collapsible.Elements mutated provider storage: got=%+v want=%+v", panel.elements, original)
	}
}

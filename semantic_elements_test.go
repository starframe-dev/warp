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
	if got := collectElements(collapsible, 8, 4); len(got) != 0 {
		t.Fatalf("collapsed Collapsible exposed hidden elements: %+v", got)
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

package warp

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type viewportRequest struct {
	width  int
	height int
	offset int
}

type viewportExtentPanel struct {
	rows            []string
	viewRequests    []int
	viewAtRequests  []viewportRequest
	elementRequests []int
	elementsAtCalls []viewportRequest
}

func (p *viewportExtentPanel) View(_, height int) string {
	p.viewRequests = append(p.viewRequests, height)
	return strings.Join(p.rows[:min(len(p.rows), max(0, height))], "\n")
}

func (*viewportExtentPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *viewportExtentPanel) ContentHeight(int) (int, bool) {
	return len(p.rows), true
}

func (p *viewportExtentPanel) ViewAt(width, height, offset int) string {
	p.viewAtRequests = append(p.viewAtRequests, viewportRequest{width: width, height: height, offset: offset})
	start := min(len(p.rows), max(0, offset))
	end := min(len(p.rows), saturatingAddNonNegative(start, max(0, height)))
	return strings.Join(p.rows[start:end], "\n")
}

func (p *viewportExtentPanel) Elements(_, height int) []Element {
	p.elementRequests = append(p.elementRequests, height)
	return viewportRows(p.rows, 0, min(len(p.rows), max(0, height)))
}

func (p *viewportExtentPanel) ElementsAt(width, height, offset int) []Element {
	p.elementsAtCalls = append(p.elementsAtCalls, viewportRequest{width: width, height: height, offset: offset})
	start := min(len(p.rows), max(0, offset))
	end := min(len(p.rows), saturatingAddNonNegative(start, max(0, height)))
	return viewportRows(p.rows, start, end)
}

func viewportRows(rows []string, start, end int) []Element {
	elements := make([]Element, max(0, end-start))
	for i := range elements {
		index := start + i
		elements[i] = Element{
			Role:   "text",
			Name:   rows[index],
			Bounds: Bounds{Y: index, W: len(rows[index]), H: 1},
		}
	}
	return elements
}

func TestScrollableUsesViewportInterfacesForKnownExtent(t *testing.T) {
	panel := &viewportExtentPanel{rows: []string{
		"row0", "row1", "row2", "row3", "row4", "row5", "row6", "row7", "row8", "row9",
	}}
	scrollable := NewScrollable(panel)
	scrollable.Offset = 100

	view := strings.Split(scrollable.View(8, 3), "\n")
	if got, want := view, []string{"row7    ", "row8    ", "row9    "}; !equalStrings(got, want) {
		t.Fatalf("viewport View = %q, want %q", got, want)
	}
	if len(panel.viewRequests) != 0 {
		t.Fatalf("legacy View requests = %v, want none", panel.viewRequests)
	}
	if got, want := panel.viewAtRequests, []viewportRequest{{width: 8, height: 3, offset: 7}}; !equalViewportRequests(got, want) {
		t.Fatalf("ViewAt requests = %+v, want %+v", got, want)
	}
	if scrollable.Offset != 7 {
		t.Fatalf("stored offset = %d, want clamped 7", scrollable.Offset)
	}

	elements := scrollable.Elements(8, 3)
	wantElements := []Element{
		{Role: "text", Name: "row7", Bounds: Bounds{Y: 0, W: 4, H: 1}},
		{Role: "text", Name: "row8", Bounds: Bounds{Y: 1, W: 4, H: 1}},
		{Role: "text", Name: "row9", Bounds: Bounds{Y: 2, W: 4, H: 1}},
	}
	if len(elements) != len(wantElements) {
		t.Fatalf("viewport Elements count = %d, want %d", len(elements), len(wantElements))
	}
	for i := range wantElements {
		got, want := elements[i], wantElements[i]
		if got.Role != want.Role || got.Name != want.Name || got.Bounds != want.Bounds || len(got.Children) != 0 {
			t.Fatalf("viewport Element[%d] = %+v, want %+v", i, got, want)
		}
	}
	if len(panel.elementRequests) != 0 {
		t.Fatalf("legacy Elements requests = %v, want none", panel.elementRequests)
	}
	if got, want := panel.elementsAtCalls, []viewportRequest{{width: 8, height: 3, offset: 7}}; !equalViewportRequests(got, want) {
		t.Fatalf("ElementsAt requests = %+v, want %+v", got, want)
	}
}

func equalViewportRequests(got, want []viewportRequest) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestScrollableViewportInterfacesPreserveContentCoordinates(t *testing.T) {
	rows := make([]string, 128)
	for i := range rows {
		rows[i] = fmt.Sprintf("row-%d", i)
	}
	panel := &viewportExtentPanel{rows: rows}
	scrollable := NewScrollable(panel)
	scrollable.Offset = 120

	elements := scrollable.Elements(12, 5)
	if len(elements) != 5 {
		t.Fatalf("viewport element count = %d, want 5", len(elements))
	}
	for i, element := range elements {
		if element.Name != fmt.Sprintf("row-%d", 120+i) || element.Bounds.Y != i {
			t.Fatalf("viewport element[%d] = %+v, want row-%d at y=%d", i, element, 120+i, i)
		}
	}
}

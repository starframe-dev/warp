package warp

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type knownExtentPanel struct {
	heightAtWidth  func(int) (int, bool)
	padToHeight    bool
	viewRequests   []int
	elementRequest []int
	viewCalls      int
}

func (p *knownExtentPanel) View(width, height int) string {
	p.viewCalls++
	p.viewRequests = append(p.viewRequests, height)
	contentHeight, known := p.heightAtWidth(width)
	if !known {
		contentHeight = 4
	}
	lineCount := min(max(0, height), max(0, contentHeight))
	lines := make([]string, lineCount)
	for i := range lines {
		lines[i] = rowName(i)
	}
	if p.padToHeight {
		for len(lines) < height {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

func (p *knownExtentPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *knownExtentPanel) ContentHeight(width int) (int, bool) {
	return p.heightAtWidth(width)
}

func (p *knownExtentPanel) Elements(width, height int) []Element {
	p.elementRequest = append(p.elementRequest, height)
	contentHeight, known := p.heightAtWidth(width)
	if !known {
		contentHeight = 4
	}
	return rowElements(min(max(0, height), max(0, contentHeight)))
}

type naturalHeightPanel struct {
	rows         []string
	viewRequests []int
}

func (p *naturalHeightPanel) View(_, height int) string {
	p.viewRequests = append(p.viewRequests, height)
	return strings.Join(p.rows[:min(len(p.rows), max(0, height))], "\n")
}

func (*naturalHeightPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *naturalHeightPanel) Elements(_, height int) []Element {
	return rowElements(min(len(p.rows), max(0, height)))
}

type paddedUnknownPanel struct {
	viewRequests    []int
	elementRequests []int
}

func (p *paddedUnknownPanel) View(_, height int) string {
	p.viewRequests = append(p.viewRequests, height)
	if height <= 0 {
		return ""
	}
	return strings.Repeat("\n", height-1)
}

func (*paddedUnknownPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *paddedUnknownPanel) Elements(_, height int) []Element {
	p.elementRequests = append(p.elementRequests, height)
	return rowElements(min(4, max(0, height)))
}

type largeExtentPanel struct {
	requestedView     int
	requestedElements int
}

func (p *largeExtentPanel) View(_, height int) string {
	p.requestedView = height
	return "tail"
}

func (*largeExtentPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *largeExtentPanel) ContentHeight(int) (int, bool) {
	return math.MaxInt, true
}

func (p *largeExtentPanel) Elements(_, height int) []Element {
	p.requestedElements = height
	return nil
}

func TestScrollableKnownExtentClampsOverscrollForViewAndElements(t *testing.T) {
	content := &knownExtentPanel{
		heightAtWidth: func(int) (int, bool) { return 4, true },
		padToHeight:   true,
	}
	scrollable := NewScrollable(content)
	scrollable.Offset = 100

	gotView := scrollable.View(5, 3)
	if gotView != "row1 \nrow2 \nrow3 " {
		t.Fatalf("View = %q, want rows 1-3", gotView)
	}
	if scrollable.Offset != 1 {
		t.Fatalf("View stored offset %d, want 1", scrollable.Offset)
	}
	if got := content.viewRequests[len(content.viewRequests)-1]; got != 4 {
		t.Fatalf("Content.View requested %d rows, want effective offset + viewport = 4", got)
	}

	scrollable.Offset = 100
	gotElements := scrollable.Elements(8, 3)
	wantElements := rowElementsAt(1, 3)
	if !reflect.DeepEqual(gotElements, wantElements) {
		t.Fatalf("Elements = %+v, want %+v", gotElements, wantElements)
	}
	if scrollable.Offset != 100 {
		t.Fatalf("Elements mutated Offset to %d, want 100", scrollable.Offset)
	}
	if got := content.elementRequest[len(content.elementRequest)-1]; got != 4 {
		t.Fatalf("Content.Elements requested %d rows, want effective offset + viewport = 4", got)
	}
}

func TestScrollableUpdateThenElementsBeforeViewUsesEffectiveOffset(t *testing.T) {
	content := &knownExtentPanel{
		heightAtWidth: func(int) (int, bool) { return 4, true },
		padToHeight:   true,
	}
	scrollable := NewScrollable(content)
	warp := New()
	warp.SetRoot(scrollable)
	if err := warp.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}
	defer warp.CloseHTTP()

	warp.Update(tea.WindowSizeMsg{Width: 8, Height: 3})
	warp.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if scrollable.Offset != 1 {
		t.Fatalf("Offset after PgDown = %d, want clamped 1", scrollable.Offset)
	}
	if content.viewCalls != 0 {
		t.Fatalf("explicit extent caused %d View calls before rendering", content.viewCalls)
	}

	gotElements := snapshotElements(t, warp)
	wantElements := rowElementsAt(1, 3)
	if !reflect.DeepEqual(gotElements, wantElements) {
		t.Fatalf("/elements before View = %+v, want %+v", gotElements, wantElements)
	}
	if content.viewCalls != 0 {
		t.Fatalf("/elements rendered content %d times despite explicit extent", content.viewCalls)
	}

	gotView := strings.Split(ansi.Strip(warp.View()), "\n")
	if !reflect.DeepEqual(gotView, []string{"row1    ", "row2    ", "row3    "}) {
		t.Fatalf("View after snapshot = %q, want rows 1-3", gotView)
	}
	if scrollable.Offset != 1 {
		t.Fatalf("Offset after View = %d, want 1", scrollable.Offset)
	}
}

func TestScrollableContentShrinkClampsElementsAndView(t *testing.T) {
	contentHeight := 100
	content := &knownExtentPanel{
		heightAtWidth: func(int) (int, bool) { return contentHeight, true },
		padToHeight:   true,
	}
	scrollable := NewScrollable(content)
	scrollable.Offset = 80
	if got := scrollable.effectiveOffset(10, 3).offset; got != 80 {
		t.Fatalf("initial effective offset = %d, want 80", got)
	}

	contentHeight = 4
	if got, want := scrollable.Elements(10, 3), rowElementsAt(1, 3); !reflect.DeepEqual(got, want) {
		t.Fatalf("Elements after shrink = %+v, want %+v", got, want)
	}
	if scrollable.Offset != 80 {
		t.Fatalf("Elements changed stored Offset to %d, want 80", scrollable.Offset)
	}

	if got := strings.Split(ansi.Strip(scrollable.View(10, 3)), "\n"); !reflect.DeepEqual(got, []string{"row1      ", "row2      ", "row3      "}) {
		t.Fatalf("View after shrink = %q, want rows 1-3", got)
	}
	if scrollable.Offset != 1 {
		t.Fatalf("View after shrink stored offset %d, want 1", scrollable.Offset)
	}
}

func TestScrollableWidthDependentExtentAndResize(t *testing.T) {
	content := &knownExtentPanel{
		heightAtWidth: func(width int) (int, bool) {
			if width >= 20 {
				return 5, true
			}
			return 12, true
		},
		padToHeight: true,
	}
	scrollable := NewScrollable(content)
	scrollable.Offset = 100

	scrollable.Update(ResizeMsg{Width: 25, Height: 3})
	if scrollable.Offset != 2 {
		t.Fatalf("wide ResizeMsg clamped offset to %d, want 2", scrollable.Offset)
	}
	scrollable.Offset = 100
	scrollable.Update(ResizeMsg{Width: 10, Height: 3})
	if scrollable.Offset != 9 {
		t.Fatalf("narrow ResizeMsg clamped offset to %d, want 9", scrollable.Offset)
	}

	scrollable.Offset = 100
	if got := scrollable.Elements(25, 3); !reflect.DeepEqual(got, rowElementsAt(2, 3)) {
		t.Fatalf("wide Elements = %+v, want rows 2-4", got)
	}
	if got := scrollable.Elements(10, 3); !reflect.DeepEqual(got, rowElementsAt(9, 3)) {
		t.Fatalf("narrow Elements = %+v, want rows 9-11", got)
	}
	if scrollable.Offset != 100 {
		t.Fatalf("Elements changed stored Offset to %d, want 100", scrollable.Offset)
	}

	selectable := NewSelectable(content)
	if height, known := selectable.ContentHeight(25); !known || height != 5 {
		t.Fatalf("wide Selectable extent = (%d, %t), want (5, true)", height, known)
	}
	if height, known := selectable.ContentHeight(10); !known || height != 12 {
		t.Fatalf("narrow Selectable extent = (%d, %t), want (12, true)", height, known)
	}
	collapsible := NewCollapsible("section", content)
	if height, known := collapsible.ContentHeight(25); !known || height != 6 {
		t.Fatalf("wide Collapsible extent = (%d, %t), want (6, true)", height, known)
	}
	if height, known := collapsible.ContentHeight(10); !known || height != 13 {
		t.Fatalf("narrow Collapsible extent = (%d, %t), want (13, true)", height, known)
	}
}

func TestScrollableNaturalHeightFallbackDiscoversEndConservatively(t *testing.T) {
	content := &naturalHeightPanel{rows: []string{"row0", "row1", "row2", "row3"}}
	scrollable := NewScrollable(content)
	scrollable.Offset = 100

	gotElements := scrollable.Elements(8, 3)
	if want := rowElementsAt(1, 3); !reflect.DeepEqual(gotElements, want) {
		t.Fatalf("Elements = %+v, want %+v", gotElements, want)
	}
	if scrollable.Offset != 100 {
		t.Fatalf("Elements mutated Offset to %d, want 100", scrollable.Offset)
	}
	if got := content.viewRequests[len(content.viewRequests)-1]; got != 104 {
		t.Fatalf("fallback probe height = %d, want Offset + viewport + 1 = 104", got)
	}

	gotView := strings.Split(ansi.Strip(scrollable.View(8, 3)), "\n")
	if !reflect.DeepEqual(gotView, []string{"row1    ", "row2    ", "row3    "}) {
		t.Fatalf("View = %q, want rows 1-3", gotView)
	}
	if scrollable.Offset != 1 {
		t.Fatalf("View stored offset %d, want 1", scrollable.Offset)
	}
}

func TestScrollablePaddedUnknownFallbackDoesNotInventEnd(t *testing.T) {
	content := &paddedUnknownPanel{}
	scrollable := NewScrollable(content)
	scrollable.Offset = 10

	if got := scrollable.Elements(4, 3); len(got) != 0 {
		t.Fatalf("Elements at unknown offset returned %+v, want no visible rows", got)
	}
	if scrollable.Offset != 10 {
		t.Fatalf("Elements mutated Offset to %d, want 10", scrollable.Offset)
	}
	if got := content.viewRequests[len(content.viewRequests)-1]; got != 14 {
		t.Fatalf("fallback probe height = %d, want 14", got)
	}
	if got := content.elementRequests[len(content.elementRequests)-1]; got != 13 {
		t.Fatalf("Elements request height = %d, want effective offset + viewport = 13", got)
	}

	if view := scrollable.View(4, 3); len(strings.Split(view, "\n")) != 3 {
		t.Fatalf("View returned %q, want three rows", view)
	}
	if scrollable.Offset != 10 {
		t.Fatalf("padded unknown content was assigned fake end; offset=%d, want 10", scrollable.Offset)
	}
}

func TestContentHeightWrapperPropagation(t *testing.T) {
	inner := &knownExtentPanel{heightAtWidth: func(int) (int, bool) { return 7, true }}
	selectable := NewSelectable(inner)
	if height, known := selectable.ContentHeight(20); !known || height != 7 {
		t.Fatalf("Selectable extent = (%d, %t), want (7, true)", height, known)
	}

	unknown := &scrollableMockPanel{view: "unknown"}
	if height, known := NewSelectable(unknown).ContentHeight(20); known || height != 0 {
		t.Fatalf("Selectable unknown extent = (%d, %t), want (0, false)", height, known)
	}
	if height, known := NewSelectable(nil).ContentHeight(20); known || height != 0 {
		t.Fatalf("Selectable nil-content extent = (%d, %t), want (0, false)", height, known)
	}

	collapsible := NewCollapsible("section", inner)
	if height, known := collapsible.ContentHeight(20); !known || height != 8 {
		t.Fatalf("expanded Collapsible extent = (%d, %t), want (8, true)", height, known)
	}
	collapsible.Collapsed = true
	if height, known := collapsible.ContentHeight(20); !known || height != 1 {
		t.Fatalf("collapsed Collapsible extent = (%d, %t), want (1, true)", height, known)
	}
	collapsible.Collapsed = false
	collapsible.Content = unknown
	if height, known := collapsible.ContentHeight(20); known || height != 0 {
		t.Fatalf("unknown expanded Collapsible extent = (%d, %t), want (0, false)", height, known)
	}
	if height, known := NewCollapsible("empty", nil).ContentHeight(20); known || height != 0 {
		t.Fatalf("expanded nil-content Collapsible extent = (%d, %t), want (0, false)", height, known)
	}

	if height, known := NewScrollable(inner).ContentHeight(20); known || height != 0 {
		t.Fatalf("nested Scrollable extent = (%d, %t), want (0, false)", height, known)
	}
	if inner.viewCalls != 0 {
		t.Fatalf("ContentHeight methods rendered the inner panel %d times", inner.viewCalls)
	}
}

func TestScrollableNormalizesNegativeKnownContentHeight(t *testing.T) {
	content := &knownExtentPanel{
		heightAtWidth: func(int) (int, bool) { return -5, true },
		padToHeight:   true,
	}
	scrollable := NewScrollable(content)
	scrollable.Offset = 12
	scrollable.View(5, 3)
	if scrollable.Offset != 0 {
		t.Fatalf("negative known extent clamped offset to %d, want 0", scrollable.Offset)
	}
}

func TestScrollableSaturatesProbeAndContentRequests(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	finite := &naturalHeightPanel{rows: []string{"only-row"}}
	scrollable := NewScrollable(finite)
	scrollable.Offset = maxInt - 2
	scrollable.View(8, 3)
	if got := finite.viewRequests[0]; got != maxInt {
		t.Fatalf("fallback probe height = %d, want saturated MaxInt %d", got, maxInt)
	}
	if scrollable.Offset != 0 {
		t.Fatalf("finite fallback offset = %d, want 0", scrollable.Offset)
	}

	large := &largeExtentPanel{}
	scrollable = NewScrollable(large)
	scrollable.Offset = maxInt - 1
	scrollable.View(8, 3)
	wantEffective := maxInt - 3
	if scrollable.Offset != wantEffective {
		t.Fatalf("known large extent offset = %d, want %d", scrollable.Offset, wantEffective)
	}
	if large.requestedView != maxInt {
		t.Fatalf("known extent View request = %d, want MaxInt %d", large.requestedView, maxInt)
	}
	scrollable.Elements(8, 3)
	if large.requestedElements != maxInt {
		t.Fatalf("known extent Elements request = %d, want MaxInt %d", large.requestedElements, maxInt)
	}
}

func snapshotElements(t *testing.T, warp *Warp) []Element {
	t.Helper()
	recorder := httptest.NewRecorder()
	warp.handleElements(recorder, httptest.NewRequest("GET", "/elements", nil))
	var elements []Element
	if err := json.NewDecoder(recorder.Body).Decode(&elements); err != nil {
		t.Fatalf("decode /elements snapshot: %v", err)
	}
	return elements
}

func rowElements(count int) []Element {
	elements := make([]Element, count)
	for i := range elements {
		elements[i] = Element{
			Role:   "text",
			Name:   rowName(i),
			Bounds: Bounds{X: 0, Y: i, W: 4, H: 1},
		}
	}
	return elements
}

func rowElementsAt(start, count int) []Element {
	elements := rowElements(start + count)
	for i := range elements {
		elements[i].Bounds.Y -= start
	}
	return elements[start:]
}

func rowName(index int) string {
	return fmt.Sprintf("row%d", index)
}
